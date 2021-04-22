package fcmtest

import (
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/user/fcm"
	"github.com/0xor1/tlbx/pkg/web/app/user/fcm/fcmeps"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.GetProcessed(config.GetBase()),
		fcmeps.New(func(_ app.Tlbx, _ IDs) sql.Tx {
			return nil
		}),
		ratelimit.MeMware,
		nil,
		true,
		nil,
		func(c *autheps.Config) {
			c.OnLogout = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
				fcmeps.OnLogout(tlbx, me, service.Get(tlbx).User().WriteTx())
			}
		})
	defer r.CleanUp()

	ac := r.Ali().Client()

	// // test fcm eps
	fcmToken := "123:abc" + r.Unique()
	(&fcm.SetEnabled{
		Val: false,
	}).MustDo(ac)

	(&fcm.SetEnabled{
		Val: true,
	}).MustDo(ac)

	client1 := (&fcm.Register{
		Topic: IDs{app.ExampleID()},
		Token: fcmToken,
	}).MustDo(ac)
	a.NotNil(client1)

	idGen := NewIDGen()
	// using client1 so this should overwrite existing fcmTokens row.
	client2 := (&fcm.Register{
		Topic:  IDs{idGen.MustNew(), idGen.MustNew()},
		Client: client1,
		Token:  fcmToken,
	}).MustDo(ac)
	a.True(client1.Equal(*client2))

	client2 = (&fcm.Register{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	client2 = (&fcm.Register{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	client2 = (&fcm.Register{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	//registered to 5 topics now which is max allowed
	client2 = (&fcm.Register{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	// this 6th topic should cause the oldest to be bumped out
	// leaving this as the newest of the allowed 5
	client2 = (&fcm.Register{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	// c := r.NewClient()
	// (&auth.Login{
	// 	Email: r.Ali().Email(),
	// 	Pwd:   r.Ali().Pwd(),
	// }).MustDo(c)
	// (&auth.Logout{}).MustDo(c)

	// toggle off and back on to test sending fcmEnabled:true/false data push
	(&fcm.SetEnabled{
		Val: false,
	}).MustDo(ac)

	(&fcm.SetEnabled{
		Val: true,
	}).MustDo(ac)

	(&fcm.Unregister{
		Client: app.ExampleID(),
	}).MustDo(ac)
}
