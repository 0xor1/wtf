package fcm

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type SetEnabled struct {
	Val bool `json:"val"`
}

func (_ *SetEnabled) Path() string {
	return "/user/fcm/setEnabled"
}

func (a *SetEnabled) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *SetEnabled) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Register struct {
	Topic  IDs    `json:"topic"`
	Client *ID    `json:"client"`
	Token  string `json:"token"`
}

func (_ *Register) Path() string {
	return "/user/fcm/register"
}

func (a *Register) Do(c *app.Client) (*ID, error) {
	res := &ID{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Register) MustDo(c *app.Client) *ID {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Unregister struct {
	Client ID `json:"client"`
}

func (_ *Unregister) Path() string {
	return "/user/fcm/unregister"
}

func (a *Unregister) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Unregister) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
