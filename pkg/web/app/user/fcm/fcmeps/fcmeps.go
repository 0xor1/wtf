package fcmeps

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=fcmeps.sql

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user/fcm"
)

func OnLogout(tlbx app.Tlbx, me ID, tx sql.Tx) {
	srv := service.Get(tlbx)
	tokens := make([]string, 0, 5)
	qryArgs := sqlh.NewArgs(0)
	qry := qryDistinctTokens(qryArgs, me)
	tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			token := ""
			PanicOn(rows.Scan(&token))
			tokens = append(tokens, token)
		}
	}, qry, qryArgs.Is()...)
	qry = qryDelete(qryArgs, me)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
	srv.FCM().RawAsyncSend("logout", tokens, map[string]string{}, 0)
}

var (
	Eps = []*app.Endpoint{
		{
			Description:  "set fcm enabled",
			Path:         (&fcm.SetEnabled{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &fcm.SetEnabled{
					Val: true,
				}
			},
			GetExampleArgs: func() interface{} {
				return &fcm.SetEnabled{
					Val: true,
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*fcm.SetEnabled)
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).User().WriteTx()
				defer tx.Rollback()
				f := getUser(tx, nil, &me)
				if f.FcmEnabled == args.Val {
					// not changing anything
					return nil
				}
				u.FcmEnabled = &args.Val
				updateUser(tx, u)
				tokens := make([]string, 0, 5)
				tx.Query(func(rows isql.Rows) {
					for rows.Next() {
						token := ""
						PanicOn(rows.Scan(&token))
						tokens = append(tokens, token)
					}
				}, `SELECT DISTINCT token FROM fcmTokens WHERE user=?`, me)
				tx.Commit()
				if len(tokens) == 0 {
					// no tokens to notify
					return nil
				}
				fcmType := "enabled"
				if !args.Val {
					fcmType = "disabled"
				}
				service.Get(tlbx).FCM().RawAsyncSend(fcmType, tokens, map[string]string{}, 0)
				return nil
			},
		},
		{
			Description:  "register for fcm",
			Path:         (&fcm.Register{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &fcm.Register{}
			},
			GetExampleArgs: func() interface{} {
				return &fcm.Register{
					Topic:  IDs{app.ExampleID()},
					Client: ptr.ID(app.ExampleID()),
					Token:  "abc:123",
				}
			},
			GetExampleResponse: func() interface{} {
				return app.ExampleID()
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*fcm.Register)
				app.BadReqIf(len(args.Topic) == 0 || len(args.Topic) > 5, "topic must contain 1 to 5 ids")
				app.BadReqIf(args.Token == "", "empty string is not a valid fcm token")
				client := args.Client
				if client == nil {
					client = ptr.ID(tlbx.NewID())
				}
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).User().Begin()
				defer tx.Rollback()
				u := getUser(tx, nil, &me)
				app.BadReqIf(u.FcmEnabled == nil || !*u.FcmEnabled, "fcm not enabled for user, please enable first then register for topics")
				// this query is used to get a users 5th token createdOn value if they have one
				row := tx.QueryRow(`SELECT createdOn FROM fcmTokens WHERE user=? ORDER BY createdOn DESC LIMIT 4, 1`, me)
				fifthYoungestTokenCreatedOn := time.Time{}
				sqlh.PanicIfIsntNoRows(row.Scan(&fifthYoungestTokenCreatedOn))
				if !fifthYoungestTokenCreatedOn.IsZero() {
					// this user has 5 topics they're subscribed too already so delete the older ones
					// to make room for this new one
					_, err := tx.Exec(`DELETE FROM fcmTokens WHERE user=? AND createdOn<=?`, me, fifthYoungestTokenCreatedOn)
					PanicOn(err)
				}
				appTx, err := validateFcmTopic(tlbx, args.Topic)
				if appTx != nil {
					defer appTx.Rollback()
				}
				PanicOn(err)
				_, err = tx.Exec(`INSERT INTO fcmTokens (topic, token, user, client, createdOn) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE topic=VALUES(topic), token=VALUES(token), user=VALUES(user), client=VALUES(client), createdOn=VALUES(createdOn)`, args.Topic.StrJoin("_"), args.Token, me, client, tlbx.Start())
				PanicOn(err)
				tx.Commit()
				if appTx != nil {
					appTx.Commit()
				}
				return client
			},
		},
		{
			Description:      "unregister from fcm",
			SkipXClientCheck: true,
			Path:             (&fcm.Unregister{}).Path(),
			Timeout:          500,
			MaxBodyBytes:     app.KB,
			IsPrivate:        false,
			GetDefaultArgs: func() interface{} {
				return &fcm.Unregister{}
			},
			GetExampleArgs: func() interface{} {
				return &fcm.Unregister{
					Client: app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*fcm.Unregister)
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).User().Begin()
				defer tx.Rollback()
				_, err := tx.Exec(`DELETE FROM fcmTokens WHERE user=? AND client=?`, me, args.Client)
				PanicOn(err)
				tx.Commit()
				return nil
			},
		},
	}
)
