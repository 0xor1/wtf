// Code generated by qtc from "socialeps.sql". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line socialeps.sql:1
package socialeps

//line socialeps.sql:1
import . "github.com/0xor1/tlbx/pkg/core"

//line socialeps.sql:2
import "github.com/0xor1/tlbx/pkg/web/app"

//line socialeps.sql:3
import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"

//line socialeps.sql:4
import "github.com/0xor1/tlbx/pkg/web/app/user/social"

//line socialeps.sql:6
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line socialeps.sql:6
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line socialeps.sql:6
func streamqryInsert(qw422016 *qt422016.Writer, args *sqlh.Args, social *social.Social) {
//line socialeps.sql:7
	qw422016.N().S(`INSERT INTO socials( id, handle, alias, hasAvatar ) VALUES ( ?, ?, ?, ? ) `)
//line socialeps.sql:21
	*args = *sqlh.NewArgs(4)
	args.Append(
		social.ID,
		social.Handle,
		social.Alias,
		social.HasAvatar,
	)

//line socialeps.sql:29
}

//line socialeps.sql:29
func writeqryInsert(qq422016 qtio422016.Writer, args *sqlh.Args, social *social.Social) {
//line socialeps.sql:29
	qw422016 := qt422016.AcquireWriter(qq422016)
//line socialeps.sql:29
	streamqryInsert(qw422016, args, social)
//line socialeps.sql:29
	qt422016.ReleaseWriter(qw422016)
//line socialeps.sql:29
}

//line socialeps.sql:29
func qryInsert(args *sqlh.Args, social *social.Social) string {
//line socialeps.sql:29
	qb422016 := qt422016.AcquireByteBuffer()
//line socialeps.sql:29
	writeqryInsert(qb422016, args, social)
//line socialeps.sql:29
	qs422016 := string(qb422016.B)
//line socialeps.sql:29
	qt422016.ReleaseByteBuffer(qb422016)
//line socialeps.sql:29
	return qs422016
//line socialeps.sql:29
}

//line socialeps.sql:31
func streamqryUpdate(qw422016 *qt422016.Writer, args *sqlh.Args, social *social.Social) {
//line socialeps.sql:32
	qw422016.N().S(`UPDATE socials SET handle=?, alias=?, hasAvatar=? WHERE id=? `)
//line socialeps.sql:40
	*args = *sqlh.NewArgs(4)
	args.Append(
		social.Handle,
		social.Alias,
		social.HasAvatar,
		social.ID,
	)

//line socialeps.sql:48
}

//line socialeps.sql:48
func writeqryUpdate(qq422016 qtio422016.Writer, args *sqlh.Args, social *social.Social) {
//line socialeps.sql:48
	qw422016 := qt422016.AcquireWriter(qq422016)
//line socialeps.sql:48
	streamqryUpdate(qw422016, args, social)
//line socialeps.sql:48
	qt422016.ReleaseWriter(qw422016)
//line socialeps.sql:48
}

//line socialeps.sql:48
func qryUpdate(args *sqlh.Args, social *social.Social) string {
//line socialeps.sql:48
	qb422016 := qt422016.AcquireByteBuffer()
//line socialeps.sql:48
	writeqryUpdate(qb422016, args, social)
//line socialeps.sql:48
	qs422016 := string(qb422016.B)
//line socialeps.sql:48
	qt422016.ReleaseByteBuffer(qb422016)
//line socialeps.sql:48
	return qs422016
//line socialeps.sql:48
}

//line socialeps.sql:50
func streamqrySelect(qw422016 *qt422016.Writer, qryArgs *sqlh.Args, args *social.Get) {
//line socialeps.sql:53
	args.Limit = sqlh.Limit100(args.Limit)
	app.BadReqIf(len(args.IDs) > 100, "max ids to query is 100")
	app.BadReqIf(args.HandlePrefix != "" && StrLen(args.HandlePrefix) < 3, "min handlePrefix len is 3")
	app.BadReqIf(len(args.IDs) == 0 && args.HandlePrefix == "", "no query parameters provided please")
	*qryArgs = *sqlh.NewArgs(len(args.IDs) + 1)

//line socialeps.sql:58
	qw422016.N().S(`SELECT id, handle, alias, hasAvatar FROM socials WHERE `)
//line socialeps.sql:65
	switch true {
//line socialeps.sql:66
	case len(args.IDs) > 0:
//line socialeps.sql:66
		qw422016.N().S(`id IN (`)
//line socialeps.sql:67
		for i := range args.IDs {
//line socialeps.sql:67
			if i > 0 {
//line socialeps.sql:67
				qw422016.N().S(`,`)
//line socialeps.sql:67
			}
//line socialeps.sql:67
			qw422016.N().S(`?`)
//line socialeps.sql:67
		}
//line socialeps.sql:67
		qw422016.N().S(`) ORDER BY FIELD (id,`)
//line socialeps.sql:68
		for i := range args.IDs {
//line socialeps.sql:68
			if i > 0 {
//line socialeps.sql:68
				qw422016.N().S(`,`)
//line socialeps.sql:68
			}
//line socialeps.sql:68
			qw422016.N().S(`?`)
//line socialeps.sql:68
		}
//line socialeps.sql:68
		qw422016.N().S(`) `)
//line socialeps.sql:70
		*qryArgs = *sqlh.NewArgs(len(args.IDs) * 2)
		qryArgs.Append(args.IDs.ToIs()...)
		qryArgs.Append(args.IDs.ToIs()...)

//line socialeps.sql:74
	case args.HandlePrefix != "":
//line socialeps.sql:74
		qw422016.N().S(`handle LIKE ? ORDER BY handle ASC LIMIT ? `)
//line socialeps.sql:79
		*qryArgs = *sqlh.NewArgs(2)
		qryArgs.Append(sqlh.LikePrefix(args.HandlePrefix), args.Limit)

//line socialeps.sql:82
	}
//line socialeps.sql:82
	qw422016.N().S(` `)
//line socialeps.sql:84
}

//line socialeps.sql:84
func writeqrySelect(qq422016 qtio422016.Writer, qryArgs *sqlh.Args, args *social.Get) {
//line socialeps.sql:84
	qw422016 := qt422016.AcquireWriter(qq422016)
//line socialeps.sql:84
	streamqrySelect(qw422016, qryArgs, args)
//line socialeps.sql:84
	qt422016.ReleaseWriter(qw422016)
//line socialeps.sql:84
}

//line socialeps.sql:84
func qrySelect(qryArgs *sqlh.Args, args *social.Get) string {
//line socialeps.sql:84
	qb422016 := qt422016.AcquireByteBuffer()
//line socialeps.sql:84
	writeqrySelect(qb422016, qryArgs, args)
//line socialeps.sql:84
	qs422016 := string(qb422016.B)
//line socialeps.sql:84
	qt422016.ReleaseByteBuffer(qb422016)
//line socialeps.sql:84
	return qs422016
//line socialeps.sql:84
}
