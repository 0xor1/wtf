package main

import (
	"github.com/0xor1/tlbx/cmd/games/pkg/blockers/blockerseps"
	"github.com/0xor1/tlbx/cmd/games/pkg/config"
	"github.com/0xor1/tlbx/cmd/games/pkg/game"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
)

func main() {
	config := config.Get()
	if config.IsLocal {
		defer config.Store.(store.LocalClient).MustDeleteStore()
	}
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "games"
		c.Description = "a web app to play turn based multiplayer games"
		c.TlbxMwares = app.TlbxMwares{
			session.BasicMware(config.SessionAuthKey64s, config.SessionEncrKey32s, config.IsLocal),
			me.RateLimitMware(config.Cache),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store),
		}
		c.Log = config.Log
		c.Endpoints = append(append(c.Endpoints, game.Eps...), blockerseps.Eps...)
	})
}
