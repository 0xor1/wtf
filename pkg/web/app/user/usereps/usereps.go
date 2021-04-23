package usereps

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/user/fcm/fcmeps"
	"github.com/0xor1/tlbx/pkg/web/app/user/jin/jineps"
	"github.com/0xor1/tlbx/pkg/web/app/user/social"
	"github.com/0xor1/tlbx/pkg/web/app/user/social/socialeps"
)

type Config struct {
	EnableSocial        bool
	GetSocialRegAppData func(appData interface{}) *social.RegisterAppData
	OnSetSocial         func(app.Tlbx, *social.Social, sql.DoTxAdder)
	EnableJin           bool
	ValidateFcmTopic    func(app.Tlbx, IDs, sql.DoTxAdder)
	// autheps configs
	AppDataDefault func() interface{}
	AppDataExample func() interface{}
	OnRegister     func(tlbx app.Tlbx, me ID, appData interface{}, txAdder sql.DoTxAdder)
	OnActivate     func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder)
	OnDelete       func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder)
	OnLogout       func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder)
}

func AuthExtensions(
	configs ...func(*Config),
) ([]*app.Endpoint, func(*autheps.Config)) {
	uc := config(configs...)
	c := func(c *autheps.Config) {
		c.AppDataDefault = socialeps.AppDataDefault
		c.AppDataExample = socialeps.AppDataExample
		c.OnRegister = func(tlbx app.Tlbx, me ID, ad interface{}, txAdder sql.DoTxAdder) {
			if uc.EnableSocial {
				socialeps.OnRegister(tlbx, me, uc.GetSocialRegAppData(ad), service.Get(tlbx).User().WriteTx())
			}
			uc.OnRegister(tlbx, me, ad, txAdder)
		}
		c.OnLogout = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
			if uc.ValidateFcmTopic != nil {
				txAdder.Add(service.Get(tlbx).User().WriteTx(), func(tx sql.Tx) {
					fcmeps.OnLogout(tlbx, me, tx)
				})
			}
			uc.OnLogout(tlbx, me, txAdder)
		}
		c.OnDelete = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
			if uc.EnableSocial {
				socialeps.OnDelete(tlbx, me, service.Get(tlbx).User().WriteTx())
			}
			uc.OnDelete(tlbx, me, txAdder)
		}
	}
	var eps []*app.Endpoint
	if uc.EnableSocial {
		eps = app.JoinEps(eps, socialeps.New(func(c *socialeps.Config) {
			c.OnSetSocial = uc.OnSetSocial
		}))
	}
	if uc.EnableJin {
		eps = app.JoinEps(eps, jineps.Eps)
	}
	if uc.ValidateFcmTopic != nil {
		eps = app.JoinEps(eps, fcmeps.New(uc.ValidateFcmTopic))
	}
	return eps, c
}

func config(configs ...func(*Config)) *Config {
	noopDoTx := func(_ app.Tlbx, _ ID, _ sql.DoTxAdder) {}
	c := &Config{
		EnableSocial: true,
		GetSocialRegAppData: func(appData interface{}) *social.RegisterAppData {
			return appData.(*social.RegisterAppData)
		},
		OnSetSocial:      func(_ app.Tlbx, _ *social.Social, _ sql.DoTxAdder) {},
		EnableJin:        false,
		ValidateFcmTopic: nil,
		// autheps configs
		AppDataDefault: func() interface{} {
			return &social.RegisterAppData{}
		},
		AppDataExample: func() interface{} {
			return &social.RegisterAppData{
				Handle: "bloe_joggs",
				Alias:  "joe bloggs",
			}
		},
		OnRegister: func(_ app.Tlbx, _ ID, _ interface{}, _ sql.DoTxAdder) {},
		OnActivate: noopDoTx,
		OnDelete:   noopDoTx,
		OnLogout:   noopDoTx,
	}
	for _, config := range configs {
		config(c)
	}
	return c
}
