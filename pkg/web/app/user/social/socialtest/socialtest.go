package socialtest

import (
	"encoding/base64"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/user/social"
	"github.com/0xor1/tlbx/pkg/web/app/user/social/socialeps"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.GetProcessed(config.GetBase()),
		socialeps.New(),
		ratelimit.MeMware,
		[]string{socialeps.AvatarBucket},
		true,
		func(name string, reg *auth.Register) {
			reg.AppData = &social.RegisterAppData{
				Handle: name,
				Alias:  name,
			}
		},
		func(c *autheps.Config) {
			c.AppDataDefault = socialeps.AppDataDefault
			c.AppDataExample = socialeps.AppDataExample
			c.OnRegister = func(tlbx app.Tlbx, me ID, ad interface{}, txAdder sql.DoTxAdder) {
				txAdder.Add(service.Get(tlbx).User().WriteTx(), func(tx sql.Tx) {
					appData := ad.(*social.RegisterAppData)
					socialeps.OnRegister(tlbx, me, appData, tx)
				})
			}
			c.OnDelete = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
				txAdder.Add(service.Get(tlbx).User().WriteTx(), func(tx sql.Tx) {
					socialeps.OnDelete(tlbx, me, tx)
				})
			}
		})
	defer r.CleanUp()

	c := r.Ali().Client()

	socials := (&social.Get{
		IDs: IDs{
			r.Dan().ID(),
			r.Ali().ID(),
			r.Bob().ID(),
			r.Cat().ID(),
		},
	}).MustDo(c)
	a.Equal(4, len(socials.Set))
	a.True(socials.Set[0].ID.Equal(r.Dan().ID()))
	a.True(socials.Set[1].ID.Equal(r.Ali().ID()))
	a.True(socials.Set[2].ID.Equal(r.Bob().ID()))
	a.True(socials.Set[3].ID.Equal(r.Cat().ID()))
	a.False(socials.More)

	me := (&social.GetMe{}).MustDo(c)
	a.Equal("ali", me.Handle)
	a.Equal("ali", me.Alias)
	a.False(me.HasAvatar)

	handle := "new_" + r.Unique()
	alias := "shabba!"
	(&social.Update{
		Handle: &field.String{V: handle},
		Alias:  &field.String{V: alias},
	}).MustDo(c)

	me = (&social.GetMe{}).MustDo(c)
	a.Equal(handle, me.Handle)
	a.Equal(alias, me.Alias)
	a.False(me.HasAvatar)

	(&social.SetAvatar{
		Avatar: ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))),
	}).MustDo(c)

	me = (&social.GetMe{}).MustDo(c)
	a.True(me.HasAvatar)

	avatar := (&social.GetAvatar{
		ID: me.ID,
	}).MustDo(c)
	a.Equal("image/png", avatar.Type)
	a.True(me.ID.Equal(avatar.ID))
	a.False(avatar.IsDownload)
	a.Equal(int64(126670), avatar.Size)
	avatar.Content.Close()

	(&social.SetAvatar{
		Avatar: ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgNotSquare))),
	}).MustDo(c)

	me = (&social.GetMe{}).MustDo(c)
	a.True(me.HasAvatar)

	(&social.SetAvatar{
		Avatar: nil,
	}).MustDo(c)

	me = (&social.GetMe{}).MustDo(c)
	a.False(me.HasAvatar)
}
