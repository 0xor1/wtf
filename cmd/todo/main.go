package main

import (
	"github.com/0xor1/tlbx/cmd/todo/pkg/config"
	"github.com/0xor1/tlbx/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list/listeps"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
)

func main() {
	config := config.Get()
	eps := []*app.Endpoint{}
	app.Run(func(c *app.Config) {
		c.StaticDir = config.Web.StaticDir
		c.ContentSecurityPolicies = config.Web.ContentSecurityPolicies
		c.Name = "Todo"
		c.Description = "A simple Todo list application, create multiple lists with many items which can be marked complete or uncomplete"
		c.TlbxSetup = app.TlbxMwares{
			session.BasicMware(
				config.Web.Session.AuthKey64s,
				config.Web.Session.EncrKey32s,
				config.Web.Session.Secure),
			ratelimit.MeMware(config.Redis.RateLimit, config.Web.RateLimit),
			service.Mware(config.Redis.Cache, config.SQL.User, config.SQL.Pwd, config.SQL.Data, config.Email, config.Store, config.FCM),
		}
		c.Version = config.Version
		c.Log = config.Log
		c.Endpoints = append(
			append(
				append(
					eps,
					usereps.New(
						config.App.FromEmail,
						config.App.ActivateFmtLink,
						config.App.ConfirmChangeEmailFmtLink,
						me.AuthedExists,
						me.SetAuthed,
						me.GetAuthed,
						me.Del,
						nil,
						listeps.OnDelete,
						usereps.NopOnSetSocials,
						nil,
						false)...),
				listeps.Eps...),
			itemeps.Eps...)
	})
}
