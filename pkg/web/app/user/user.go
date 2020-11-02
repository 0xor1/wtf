package user

import (
	"io"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Register struct {
	Alias      *string `json:"alias,omitempty"`
	Handle     *string `json:"handle,omitempty"`
	Email      string  `json:"email"`
	Pwd        string  `json:"pwd"`
	ConfirmPwd string  `json:"confirmPwd"`
}

func (_ *Register) Path() string {
	return "/user/register"
}

func (a *Register) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Register) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResendActivateLink struct {
	Email string `json:"email"`
}

func (_ *ResendActivateLink) Path() string {
	return "/user/resendActivateLink"
}

func (a *ResendActivateLink) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ResendActivateLink) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Activate struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (_ *Activate) Path() string {
	return "/user/activate"
}

func (a *Activate) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Activate) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ChangeEmail struct {
	NewEmail string `json:"newEmail"`
}

func (_ *ChangeEmail) Path() string {
	return "/user/changeEmail"
}

func (a *ChangeEmail) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ChangeEmail) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResendChangeEmailLink struct{}

func (_ *ResendChangeEmailLink) Path() string {
	return "/user/resendChangeEmailLink"
}

func (a *ResendChangeEmailLink) Do(c *app.Client) error {
	return app.Call(c, a.Path(), nil, nil)
}

func (a *ResendChangeEmailLink) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ConfirmChangeEmail struct {
	Me   ID     `json:"me"`
	Code string `json:"code"`
}

func (_ *ConfirmChangeEmail) Path() string {
	return "/user/confirmChangeEmail"
}

func (a *ConfirmChangeEmail) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ConfirmChangeEmail) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResetPwd struct {
	Email string `json:"email"`
}

func (_ *ResetPwd) Path() string {
	return "/user/resetPwd"
}

func (a *ResetPwd) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ResetPwd) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type SetHandle struct {
	Handle string `json:"handle"`
}

func (_ *SetHandle) Path() string {
	return "/user/setHandle"
}

func (a *SetHandle) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *SetHandle) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type SetAlias struct {
	Alias *string `json:"alias"`
}

func (_ *SetAlias) Path() string {
	return "/user/setAlias"
}

func (a *SetAlias) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *SetAlias) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type SetAvatar struct {
	Type   string
	Avatar io.ReadCloser
}

func (_ *SetAvatar) Path() string {
	return "/user/setAvatar"
}

func (a *SetAvatar) Do(c *app.Client) error {
	var stream *app.UpStream
	if a.Avatar != nil {
		stream = &app.UpStream{}
		stream.Type = a.Type
		stream.Content = a.Avatar
	}
	return app.Call(c, a.Path(), stream, nil)
}

func (a *SetAvatar) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type SetPwd struct {
	CurrentPwd    string `json:"currentPwd"`
	NewPwd        string `json:"newPwd"`
	ConfirmNewPwd string `json:"confirmNewPwd"`
}

func (_ *SetPwd) Path() string {
	return "/user/setPwd"
}

func (a *SetPwd) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *SetPwd) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Delete struct {
	Pwd string `json:"pwd"`
}

func (_ *Delete) Path() string {
	return "/user/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Login struct {
	Email string `json:"email"`
	Pwd   string `json:"pwd"`
}

func (_ *Login) Path() string {
	return "/user/login"
}

func (a *Login) Do(c *app.Client) (*User, error) {
	res := &User{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Login) MustDo(c *app.Client) *User {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Logout struct{}

func (_ *Logout) Path() string {
	return "/user/logout"
}

func (a *Logout) Do(c *app.Client) error {
	return app.Call(c, a.Path(), nil, nil)
}

func (a *Logout) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type GetMe struct{}

func (_ *GetMe) Path() string {
	return "/user/me"
}

func (a *GetMe) Do(c *app.Client) (*User, error) {
	res := &User{}
	err := app.Call(c, a.Path(), nil, &res)
	return res, err
}

func (a *GetMe) MustDo(c *app.Client) *User {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Users []ID `json:"users"`
}

func (_ *Get) Path() string {
	return "/user/get"
}

func (a *Get) Do(c *app.Client) ([]*User, error) {
	res := []*User{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) []*User {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetAvatar struct {
	User ID `json:"user"`
}

func (_ *GetAvatar) Path() string {
	return "/user/getAvatar"
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

type User struct {
	ID        ID      `json:"id"`
	Handle    *string `json:"handle,omitempty"`
	Alias     *string `json:"alias,omitempty"`
	HasAvatar *bool   `json:"hasAvatar,omitempty"`
}
