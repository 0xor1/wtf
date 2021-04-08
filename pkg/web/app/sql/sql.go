package sql

import (
	"database/sql"
	"net/http"
	"strings"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

func NewArgs(size int) *Args {
	return &Args{
		args: make([]interface{}, 0, size),
	}
}

type Args struct {
	args []interface{}
}

func (a *Args) AppendOne(arg interface{}) {
	a.args = append(a.args, arg)
}

func (a *Args) Append(arg ...interface{}) {
	a.args = append(a.args, arg...)
}

func (a *Args) Is() []interface{} {
	return a.args
}

func ReturnNotFoundIfIsNoRows(err error) {
	app.ReturnIf(IsNoRows(err), http.StatusNotFound, "")
	PanicOn(err)
}

func PanicIfIsntNoRows(err error) {
	if !IsNoRows(err) {
		PanicOn(err)
	}
}

func IsNoRows(err error) bool {
	return err != nil && err == sql.ErrNoRows
}

func Asc(asc bool) string {
	if asc {
		return ` ASC`
	}
	return ` DESC`
}

func GtLtSymbol(asc bool) string {
	if asc {
		return ">"
	} else {
		return "<"
	}
}

func Limit(l, max uint16) uint16 {
	switch {
	case l >= max:
		return max + 1
	case l < 1:
		return 2 // 1 + 1 for "more": true/false detection
	default:
		return l + 1
	}
}

func Limit100(l uint16) uint16 {
	return Limit(l, 100)
}

func OrderLimit(field string, asc bool, l, max uint16) string {
	return Strf(` ORDER BY %s %s LIMIT %d`, field, Asc(asc), Limit(l, max))
}

func OrderLimit100(field string, asc bool, l uint16) string {
	return OrderLimit(field, asc, l, 100)
}

func InCondition(and bool, field string, setLen int) string {
	PanicIf(setLen <= 0, "setLen must be > 0")
	op := `AND`
	if !and {
		op = `OR`
	}
	return Strf(` %s %s IN (%s)`, op, field, PList(setLen))
}

func OrderByField(field string, setLen int) string {
	if setLen <= 0 {
		return ``
	}
	return Strf(` ORDER BY FIELD (%s,%s)`, field, PList(setLen))
}

func PList(count int) string {
	PanicIf(count < 1, `count must be >= 1`)
	return `?` + strings.Repeat(`,?`, count-1)
}
