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
	tokens := getDistinctTokens(tx, me)
	qryArgs := sqlh.NewArgs(0)
	qry := qryDelete(qryArgs, me, nil, nil)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
	srv.FCM().RawAsyncSend("logout", tokens, map[string]string{}, 0)
}

func New(
	validateFcmTopic func(app.Tlbx, IDs) sql.Tx,
) []*app.Endpoint {
	return []*app.Endpoint{
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
				enabled := getEnabled(tx, me)
				if enabled == args.Val {
					// not changing anything
					return nil
				}
				qryArgs := sqlh.NewArgs(0)
				qry := qrySetEnabled(qryArgs, me, args.Val)
				_, err := tx.Exec(qry, qryArgs.Is()...)
				PanicOn(err)
				tokens := getDistinctTokens(tx, me)
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
				tx := service.Get(tlbx).User().WriteTx()
				defer tx.Rollback()
				enabled := getEnabled(tx, me)
				app.BadReqIf(!enabled, "fcm not enabled for user, please enable first then register for topics")
				// this query is used to get a users 5th token createdOn value if they have one
				qryArgs := sqlh.NewArgs(0)
				qry := qryFifthYoungestToken(qryArgs, me)
				row := tx.QueryRow(qry, qryArgs.Is()...)
				fifthYoungestTokenCreatedOn := time.Time{}
				sqlh.PanicIfIsntNoRows(row.Scan(&fifthYoungestTokenCreatedOn))
				if !fifthYoungestTokenCreatedOn.IsZero() {
					// this user has 5 topics they're subscribed too already so delete the older ones
					// to make room for this new one
					deleteTokens(tx, me, nil, &fifthYoungestTokenCreatedOn)
				}
				appTx := validateFcmTopic(tlbx, args.Topic)
				if appTx != nil {
					defer appTx.Rollback()
				}
				qry = qryInsert(qryArgs, args.Topic.StrJoin("_"), args.Token, me, *client, tlbx.Start())
				_, err := tx.Exec(qry, qryArgs.Is()...)
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
				tx := service.Get(tlbx).User().WriteTx()
				defer tx.Rollback()
				deleteTokens(tx, me, &args.Client, nil)
				tx.Commit()
				return nil
			},
		},
	}
}

func getEnabled(tx sql.Tx, me ID) bool {
	qryArgs := sqlh.NewArgs(0)
	qry := qryGetEnabled(qryArgs, me)
	row := tx.QueryRow(qry, qryArgs.Is()...)
	enabled := false
	sqlh.PanicIfIsntNoRows(row.Scan(&enabled))
	return enabled
}

func getDistinctTokens(tx sql.Tx, me ID) []string {
	tokens := make([]string, 0, 5)
	qryArgs := sqlh.NewArgs(0)
	qry := qryDistinctTokens(qryArgs, me)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			token := ""
			PanicOn(rows.Scan(&token))
			tokens = append(tokens, token)
		}
	}, qry, qryArgs.Is()...))
	return tokens
}

func deleteTokens(tx sql.Tx, me ID, client *ID, createdOn *time.Time) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryDelete(qryArgs, me, client, createdOn)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}
