package fcmtest

import (
	"testing"
)

func Everything(t *testing.T) {
	// r := test.NewRig(
	// 	config.GetProcessed(config.GetBase()),
	// 	nil,
	// 	func(tlbx app.Tlbx, user *user.User) {},
	// 	func(tlbx app.Tlbx, id ID) {},
	// 	usereps.NopOnSetSocials,
	// 	func(t app.Tlbx, i IDs) (sql.Tx, error) {
	// 		tx := service.Get(t).Auth().Begin()
	// 		return tx, nil
	// 	},
	// 	true)
	// defer r.CleanUp()

	// a := assert.New(t)
	// c := r.NewClient()

	// // test fcm eps
	// ac := r.Ali().Client()
	// fcmToken := "123:abc"
	// (&user.SetFCMEnabled{
	// 	Val: false,
	// }).MustDo(ac)

	// (&user.SetFCMEnabled{
	// 	Val: true,
	// }).MustDo(ac)

	// client1 := (&user.RegisterForFCM{
	// 	Topic: IDs{app.ExampleID()},
	// 	Token: fcmToken,
	// }).MustDo(ac)
	// a.NotNil(client1)

	// idGen := NewIDGen()
	// // using client1 so this should overwrite existing fcmTokens row.
	// client2 := (&user.RegisterForFCM{
	// 	Topic:  IDs{idGen.MustNew(), idGen.MustNew()},
	// 	Client: client1,
	// 	Token:  fcmToken,
	// }).MustDo(ac)
	// a.True(client1.Equal(*client2))

	// client2 = (&user.RegisterForFCM{
	// 	Topic: IDs{idGen.MustNew(), idGen.MustNew()},
	// 	Token: fcmToken,
	// }).MustDo(ac)
	// a.False(client1.Equal(*client2))

	// client2 = (&user.RegisterForFCM{
	// 	Topic: IDs{idGen.MustNew(), idGen.MustNew()},
	// 	Token: fcmToken,
	// }).MustDo(ac)
	// a.False(client1.Equal(*client2))

	// client2 = (&user.RegisterForFCM{
	// 	Topic: IDs{idGen.MustNew(), idGen.MustNew()},
	// 	Token: fcmToken,
	// }).MustDo(ac)
	// a.False(client1.Equal(*client2))

	// //registered to 5 topics now which is max allowed
	// client2 = (&user.RegisterForFCM{
	// 	Topic: IDs{idGen.MustNew(), idGen.MustNew()},
	// 	Token: fcmToken,
	// }).MustDo(ac)
	// a.False(client1.Equal(*client2))

	// // this 6th topic should cause the oldest to be bumped out
	// // leaving this as the newest of the allowed 5
	// client2 = (&user.RegisterForFCM{
	// 	Topic: IDs{idGen.MustNew(), idGen.MustNew()},
	// 	Token: fcmToken,
	// }).MustDo(ac)
	// a.False(client1.Equal(*client2))

	// // toggle off and back on to test sending fcmEnabled:true/false data push
	// (&user.SetFCMEnabled{
	// 	Val: false,
	// }).MustDo(ac)

	// (&user.SetFCMEnabled{
	// 	Val: true,
	// }).MustDo(ac)

	// (&user.UnregisterFromFCM{
	// 	Client: app.ExampleID(),
	// }).MustDo(ac)
}
