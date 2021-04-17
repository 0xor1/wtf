package socialeps

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=socialeps.sql

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/crypt"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/social"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/disintegration/imaging"
)

const (
	AvatarBucket = "avatars"
	AvatarPrefix = ""
)

func NopOnSetSocials(_ app.Tlbx, _ *social.Socials, txAdder sql.DoTxAdder) error {
	return nil
}

func OnRegister(tlbx app.Tlbx, me ID, appData social.RegisterAppData, tx sql.Tx) {
	h := strings.ReplaceAll(
		StrLower(
			StrTrimWS(appData.Handle)), " ", "_")
	validate.Str("handle", h, tlbx, handleMinLen, handleMaxLen, handleRegex)
	a := StrTrimWS(appData.Alias)
	validate.Str("alias", a, tlbx, 0, aliasMaxLen)
	qryArgs := sqlh.NewArgs(0)
	qry := qryInsert(qryArgs, &social.Socials{
		ID:        me,
		Handle:    h,
		Alias:     a,
		HasAvatar: false,
	})
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}

func OnDelete(tlbx app.Tlbx, me ID, tx sql.Tx) {
	service.Get(tlbx).Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
}

func New(
	onSetSocials func(app.Tlbx, sql.Tx, *social.Socials) error,
) []*app.Endpoint {
	return []*app.Endpoint{
		&app.Endpoint{
			Description:  "get users",
			Path:         (&user.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.Get{}
			},
			GetExampleArgs: func() interface{} {
				return &user.Get{
					Users: []ID{app.ExampleID()},
				}
			},
			GetExampleResponse: func() interface{} {
				var fcmEnabled *bool
				if enableFCM {
					fcmEnabled = ptr.Bool(true)
				}
				ex := []user.User{
					{
						ID:         app.ExampleID(),
						Handle:     ptr.String("bloe_joggs"),
						Alias:      ptr.String("Joe Bloggs"),
						HasAvatar:  ptr.Bool(true),
						FcmEnabled: fcmEnabled,
					},
				}
				return ex
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.Get)
				if len(args.Users) == 0 {
					return nil
				}
				validate.MaxIDs(tlbx, "users", args.Users, 1000)
				srv := service.Get(tlbx)
				query := bytes.NewBufferString(`SELECT id, handle, alias, hasAvatar, fcmEnabled FROM users WHERE id IN(?`)
				queryArgs := make([]interface{}, 0, len(args.Users))
				queryArgs = append(queryArgs, args.Users[0])
				for _, id := range args.Users[1:] {
					query.WriteString(`,?`)
					queryArgs = append(queryArgs, id)
				}
				query.WriteString(`)`)
				res := make([]*user.User, 0, len(args.Users))
				PanicOn(srv.User().Query(func(rows isql.Rows) {
					for rows.Next() {
						u := &user.User{}
						PanicOn(rows.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar, &u.FcmEnabled))
						res = append(res, u)
					}
				}, query.String(), queryArgs...))
				return res
			},
		}, &app.Endpoint{
			Description:  "set handle",
			Path:         (&user.SetHandle{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.SetHandle{}
			},
			GetExampleArgs: func() interface{} {
				return &user.SetHandle{
					Handle: "joe_bloggs",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.SetHandle)
				validate.Str("handle", args.Handle, tlbx, handleMinLen, handleMaxLen, handleRegex)
				srv := service.Get(tlbx)
				me := sessionGet(tlbx)
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getUser(tx, nil, &me)
				user.Handle = &args.Handle
				updateUser(tx, user)
				if onSetSocials != nil {
					PanicOn(onSetSocials(tlbx, &user.User))
				}
				tx.Commit()
				return nil
			},
		}, &app.Endpoint{
			Description:  "set alias",
			Path:         (&user.SetAlias{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.SetAlias{}
			},
			GetExampleArgs: func() interface{} {
				return &user.SetAlias{
					Alias: ptr.String("Boe Jloggs"),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.SetAlias)
				if args.Alias != nil {
					validate.Str("alias", *args.Alias, tlbx, 0, aliasMaxLen)
				}
				srv := service.Get(tlbx)
				me := sessionGet(tlbx)
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getUser(tx, nil, &me)
				user.Alias = args.Alias
				updateUser(tx, user)
				if onSetSocials != nil {
					PanicOn(onSetSocials(tlbx, &user.User))
				}
				tx.Commit()
				return nil
			},
		}, &app.Endpoint{
			Description:  "set avatar",
			Path:         (&user.SetAvatar{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.MB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &app.UpStream{}
			},
			GetExampleArgs: func() interface{} {
				return &app.UpStream{}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*app.UpStream)
				defer args.Content.Close()
				me := sessionGet(tlbx)
				srv := service.Get(tlbx)
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getUser(tx, nil, &me)
				content, err := ioutil.ReadAll(args.Content)
				PanicOn(err)
				args.Size = int64(len(content))
				if args.Size > 0 {
					if *user.HasAvatar {
						srv.Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
					}
					avatar, _, err := image.Decode(bytes.NewBuffer(content))
					PanicOn(err)
					bounds := avatar.Bounds()
					xDiff := bounds.Max.X - bounds.Min.X
					yDiff := bounds.Max.Y - bounds.Min.Y
					if xDiff != yDiff || xDiff != avatarDim || yDiff != avatarDim {
						avatar = imaging.Fill(avatar, avatarDim, avatarDim, imaging.Center, imaging.Lanczos)
					}
					buff := &bytes.Buffer{}
					PanicOn(png.Encode(buff, avatar))
					srv.Store().MustPut(
						AvatarBucket,
						store.Key(AvatarPrefix, me),
						args.Name,
						"image/png",
						int64(buff.Len()),
						true,
						false,
						bytes.NewReader(buff.Bytes()))
				} else if *user.HasAvatar == true {
					srv.Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
				}
				nowHasAvatar := args.Size > 0
				if *user.HasAvatar != nowHasAvatar {
					user.HasAvatar = ptr.Bool(nowHasAvatar)
					if onSetSocials != nil {
						PanicOn(onSetSocials(tlbx, &user.User))
					}
				}
				updateUser(tx, user)
				tx.Commit()
				return nil
			},
		},
		&app.Endpoint{
			Description:      "get avatar",
			Path:             (&user.GetAvatar{}).Path(),
			Timeout:          500,
			MaxBodyBytes:     app.KB,
			SkipXClientCheck: true,
			IsPrivate:        false,
			GetDefaultArgs: func() interface{} {
				return &user.GetAvatar{}
			},
			GetExampleArgs: func() interface{} {
				return &user.GetAvatar{
					User: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return &app.DownStream{}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.GetAvatar)
				srv := service.Get(tlbx)
				name, mimeType, size, content := srv.Store().MustGet(AvatarBucket, store.Key(AvatarPrefix, args.User))
				ds := &app.DownStream{}
				ds.ID = args.User
				ds.Name = name
				ds.Type = mimeType
				ds.Size = size
				ds.Content = content
				return ds
			},
		},
	}
}

var (
	handleRegex  = regexp.MustCompile(`\A[_a-z0-9]{1,20}\z`)
	handleMinLen = 3
	handleMaxLen = 20
	emailRegex   = regexp.MustCompile(`\A.+@.+\..+\z`)
	emailMaxLen  = 250
	aliasMaxLen  = 50
	pwdRegexs    = []*regexp.Regexp{
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
	avatarDim     = 250
	exampleJin    = json.MustFromString(`{"v":1, "saveDir":"/my/save/dir", "startTab":"favourites"}`)
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

type fullUser struct {
	user.User
	Email           string
	RegisteredOn    time.Time
	ActivatedOn     time.Time
	NewEmail        *string
	ActivateCode    *string
	ChangeEmailCode *string
	LastPwdResetOn  *time.Time
}

func getUser(tx sql.Tx, email *string, id *ID) *fullUser {
	PanicIf(email == nil && id == nil, "one of email or id must not be nil")
	query := `SELECT id, email, handle, alias, hasAvatar, fcmEnabled, registeredOn, activatedOn, newEmail, activateCode, changeEmailCode, lastPwdResetOn FROM users WHERE `
	var arg interface{}
	if email != nil {
		query += `email=?`
		arg = *email
	} else {
		query += `id=?`
		arg = *id
	}
	row := tx.QueryRow(query, arg)
	res := &fullUser{}
	err := row.Scan(&res.ID, &res.Email, &res.Handle, &res.Alias, &res.HasAvatar, &res.FcmEnabled, &res.RegisteredOn, &res.ActivatedOn, &res.NewEmail, &res.ActivateCode, &res.ChangeEmailCode, &res.LastPwdResetOn)
	if err == isql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateUser(tx sql.Tx, user *fullUser) {
	_, err := tx.Exec(`UPDATE users SET email=?, handle=?, alias=?, hasAvatar=?, fcmEnabled=?, registeredOn=?, activatedOn=?, newEmail=?, activateCode=?, changeEmailCode=?, lastPwdResetOn=? WHERE id=?`, user.Email, user.Handle, user.Alias, user.HasAvatar, user.FcmEnabled, user.RegisteredOn, user.ActivatedOn, user.NewEmail, user.ActivateCode, user.ChangeEmailCode, user.LastPwdResetOn, user.ID)
	PanicOn(err)
}

type pwd struct {
	ID   ID
	Salt []byte
	Pwd  []byte
	N    int
	R    int
	P    int
}

func getPwd(srv service.Layer, id ID) *pwd {
	row := srv.Auth().QueryRow(`SELECT id, salt, pwd, n, r, p FROM pwds WHERE id=?`, id)
	res := &pwd{}
	err := row.Scan(&res.ID, &res.Salt, &res.Pwd, &res.N, &res.R, &res.P)
	if err == isql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func setPwd(tlbx app.Tlbx, pwdtx sql.Tx, id ID, pwd, confirmPwd string) {
	app.BadReqIf(pwd != confirmPwd, "pwds do not match")
	validate.Str("pwd", pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
	salt := crypt.Bytes(scryptSaltLen)
	pwdBs := crypt.ScryptKey([]byte(pwd), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	_, err := pwdtx.Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE salt=VALUE(salt), pwd=VALUE(pwd), n=VALUE(n), r=VALUE(r), p=VALUE(p)`, id, salt, pwdBs, scryptN, scryptR, scryptP)
	PanicOn(err)
}
