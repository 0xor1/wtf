package jineps

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=jineps.sql

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/jin"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/sql"

	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "set users jin (json bin), adhoc json content",
			Path:         (&jin.Set{}).Path(),
			Timeout:      500,
			MaxBodyBytes: 10 * app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &jin.Set{}
			},
			GetExampleArgs: func() interface{} {
				return &jin.Set{
					Val: exampleJin,
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*jin.Set)
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				qryArgs := sql.NewArgs(0)
				var qry string
				if args.Val == nil {
					qry = qryDelete(qryArgs, me)
				} else {
					qry = qryInsert(qryArgs, me, args.Val)
				}
				_, err := srv.Data().Exec(qry, qryArgs.Is()...)
				PanicOn(err)
				return nil
			},
		},
		{
			Description:  "get users jin (json bin), adhoc json content",
			Path:         (&jin.Get{}).Path(),
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
				return exampleJin
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				me := me.AuthedGet(tlbx)
				srv := service.Get(tlbx)
				res := &json.Json{}
				qryArgs := sql.NewArgs(0)
				qry := qrySelect(qryArgs, me)
				sqlh.PanicIfIsntNoRows(srv.Data().QueryRow(qry, qryArgs.Is()...).Scan(&res))
				return res
			},
		},
	}
	exampleJin = json.MustFromString(`{"v":1, "saveDir":"/my/save/dir", "startTab":"favourites"}`)
)
