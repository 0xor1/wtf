package authtest

import (
	"regexp"
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/auth"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	r := test.NewRig(
		config.GetProcessed(config.GetBase()),
		nil,
		ratelimit.MeMware,
		nil,
		true,
		nil)
	defer r.CleanUp()

	a := assert.New(t)
	c := r.NewClient()
	email := "test@test.localhost%s" + r.Unique()
	pwd := "1aA$_t;3"

	(&auth.Register{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c)

	// check existing email err
	err := (&auth.Register{
		Email: email,
		Pwd:   pwd,
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: "email already registered"}, err)

	(&auth.ResendActivateLink{
		Email: email,
	}).MustDo(c)

	var code string
	row := r.Auth().Primary().QueryRow(`SELECT activateCode FROM auths WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&auth.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	// check return ealry path
	(&auth.ResendActivateLink{
		Email: email,
	}).MustDo(c)

	id := *(&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c)

	tmpFirstID := id.Copy()
	defer func() {
		_, err = r.Auth().Primary().Exec(`DELETE FROM auths WHERE id=?`, tmpFirstID)
		PanicOn(err)
	}()

	(&auth.ChangeEmail{
		NewEmail: Strf("change@test.localhost%s", r.Unique()),
	}).MustDo(c)

	(&auth.ResendChangeEmailLink{}).MustDo(c)

	row = r.Auth().Primary().QueryRow(`SELECT changeEmailCode FROM auths WHERE id=?`, id)
	PanicOn(row.Scan(&code))

	(&auth.ConfirmChangeEmail{
		Me:   id,
		Code: code,
	}).MustDo(c)

	(&auth.ChangeEmail{
		NewEmail: email,
	}).MustDo(c)

	row = r.Auth().Primary().QueryRow(`SELECT changeEmailCode FROM auths WHERE id=?`, id)
	PanicOn(row.Scan(&code))

	(&auth.ConfirmChangeEmail{
		Me:   id,
		Code: code,
	}).MustDo(c)

	newPwd := pwd + "123abc"
	(&auth.ChangePwd{
		OldPwd: pwd,
		NewPwd: newPwd,
	}).MustDo(c)

	(&auth.Logout{}).MustDo(c)

	(&auth.Login{
		Email: email,
		Pwd:   newPwd,
	}).MustDo(c)

	me := (&auth.GetMe{}).MustDo(c)
	a.True(id.Equal(me.ID))
	a.Equal(email, me.Email)

	(&auth.Delete{
		Pwd: newPwd,
	}).MustDo(c)

	(&auth.Register{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c)

	row = r.Auth().Primary().QueryRow(`SELECT activateCode FROM auths WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&auth.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	id = *(&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c)
	a.Equal(id, (&auth.GetMe{}).MustDo(c).ID)

	defer func() {
		_, err = r.Auth().Primary().Exec(`DELETE FROM auths WHERE id=?`, id)
		PanicOn(err)
	}()

	(&auth.ResetPwd{
		Email: email,
	}).MustDo(c)

	err = (&auth.ResetPwd{
		Email: email,
	}).Do(c)
	a.Equal(400, err.(*app.ErrMsg).Status)
	a.True(regexp.MustCompile(`must wait [1-9][0-9]{2} seconds before reseting pwd again`).MatchString(err.(*app.ErrMsg).Msg))
}
