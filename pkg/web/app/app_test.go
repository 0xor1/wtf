package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user/social/socialtest"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	// use socialeps here because it's a common bundled
	// set of endpoints that tests many of app.go
	// functionality
	socialtest.Everything(t)

	// Now test all the functionality that autheps tests
	// doesnt use, i.e. mdo/upstreams/downstreams/redirects

	r := test.NewRig(
		config.GetProcessed(config.GetBase()),
		[]*app.Endpoint{
			{
				Description:  "echo back the json obj args",
				Path:         "/test/echo",
				Timeout:      500,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return &map[string]interface{}{}
				},
				GetExampleArgs: func() interface{} {
					return &map[string]interface{}{
						"a": "ali",
						"b": "bob",
						"c": "cat",
					}
				},
				GetExampleResponse: func() interface{} {
					return &map[string]interface{}{
						"a": "ali",
						"b": "bob",
						"c": "cat",
					}
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					return args
				},
			},
			{
				Description:  "redirect",
				Path:         "/test/redirect",
				Timeout:      500,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					app.Redirect(http.StatusMovedPermanently, "https://github.com/0xor1/tlbx")
					return nil
				},
			},
			{
				Description:  "timeout",
				Path:         "/test/timeout",
				Timeout:      100,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					time.Sleep(150 * time.Millisecond)
					return nil
				},
			},
			{
				Description:  "panic",
				Path:         "/test/panic",
				Timeout:      100,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					PanicOn("yolo")
					return nil
				},
			},
		},
		ratelimit.MeMware,
		nil,
		false,
		nil)
	defer r.CleanUp()

	a := assert.New(t)
	c := r.NewClient()
	a.Equal("pong", (&app.Ping{}).MustDo(c))

	// test mdo and basic reponse statuses
	mdoRes := (&app.MDo{
		"0": {
			Path: "/api/ping",
		},
		"1": {
			Path: "/api/test/echo",
			Args: json.FromInterface(map[string]interface{}{
				"msg": "yolo",
			}),
		},
		"2": {
			Path: "/api/test/redirect",
		},
		"3": {
			Path: "/api/test/timeout",
		},
		"4": {
			Path: "/api/test/panic",
		},
	}).MustDo(c)
	a.Equal(http.StatusOK, mdoRes["0"].Status)
	a.Equal(http.StatusOK, mdoRes["1"].Status)
	a.Equal("yolo", mdoRes["1"].Body.MustString("msg"))
	a.Equal(http.StatusMovedPermanently, mdoRes["2"].Status)
	a.Equal(http.StatusServiceUnavailable, mdoRes["3"].Status)
	a.Equal(http.StatusInternalServerError, mdoRes["4"].Status)

	// test static file headers
	req, err := http.NewRequest(http.MethodGet, "/notfound", nil)
	req.Header.Add("X-Client", "tlbx-app-tests")
	PanicOn(err)
	w := httptest.NewRecorder()
	r.RootHandler().ServeHTTP(w, req)
	a.Equal(http.StatusNotFound, w.Result().StatusCode)
	a.Equal(w.Header().Get("Cache-Control"), "public, max-age=3600, immutable")
	a.Equal(w.Header().Get("X-Frame-Options"), "DENY")
	a.Equal(w.Header().Get("X-XSS-Protection"), "1; mode=block")
	a.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}
