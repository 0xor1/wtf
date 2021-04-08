package auth

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Register struct {
	Email   string      `json:"email"`
	Pwd     string      `json:"pwd"`
	AppData interface{} `json:"appData,omitempty"`
}

func (_ *Register) Path() string {
	return "/auth/register"
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
	return "/auth/resendActivateLink"
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
	return "/auth/activate"
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
	return "/auth/changeEmail"
}

func (a *ChangeEmail) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ChangeEmail) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResendChangeEmailLink struct{}

func (_ *ResendChangeEmailLink) Path() string {
	return "/auth/resendChangeEmailLink"
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
	return "/auth/confirmChangeEmail"
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
	return "/auth/resetPwd"
}

func (a *ResetPwd) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ResetPwd) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ChangePwd struct {
	OldPwd string `json:"oldPwd"`
	NewPwd string `json:"newpwd"`
}

func (_ *ChangePwd) Path() string {
	return "/auth/changePwd"
}

func (a *ChangePwd) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ChangePwd) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Delete struct {
	Pwd string `json:"pwd"`
}

func (_ *Delete) Path() string {
	return "/auth/delete"
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
	return "/auth/login"
}

func (a *Login) Do(c *app.Client) (*ID, error) {
	res := &ID{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Login) MustDo(c *app.Client) *ID {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Logout struct{}

func (_ *Logout) Path() string {
	return "/auth/logout"
}

func (a *Logout) Do(c *app.Client) error {
	return app.Call(c, a.Path(), nil, nil)
}

func (a *Logout) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Me struct {
	ID    ID     `json:"id"`
	Email string `json:"email"`
}

type GetMe struct{}

func (_ *GetMe) Path() string {
	return "/auth/getMe"
}

func (a *GetMe) Do(c *app.Client) (*Me, error) {
	res := &Me{}
	err := app.Call(c, a.Path(), nil, &res)
	return res, err
}

func (a *GetMe) MustDo(c *app.Client) *Me {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}
