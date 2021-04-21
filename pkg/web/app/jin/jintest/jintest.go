package jintest

import (
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/jin"
	"github.com/0xor1/tlbx/pkg/web/app/jin/jineps"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.GetProcessed(config.GetBase()),
		jineps.Eps,
		ratelimit.MeMware,
		nil,
		true,
		nil,
		func(c *autheps.Config) {
			c.OnRegister = func(tlbx app.Tlbx, me ID, ad interface{}, txAdder sql.DoTxAdder) {
				txAdder.Add(service.Get(tlbx).Data().Begin(), func(tx sql.Tx) {
					autheps.OnRegister(tlbx, me, tx)
				})
			}
			c.OnActivate = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
				txAdder.Add(service.Get(tlbx).Data().Begin(), func(tx sql.Tx) {
					autheps.OnActivate(tlbx, me, tx)
				})
			}
			c.OnDelete = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
				txAdder.Add(service.Get(tlbx).Data().Begin(), func(tx sql.Tx) {
					autheps.OnDelete(tlbx, me, tx)
				})
			}
		})
	defer r.CleanUp()

	ac := r.Ali().Client()

	js := (&jin.Get{}).MustDo(ac)
	a.Nil(js)

	(&jin.Set{
		Val: json.MustFromString(`{"test":"yolo"}`),
	}).MustDo(ac)

	js = (&jin.Get{}).MustDo(ac)
	a.Equal("yolo", js.MustString("test"))

	(&jin.Set{}).MustDo(ac)

	js = (&jin.Get{}).MustDo(ac)
	a.Nil(js)
}
