package jintest

import (
	"testing"

	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user/jin"
	"github.com/0xor1/tlbx/pkg/web/app/user/jin/jineps"
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
		nil)
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
