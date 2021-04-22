package usereps

import (
	"os/user"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/user/social"
	"github.com/0xor1/tlbx/pkg/web/app/user/social/socialeps"
)

func New(
	fromEmail,
	activateFmtLink,
	confirmChangeEmailFmtLink string,
	onActivate func(app.Tlbx, *user.User),
	onDelete func(app.Tlbx, ID),
	onSetSocials func(app.Tlbx, *social.Social) error,
	validateFcmTopic func(app.Tlbx, IDs) (sql.Tx, error),
	enableJin bool,
) []*app.Endpoint {

	eps := autheps.New(
		fromEmail,
		activateFmtLink,
		confirmChangeEmailFmtLink,
		func(c *autheps.Config) {
			c.AppDataDefault = socialeps.AppDataDefault
			c.AppDataExample = socialeps.AppDataExample
			c.OnRegister = func(tlbx app.Tlbx, me ID, ad interface{}, txAdder sql.DoTxAdder) {
				appData := ad.(*social.RegisterAppData)
				socialeps.OnRegister(tlbx, me, appData, service.Get(tlbx).User().WriteTx())
			}
			c.OnDelete = func(tlbx app.Tlbx, me ID, txAdder sql.DoTxAdder) {
				socialeps.OnDelete(tlbx, me, service.Get(tlbx).User().WriteTx())
			}
		})

	return eps
}
