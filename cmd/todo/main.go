package main

import (
	"github.com/0xor1/tlbx/cmd/todo/pkg/config"
	"github.com/0xor1/tlbx/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list/listeps"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth/autheps"
)

func main() {
	config := config.Get()
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
			service.Mware(config.Redis.Cache, config.SQL.Users, config.SQL.Data, config.Email, config.Store, config.FCM),
		}
		c.Version = config.Version
		c.Log = config.Log
		c.Endpoints = app.JoinEps(autheps.New(
			config.App.FromEmail,
			config.App.ActivateFmtLink,
			config.App.ConfirmChangeEmailFmtLink,
			func(c *autheps.Config) {
				c.OnDelete = listeps.OnDelete
			}), listeps.Eps, itemeps.Eps)
	})
}
