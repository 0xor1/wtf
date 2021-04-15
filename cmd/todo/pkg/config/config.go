package config

import (
	"github.com/0xor1/tlbx/pkg/web/app/config"
)

func Get(file ...string) *config.Config {
	c := config.GetBase(file...)
	c.SetDefault("sql.auth.primary", "todo_auth:C0-Mm-0n-Auth@tcp(localhost:3306)/todo_auth?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("sql.data.primary", "todo_data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/todo_data?parseTime=true&loc=UTC&multiStatements=true")
	return config.GetProcessed(c)
}
