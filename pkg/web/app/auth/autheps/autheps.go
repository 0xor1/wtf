package autheps

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=autheps.sql

import (
	"bytes"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/crypt"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/auth"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/go-sql-driver/mysql"
)

func OnRegister(tlbx app.Tlbx, me ID, tx sql.Tx) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryOnRegister(qryArgs, me)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}

func OnActivate(tlbx app.Tlbx, me ID, tx sql.Tx) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryOnActivate(qryArgs, me)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}

func OnDelete(tlbx app.Tlbx, me ID, tx sql.Tx) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryOnDelete(qryArgs, me)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}

type Config struct {
	AppDataDefault func() interface{}
	AppDataExample func() interface{}
	OnRegister     func(tlbx app.Tlbx, me ID, appData interface{}, txAdder sql.DoTxAdder)
	OnActivate     func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder)
	OnDelete       func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder)
	OnLogout       func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder)
}

func config(configs ...func(*Config)) *Config {
	noopDoTx := func(_ app.Tlbx, _ ID, _ sql.DoTxAdder) {}
	c := &Config{
		AppDataDefault: func() interface{} { return nil },
		AppDataExample: func() interface{} { return nil },
		OnRegister:     func(_ app.Tlbx, _ ID, _ interface{}, _ sql.DoTxAdder) {},
		OnActivate:     noopDoTx,
		OnDelete:       noopDoTx,
		OnLogout:       noopDoTx,
	}
	for _, config := range configs {
		config(c)
	}
	return c
}

func New(
	fromEmail,
	activateFmtLink,
	confirmChangeEmailFmtLink string,
	configs ...func(*Config),
) []*app.Endpoint {
	c := config(configs...)
	return []*app.Endpoint{
		{
			Description:  "register a new account (requires email link)",
			Path:         (&auth.Register{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Register{
					AppData: c.AppDataDefault(),
				}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Register{
					Email:   "joe@bloggs.example",
					Pwd:     "J03-8l0-Gg5-Pwd",
					AppData: c.AppDataExample(),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				app.BadReqIf(me.AuthedExists(tlbx), "already logged in")
				args := a.(*auth.Register)
				args.Email = StrTrimWS(args.Email)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				auth := &fullAuth{}
				auth.setPwd(tlbx, args.Pwd)
				// incase they're already doing stuff as an anon user and want to
				// register to save their session state, use the anon session id
				auth.Me.ID = me.Get(tlbx).ID()
				auth.Me.Email = args.Email
				auth.ActivateCode = ptr.String(crypt.UrlSafeString(250))
				auth.RegisteredOn = Now()
				srv := service.Get(tlbx)
				authtx := srv.Auth().Begin()
				defer authtx.Rollback()
				qryArgs := sqlh.NewArgs(0)
				qry := qryInsert(qryArgs, auth)
				_, err := authtx.Exec(qry, qryArgs.Is()...)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					app.BadReqIf(ok && mySqlErr.Number == 1062, "email already registered")
					PanicOn(err)
				}
				appTx := sql.NewDoTxs()
				c.OnRegister(tlbx, auth.ID, args.AppData, appTx)
				defer appTx.Rollback()
				appTx.Do()
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, url.QueryEscape(args.Email), auth.ActivateCode))
				authtx.Commit()
				appTx.Commit()
				return nil
			},
		},
		{
			Description:  "resend activate link",
			Path:         (&auth.ResendActivateLink{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ResendActivateLink{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ResendActivateLink{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.ResendActivateLink)
				srv := service.Get(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, &args.Email, nil)
				tx.Commit()
				if auth == nil || auth.ActivateCode == nil {
					return nil
				}
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, url.QueryEscape(args.Email), *auth.ActivateCode))
				return nil
			},
		},
		{
			Description:  "activate a new account",
			Path:         (&auth.Activate{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Activate{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Activate{
					Email: "joe@bloggs.example",
					Code:  "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.Activate)
				srv := service.Get(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, &args.Email, nil)
				app.BadReqIf(*auth.ActivateCode != args.Code, "")
				auth.IsActivated = true
				auth.ActivateCode = nil
				updateAuth(tx, auth)
				appTx := sql.NewDoTxs()
				c.OnActivate(tlbx, auth.ID, appTx)
				defer appTx.Rollback()
				appTx.Do()
				appTx.Commit()
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "change email address (requires email link)",
			Path:         (&auth.ChangeEmail{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ChangeEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ChangeEmail{
					NewEmail: "new_joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.ChangeEmail)
				args.NewEmail = StrTrimWS(args.NewEmail)
				validate.Str("email", args.NewEmail, tlbx, 0, emailMaxLen, emailRegex)
				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				changeEmailCode := crypt.UrlSafeString(250)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				existingAuth := getAuth(tx, &args.NewEmail, nil)
				app.BadReqIf(existingAuth != nil, "email already registered")
				auth := getAuth(tx, nil, &me)
				auth.NewEmail = &args.NewEmail
				auth.ChangeEmailCode = &changeEmailCode
				updateAuth(tx, auth)
				tx.Commit()
				sendConfirmChangeEmailEmail(srv, args.NewEmail, fromEmail, Strf(confirmChangeEmailFmtLink, me, changeEmailCode))
				return nil
			},
		},
		{
			Description:  "resend change email link",
			Path:         (&auth.ResendChangeEmailLink{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, nil, &me)
				tx.Commit()
				sendConfirmChangeEmailEmail(srv, *auth.NewEmail, fromEmail, Strf(confirmChangeEmailFmtLink, me, *auth.ChangeEmailCode))
				return nil
			},
		},
		{
			Description:  "confirm change email",
			Path:         (&auth.ConfirmChangeEmail{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ConfirmChangeEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ConfirmChangeEmail{
					Me:   app.ExampleID(),
					Code: "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.ConfirmChangeEmail)
				srv := service.Get(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, nil, &args.Me)
				app.BadReqIf(*auth.ChangeEmailCode != args.Code, "")
				auth.ChangeEmailCode = nil
				auth.Email = *auth.NewEmail
				auth.NewEmail = nil
				updateAuth(tx, auth)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "reset password (requires email link)",
			Path:         (&auth.ResetPwd{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ResetPwd{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ResetPwd{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.ResetPwd)
				srv := service.Get(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, &args.Email, nil)
				if auth != nil {
					now := Now()
					if auth.LastPwdResetOn != nil {
						mustWaitDur := (10 * time.Minute) - Now().Sub(*auth.LastPwdResetOn)
						app.BadReqIf(mustWaitDur > 0, "must wait %d seconds before reseting pwd again", int64(math.Ceil(mustWaitDur.Seconds())))
					}
					auth.LastPwdResetOn = &now
					newPwd := `$aA1` + crypt.UrlSafeString(12)
					auth.setPwd(tlbx, newPwd)
					updateAuth(tx, auth)
					sendResetPwdEmail(srv, args.Email, fromEmail, newPwd)
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "change password",
			Path:         (&auth.ChangePwd{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ChangePwd{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ChangePwd{
					OldPwd: "J03-8l0-Gg5-Pwd",
					NewPwd: "N3w-J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.ChangePwd)
				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, nil, &me)
				app.BadReqIf(!bytes.Equal(crypt.ScryptKey([]byte(args.OldPwd), auth.Salt, auth.N, auth.R, auth.P, scryptKeyLen), auth.Pwd), "old pwd does not match")
				auth.setPwd(tlbx, args.NewPwd)
				updateAuth(tx, auth)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "delete account",
			Path:         (&auth.Delete{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Delete{
					Pwd: "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.Delete)
				srv := service.Get(tlbx)
				m := me.AuthedGet(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, nil, &m)
				app.BadReqIf(!bytes.Equal(auth.Pwd, crypt.ScryptKey([]byte(args.Pwd), auth.Salt, auth.N, auth.R, auth.P, scryptKeyLen)), "incorrect pwd")
				qryArgs := sqlh.NewArgs(0)
				qry := qryDelete(qryArgs, m)
				_, err := tx.Exec(qry, qryArgs.Is()...)
				PanicOn(err)
				appTx := sql.NewDoTxs()
				c.OnDelete(tlbx, m, appTx)
				defer appTx.Rollback()
				appTx.Do()
				appTx.Commit()
				me.Del(tlbx)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "login",
			Path:         (&auth.Login{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Login{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Login{
					Email: "joe@bloggs.example",
					Pwd:   "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return app.ExampleID()
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*auth.Login)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				validate.Str("pwd", args.Pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
				srv := service.Get(tlbx)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, &args.Email, nil)
				app.ReturnIf(auth == nil || !bytes.Equal(auth.Pwd, crypt.ScryptKey([]byte(args.Pwd), auth.Salt, auth.N, auth.R, auth.P, scryptKeyLen)), http.StatusNotFound, "email and/or pwd are not valid")
				// if encryption params have changed re encrypt on successful login
				if len(auth.Salt) != scryptSaltLen ||
					len(auth.Pwd) != scryptKeyLen ||
					auth.N != scryptN ||
					auth.R != scryptR ||
					auth.P != scryptP {
					auth.setPwd(tlbx, args.Pwd)
				}
				updateAuth(tx, auth)
				tx.Commit()
				me.AuthedSet(tlbx, auth.ID)
				return &auth.ID
			},
		},
		{
			Description:  "logout",
			Path:         (&auth.Logout{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				if me.AuthedExists(tlbx) {
					m := me.AuthedGet(tlbx)
					appTx := sql.NewDoTxs()
					c.OnLogout(tlbx, m, appTx)
					defer appTx.Rollback()
					appTx.Do()
					appTx.Commit()
				}
				me.Del(tlbx)
				return nil
			},
		},
		{
			Description:  "get me",
			Path:         (&auth.GetMe{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return &auth.Me{
					ID:    app.ExampleID(),
					Email: "joe@bloggs.example",
				}
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				if !me.AuthedExists(tlbx) {
					return nil
				}
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).Auth().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, nil, &me)
				tx.Commit()
				return &auth.Me
			},
		},
	}
}

var (
	emailRegex  = regexp.MustCompile(`\A.+@.+\..+\z`)
	emailMaxLen = 250
	pwdRegexs   = []*regexp.Regexp{
		regexp.MustCompile(`[0-9]`),
		regexp.MustCompile(`[a-z]`),
		regexp.MustCompile(`[A-Z]`),
		regexp.MustCompile(`[\w]`),
	}
	pwdMinLen     = 8
	pwdMaxLen     = 100
	scryptN       = 32768
	scryptR       = 8
	scryptP       = 1
	scryptSaltLen = 256
	scryptKeyLen  = 256
)

func sendActivateEmail(srv service.Layer, sendTo, from, link string) {
	html := `<p>Thank you for registering.</p><p>Click this link to activate your account:</p><p><a href="` + link + `">Activate</a></p><p>If you didn't register for this account you can simply ignore this email.</p>`
	txt := "Thank you for registering.\nClick this link to activate your account:\n\n" + link + "\n\nIf you didn't register for this account you can simply ignore this email."
	srv.Email().MustSend([]string{sendTo}, from, "Activate", html, txt)
}

func sendConfirmChangeEmailEmail(srv service.Layer, sendTo, from, link string) {
	srv.Email().MustSend([]string{sendTo}, from, "Confirm change email",
		`<p>Click this link to change the email associated with your account:</p><p><a href="`+link+`">Confirm change email</a></p>`,
		"Confirm change email:\n\n"+link)
}

func sendResetPwdEmail(srv service.Layer, sendTo, from, newPwd string) {
	srv.Email().MustSend([]string{sendTo}, from, "Pwd Reset", `<p>New Pwd: `+newPwd+`</p>`, `New Pwd: `+newPwd)
}

type fullAuth struct {
	auth.Me
	IsActivated     bool
	RegisteredOn    time.Time
	LastLoggedInOn  time.Time
	NewEmail        *string
	ActivateCode    *string
	ChangeEmailCode *string
	LastPwdResetOn  *time.Time
	Salt            []byte
	Pwd             []byte
	N               int
	R               int
	P               int
}

func getAuth(tx sql.Tx, email *string, id *ID) *fullAuth {
	PanicIf(email == nil && id == nil, "one of email or id must not be nil")
	qryArgs := sqlh.NewArgs(0)
	qry := qrySelect(qryArgs, email, id)
	row := tx.QueryRow(qry, qryArgs.Is()...)
	res := &fullAuth{}
	err := row.Scan(
		&res.ID,
		&res.Email,
		&res.IsActivated,
		&res.RegisteredOn,
		&res.NewEmail,
		&res.ActivateCode,
		&res.ChangeEmailCode,
		&res.LastPwdResetOn,
		&res.Salt,
		&res.Pwd,
		&res.N,
		&res.R,
		&res.P)
	if err == isql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateAuth(tx sql.Tx, auth *fullAuth) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryUpdate(qryArgs, auth)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}

func (a *fullAuth) setPwd(tlbx app.Tlbx, pwd string) {
	validate.Str("pwd", pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
	a.Salt = crypt.Bytes(scryptSaltLen)
	a.N = scryptN
	a.R = scryptR
	a.P = scryptP
	a.Pwd = crypt.ScryptKey([]byte(pwd), a.Salt, a.N, a.R, a.P, scryptKeyLen)
}
