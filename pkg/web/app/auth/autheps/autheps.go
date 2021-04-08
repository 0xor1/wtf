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
	"github.com/0xor1/tlbx/pkg/web/app/session/opt"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/go-sql-driver/mysql"
)

func NewMe(
	fromEmail,
	activateFmtLink,
	confirmChangeEmailFmtLink string,
	onDelete func(app.Tlbx, ID),
) []*app.Endpoint {
	return New(
		fromEmail,
		activateFmtLink,
		confirmChangeEmailFmtLink,
		me.Exists,
		me.Set,
		me.Get,
		me.Del,
		onDelete)
}

func NewOpt(
	fromEmail,
	activateFmtLink,
	confirmChangeEmailFmtLink string,
	onDelete func(app.Tlbx, ID),
) []*app.Endpoint {
	return New(
		fromEmail,
		activateFmtLink,
		confirmChangeEmailFmtLink,
		opt.AuthedExists,
		opt.AuthedSet,
		opt.AuthedGet,
		opt.Del,
		onDelete)
}

func New(
	fromEmail,
	activateFmtLink,
	confirmChangeEmailFmtLink string,
	sessionExists func(app.Tlbx) bool,
	sessionSet func(app.Tlbx, ID),
	sessionGet func(app.Tlbx) ID,
	sessionDel func(app.Tlbx),
	onDelete func(app.Tlbx, ID),
) []*app.Endpoint {
	return []*app.Endpoint{
		{
			Description:  "register a new account (requires email link)",
			Path:         (&auth.Register{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Register{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Register{
					Email: "joe@bloggs.example",
					Pwd:   "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				app.BadReqIf(sessionExists(tlbx), "already logged in")
				args := a.(*auth.Register)
				args.Email = StrTrimWS(args.Email)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				auth := &fullAuth{}
				auth.Me.ID = tlbx.NewID()
				auth.Me.Email = args.Email
				auth.ActivateCode = ptr.String(crypt.UrlSafeString(250))
				auth.RegisteredOn = Now()
				srv := service.Get(tlbx)

				tx := srv.Auth().Begin()
				defer usrtx.Rollback()
				var qryArgs *sqlh.Args
				qry := qryInsert(qryArgs, auth)
				_, err := tx.Exec(qry, qryArgs.Is()...)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					app.BadReqIf(ok && mySqlErr.Number == 1062, "email already registered")
					PanicOn(err)
				}
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, url.QueryEscape(args.Email), activateCode))
				tx.Commit()
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
				fullUser := getAuth(tx, &args.Email, nil)
				tx.Commit()
				if fullUser == nil || fullUser.ActivateCode == nil {
					return nil
				}
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, url.QueryEscape(args.Email), *fullUser.ActivateCode))
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
				auth.ActivatedOn = Now()
				auth.ActivateCode = nil
				updateAuth(tx, auth)
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
				me := sessionGet(tlbx)
				changeEmailCode := crypt.UrlSafeString(250)
				tx := srv.Auth().Begin()
				defer tx.Rollback()
				existingUser := getAuth(tx, &args.NewEmail, nil)
				app.BadReqIf(existingUser != nil, "email already registered")
				fullUser := getAuth(tx, nil, &me)
				fullUser.NewEmail = &args.NewEmail
				fullUser.ChangeEmailCode = &changeEmailCode
				updateAuth(tx, fullUser)
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
				me := sessionGet(tlbx)
				tx := srv.User().Begin()
				defer tx.Rollback()
				fullUser := getAuth(tx, nil, &me)
				tx.Commit()
				sendConfirmChangeEmailEmail(srv, *fullUser.NewEmail, fromEmail, Strf(confirmChangeEmailFmtLink, me, *fullUser.ChangeEmailCode))
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
				tx := srv.User().Begin()
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
				tx := srv.User().Begin()
				defer tx.Rollback()
				auth := getAuth(tx, &args.Email, nil)
				if auth != nil {
					now := Now()
					if auth.LastPwdResetOn != nil {
						mustWaitDur := (10 * time.Minute) - Now().Sub(*auth.LastPwdResetOn)
						app.BadReqIf(mustWaitDur > 0, "must wait %d seconds before reseting pwd again", int64(math.Ceil(mustWaitDur.Seconds())))
					}
					auth.LastPwdResetOn = &now
					updateAuth(tx, auth)
					pwdtx := srv.Pwd().Begin()
					defer pwdtx.Rollback()
					newPwd := `$aA1` + crypt.UrlSafeString(12)
					setPwd(tlbx, pwdtx, user.ID, newPwd, newPwd)
					sendResetPwdEmail(srv, args.Email, fromEmail, newPwd)
					pwdtx.Commit()
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
				args := a.(*auth.SetPwd)
				srv := service.Get(tlbx)
				me := sessionGet(tlbx)
				pwd := getPwd(srv, me)
				app.BadReqIf(!bytes.Equal(crypt.ScryptKey([]byte(args.OldPwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen), pwd.Pwd), "old pwd does not match")
				pwdtx := srv.Pwd().Begin()
				defer pwdtx.Rollback()
				setPwd(tlbx, pwdtx, me, args.NewPwd, args.ConfirmNewPwd)
				pwdtx.Commit()
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
				m := sessionGet(tlbx)
				pwd := getPwd(srv, m)
				app.BadReqIf(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)), "incorrect pwd")
				tx := srv.User().Begin()
				defer tx.Rollback()
				_, err := tx.Exec(`DELETE FROM auths WHERE id=?`, m)
				if onDelete != nil {
					onDelete(tlbx, m)
				}
				sessionDel(tlbx)
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
				return &auth.User{
					ID: app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				emailOrPwdMismatch := func(condition bool) {
					app.ReturnIf(condition, http.StatusNotFound, "email and/or pwd are not valid")
				}
				args := a.(*auth.Login)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				validate.Str("pwd", args.Pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
				srv := service.Get(tlbx)
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getAuth(tx, &args.Email, nil)
				emailOrPwdMismatch(user == nil)
				pwd := getPwd(srv, user.ID)
				emailOrPwdMismatch(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)))
				// if encryption params have changed re encrypt on successful login
				if len(pwd.Salt) != scryptSaltLen || len(pwd.Pwd) != scryptKeyLen || pwd.N != scryptN || pwd.R != scryptR || pwd.P != scryptP {
					pwdtx := srv.Pwd().Begin()
					defer pwdtx.Rollback()
					setPwd(tlbx, pwdtx, user.ID, args.Pwd, args.Pwd)
					pwdtx.Commit()
				}
				tx.Commit()
				sessionSet(tlbx, user.ID)
				return &auth.User
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
				if sessionExists(tlbx) {
					m := sessionGet(tlbx)
					srv := service.Get(tlbx)
					tokens := make([]string, 0, 5)
					tx := srv.User().Begin()
					defer tx.Rollback()
					tx.Query(func(rows isql.Rows) {
						for rows.Next() {
							token := ""
							PanicOn(rows.Scan(&token))
							tokens = append(tokens, token)
						}
					}, `SELECT DISTINCT token FROM fcmTokens WHERE user=?`, m)
					_, err := tx.Exec(`DELETE FROM fcmTokens WHERE user=?`, m)
					PanicOn(err)
					srv.FCM().RawAsyncSend("logout", tokens, map[string]string{}, 0)
					tx.Commit()
					sessionDel(tlbx)
				}
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
				if !sessionExists(tlbx) {
					return nil
				}
				me := sessionGet(tlbx)
				tx := service.Get(tlbx).Auth().Begin()
				defer tx.Rollback()
				user := getAuth(tx, nil, &me)
				tx.Commit()
				return &auth.User
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

func sendActivateEmail(srv service.Layer, sendTo, from, link string, handle *string) {
	html := `<p>Thank you for registering.</p><p>Click this link to activate your account:</p><p><a href="` + link + `">Activate</a></p><p>If you didn't register for this account you can simply ignore this email.</p>`
	txt := "Thank you for registering.\nClick this link to activate your account:\n\n" + link + "\n\nIf you didn't register for this account you can simply ignore this email."
	if handle != nil {
		html = Strf("Hi %s,\n\n", *handle) + html
		txt = Strf("Hi %s,\n\n", *handle) + txt
	}
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
	RegisteredOn    time.Time
	ActivatedOn     time.Time
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
	var args *sqlh.Args
	qry := qryGet(args, email, id)
	row := tx.QueryRow(qry, args.Is()...)
	res := &fullAuth{}
	err := row.Scan(
		&res.ID,
		&res.Email,
		&res.RegisteredOn,
		&res.ActivatedOn,
		&res.NewEmail,
		&res.ActivateCode,
		&res.ChangeEmailCode,
		&res.LastPwdResetOn)
	if err == isql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateAuth(tx sql.Tx, auth *fullAuth) {
	var args *sqlh.Args
	qry := qryUpdate(args, auth)
	_, err := tx.Exec(qry, args.Is()...)
	PanicOn(err)
}
