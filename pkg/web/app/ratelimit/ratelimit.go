package ratelimit

import (
	"net/http"
	"strconv"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/gomodule/redigo/redis"
	"github.com/tomasen/realip"
)

func MeMware(cache iredis.Pool) func(app.Tlbx) {
	return Mware(func(c *Config) {
		c.KeyGen = func(tlbx app.Tlbx) string {
			var key string
			if me.Exists(tlbx) {
				key = me.Get(tlbx).String()
			}
			return Sprintf("rate-limiter-%s-%s", realip.RealIP(tlbx.Req()), key)
		}
		c.Pool = cache
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

			tlbx.ExitIf(remaining < 1, http.StatusTooManyRequests, "")
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
		PerMinute:   120,
		ExitOnError: false,
		Pool:        nil,
	}
	for _, config := range configs {
		config(c)
	}
	return c
}