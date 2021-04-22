package config

import (
	"github.com/0xor1/tlbx/pkg/web/app/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("sql.users.primary", "todo_users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/todo_users?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("sql.data.primary", "todo_data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/todo_data?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
