package ratelimit

import (
	"net/http"
	"strconv"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/server/realip"
	"github.com/gomodule/redigo/redis"
)

func NoMware(cache iredis.Pool, perMinute ...int) func(app.Tlbx) {
	return BasicMware(func(t app.Tlbx) bool { return false }, nil, cache, perMinute...)
}

func MeMware(cache iredis.Pool, perMinute ...int) func(app.Tlbx) {
	return BasicMware(me.AuthedExists, func(t app.Tlbx) string {
		return me.AuthedGet(t).String()
	}, cache, perMinute...)
}

func BasicMware(sesExists func(app.Tlbx) bool, sesID func(app.Tlbx) string, cache iredis.Pool, perMinute ...int) func(app.Tlbx) {
	PanicIf(len(perMinute) != 0 && perMinute[0] < 1, "perMinute must be >= 1")
	return Mware(func(c *Config) {
		c.KeyGen = func(tlbx app.Tlbx) string {
			var key string
			if sesExists(tlbx) {
				key = sesID(tlbx)
			}
			return Strf("rate-limiter-%s-%s", realip.RealIP(tlbx.Req()), key)
		}
		c.Pool = cache
		if len(perMinute) != 0 {
			c.PerMinute = perMinute[0]
		}
	})
}

func Mware(configs ...func(*Config)) func(app.Tlbx) {
	c := config(configs...)
	return func(tlbx app.Tlbx) {
		if c.Pool == nil ||
			c.PerMinute < 1 ||
			c.KeyGen == nil {
			return
		}

		shouldReturn := func(err error) bool {
			if err != nil {
				if c.ExitOnError {
					PanicOn(err)
				}
				tlbx.Log().ErrorOn(err)
				return true
			}
			return false
		}

		remaining := c.PerMinute

		defer func() {
			tlbx.Resp().Header().Add("X-Rate-Limit-Limit", strconv.Itoa(c.PerMinute))
			tlbx.Resp().Header().Add("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			tlbx.Resp().Header().Add("X-Rate-Limit-Reset", "60")

			app.ReturnIf(remaining < 1, http.StatusTooManyRequests, "")
		}()

		key := c.KeyGen(tlbx)

		now := NowUnixNano()
		cnn := c.Pool.Get()
		defer cnn.Close()

		err := cnn.Send("MULTI")
		if shouldReturn(err) {
			return
		}

		err = cnn.Send("ZREMRANGEBYSCORE", key, 0, now-time.Minute.Nanoseconds())
		if shouldReturn(err) {
			return
		}

		err = cnn.Send("ZRANGE", key, 0, -1)
		if shouldReturn(err) {
			return
		}

		results, err := redis.Values(cnn.Do("EXEC"))
		if shouldReturn(err) {
			return
		}

		keys, err := redis.Strings(results[len(results)-1], err)
		if shouldReturn(err) {
			return
		}

		remaining = remaining - len(keys)

		if remaining > 0 {
			remaining--

			err := cnn.Send("MULTI")
			if shouldReturn(err) {
				return
			}

			err = cnn.Send("ZADD", key, now, now)
			if shouldReturn(err) {
				return
			}

			err = cnn.Send("EXPIRE", key, 60)
			if shouldReturn(err) {
				return
			}

			_, err = cnn.Do("EXEC")
			if shouldReturn(err) {
				return
			}
		}
	}
}

type Config struct {
	KeyGen      func(tlbx app.Tlbx) string
	PerMinute   int
	ExitOnError bool
	Pool        iredis.Pool
}

func config(configs ...func(*Config)) *Config {
	c := &Config{
		KeyGen:      nil,
		PerMinute:   300,
		ExitOnError: false,
		Pool:        nil,
	}
	for _, config := range configs {
		config(c)
	}
	return c
}
