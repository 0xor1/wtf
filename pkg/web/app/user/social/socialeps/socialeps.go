package socialeps

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=socialeps.sql

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user/social"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/disintegration/imaging"
)

const (
	AvatarBucket = "avatars"
	AvatarPrefix = ""
)

func AppDataDefault() interface{} {
	return &social.RegisterAppData{}
}
func AppDataExample() interface{} {
	return &social.RegisterAppData{
		Handle: "bloe_joggs",
		Alias:  "joe_bloggs",
	}
}

func OnRegister(tlbx app.Tlbx, me ID, appData *social.RegisterAppData, tx sql.Tx) {
	h := strings.ReplaceAll(
		StrLower(
			StrTrimWS(appData.Handle)), " ", "_")
	validate.Str("handle", h, tlbx, handleMinLen, handleMaxLen, handleRegex)
	a := StrTrimWS(appData.Alias)
	validate.Str("alias", a, tlbx, 0, aliasMaxLen)
	qryArgs := sqlh.NewArgs(0)
	qry := qryInsert(qryArgs, &social.Social{
		ID:        me,
		Handle:    h,
		Alias:     a,
		HasAvatar: false,
	})
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}

func OnDelete(tlbx app.Tlbx, me ID, tx sql.Tx) {
	// this needs to be called before autheps.OnDelete to ensure the users socials are still there
	s := getSocial(tx, me)
	if s != nil && s.HasAvatar {
		service.Get(tlbx).Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
	}
}

type Config struct {
	OnSetSocial func(app.Tlbx, *social.Social, sql.DoTxAdder)
}

func New(
	configs ...func(c *Config),
) []*app.Endpoint {
	c := config(configs...)
	return []*app.Endpoint{
		{
			Description:  "get my socials",
			Path:         (&social.GetMe{}).Path(),
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
				return exampleSocial
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).User().ReadTx()
				defer tx.Rollback()
				res := getSocial(tx, me)
				tx.Commit()
				return res
			},
		},
		{
			Description:  "get socials",
			Path:         (&social.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &social.Get{}
			},
			GetExampleArgs: func() interface{} {
				return &social.Get{
					HandlePrefix: "blo",
					Limit:        1,
				}
			},
			GetExampleResponse: func() interface{} {
				return &social.GetRes{
					Set: []*social.Social{
						exampleSocial,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*social.Get)
				tx := service.Get(tlbx).User().ReadTx()
				defer tx.Rollback()
				res := getSocials(tx, args)
				tx.Commit()
				return res
			},
		},
		{
			Description:  "update socials",
			Path:         (&social.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &social.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &social.Update{
					Handle: &field.String{V: "joe_bloggs"},
					Alias:  &field.String{V: "bloe joggs"},
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*social.Update)
				if args.Handle == nil && args.Alias == nil {
					// not updating anything
					return nil
				}
				if args.Handle != nil {
					args.Handle.V = StrTrimWS(args.Handle.V)
					validate.Str("handle", args.Handle.V, tlbx, handleMinLen, handleMaxLen, handleRegex)
				}
				if args.Alias != nil {
					args.Alias.V = StrTrimWS(args.Alias.V)
					validate.Str("alias", args.Alias.V, tlbx, 0, aliasMaxLen)
				}

				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				tx := srv.User().WriteTx()
				defer tx.Rollback()
				social := getSocial(tx, me)
				if args.Handle != nil {
					social.Handle = args.Handle.V
				}
				if args.Alias != nil {
					social.Alias = args.Alias.V
				}
				updateSocial(tx, social)
				appTxs := sql.NewDoTxs()
				c.OnSetSocial(tlbx, social, appTxs)
				defer appTxs.Rollback()
				appTxs.Do()
				appTxs.Commit()
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "set avatar",
			Path:         (&social.SetAvatar{}).Path(),
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
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				tx := srv.User().WriteTx()
				defer tx.Rollback()
				social := getSocial(tx, me)
				content, err := ioutil.ReadAll(args.Content)
				PanicOn(err)
				args.Size = int64(len(content))
				if args.Size > 0 {
					if social.HasAvatar {
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
				} else if social.HasAvatar == true {
					srv.Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
				}
				nowHasAvatar := args.Size > 0
				if social.HasAvatar != nowHasAvatar {
					social.HasAvatar = nowHasAvatar
					appTxs := sql.NewDoTxs()
					c.OnSetSocial(tlbx, social, appTxs)
					defer appTxs.Rollback()
					appTxs.Do()
					appTxs.Commit()
				}
				updateSocial(tx, social)
				tx.Commit()
				return nil
			},
		},
		{
			Description:      "get avatar",
			Path:             (&social.GetAvatar{}).Path(),
			Timeout:          500,
			MaxBodyBytes:     app.KB,
			SkipXClientCheck: true,
			IsPrivate:        false,
			GetDefaultArgs: func() interface{} {
				return &social.GetAvatar{}
			},
			GetExampleArgs: func() interface{} {
				return &social.GetAvatar{
					ID: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return &app.DownStream{}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*social.GetAvatar)
				srv := service.Get(tlbx)
				name, mimeType, size, content := srv.Store().MustGet(AvatarBucket, store.Key(AvatarPrefix, args.ID))
				ds := &app.DownStream{}
				ds.ID = args.ID
				ds.Name = name
				ds.Type = mimeType
				ds.Size = size
				ds.Content = content
				return ds
			},
		},
	}
}

func config(configs ...func(c *Config)) *Config {
	c := &Config{
		OnSetSocial: func(_ app.Tlbx, _ *social.Social, _ sql.DoTxAdder) {},
	}
	for _, config := range configs {
		config(c)
	}
	return c
}

var (
	handleRegex   = regexp.MustCompile(`\A[_a-z0-9]{1,20}\z`)
	handleMinLen  = 3
	handleMaxLen  = 20
	aliasMaxLen   = 50
	avatarDim     = 250
	exampleSocial = &social.Social{
		ID:        app.ExampleID(),
		Handle:    "bloe_joggs",
		Alias:     "Joe Bloggs",
		HasAvatar: true,
	}
)

func getSocial(tx sql.Tx, id ID) *social.Social {
	res := getSocials(tx, &social.Get{IDs: IDs{id}})
	if len(res.Set) == 1 {
		return res.Set[0]
	}
	return nil
}

func getSocials(tx sql.Tx, args *social.Get) *social.GetRes {
	qryArgs := sqlh.NewArgs(0)
	qry := qrySelect(qryArgs, args)
	res := &social.GetRes{
		Set:  make([]*social.Social, 0, args.Limit),
		More: false,
	}
	PanicOn(tx.Query(func(rows isql.Rows) {
		iLimit := int(args.Limit)
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == iLimit {
				res.More = true
				break
			}
			s := &social.Social{}
			PanicOn(rows.Scan(&s.ID, &s.Handle, &s.Alias, &s.HasAvatar))
			res.Set = append(res.Set, s)
		}
	}, qry, qryArgs.Is()...))
	return res
}

func updateSocial(tx sql.Tx, s *social.Social) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryUpdate(qryArgs, s)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}
