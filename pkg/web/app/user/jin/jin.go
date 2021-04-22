package jin

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Set struct {
	Val *json.Json `json:"val"`
}

func (_ *Set) Path() string {
	return "/user/jin/set"
}

func (a *Set) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Set) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Get struct{}

func (_ *Get) Path() string {
	return "/user/jin/get"
}

func (a *Get) Do(c *app.Client) (*json.Json, error) {
	res := &json.Json{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *json.Json {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}
