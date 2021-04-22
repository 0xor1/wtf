package main

import (
	"github.com/0xor1/tlbx/cmd/games/pkg/blockers/blockerseps"
	"github.com/0xor1/tlbx/cmd/games/pkg/config"
	"github.com/0xor1/tlbx/cmd/games/pkg/game"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
)

func main() {
	config := config.Get()
	app.Run(func(c *app.Config) {
		c.StaticDir = config.Web.StaticDir
		c.ContentSecurityPolicies = config.Web.ContentSecurityPolicies
		c.Name = "games"
		c.Description = "a web app to play turn based multiplayer games"
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
		c.Endpoints = app.JoinEps(game.Eps, blockerseps.Eps)
	})
}
