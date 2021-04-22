package social

import (
	"io"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type RegisterAppData struct {
	Handle string `json:"handle"`
	Alias  string `json:"alias"`
}

type Update struct {
	Handle *field.String `json:"handle,omitempty"`
	Alias  *field.String `json:"alias,omitempty"`
}

func (_ *Update) Path() string {
	return "/user/social/update"
}

func (a *Update) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Update) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type SetAvatar struct {
	Avatar io.ReadCloser
}

func (_ *SetAvatar) Path() string {
	return "/user/social/setAvatar"
}

func (a *SetAvatar) Do(c *app.Client) error {
	var stream *app.UpStream
	if a.Avatar != nil {
		stream = &app.UpStream{}
		stream.Content = a.Avatar
	}
	return app.Call(c, a.Path(), stream, nil)
}

func (a *SetAvatar) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type GetAvatar struct {
	ID ID `json:"id"`
}

func (_ *GetAvatar) Path() string {
	return "/user/social/getAvatar"
}

func (a *GetAvatar) Do(c *app.Client) (*app.DownStream, error) {
	res := &app.DownStream{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetAvatar) MustDo(c *app.Client) *app.DownStream {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetRes struct {
	Set  []*Social `json:"set"`
	More bool      `json:"more"`
}

type Get struct {
	IDs          IDs    `json:"ids,omitempty"`
	HandlePrefix string `json:"handlePrefix"`
	Limit        uint16 `json:"limit"`
}

func (_ *Get) Path() string {
	return "/user/social/get"
}

func (a *Get) Do(c *app.Client) (*GetRes, error) {
	res := &GetRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *GetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetMe struct{}

func (_ *GetMe) Path() string {
	return "/user/social/getMe"
}

func (a *GetMe) Do(c *app.Client) (*Social, error) {
	res := &Social{}
	err := app.Call(c, a.Path(), nil, &res)
	return res, err
}

func (a *GetMe) MustDo(c *app.Client) *Social {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Social struct {
	ID        ID     `json:"id"`
	Handle    string `json:"handle"`
	Alias     string `json:"alias"`
	HasAvatar bool   `json:"hasAvatar"`
}
