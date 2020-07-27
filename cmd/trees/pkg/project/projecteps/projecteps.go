package projecteps

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/user"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new project",
			Path:         (&project.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Create{
					Base: project.Base{
						HoursPerDay: 8,
						DaysPerWeek: 5,
						IsPublic:    false,
					},
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Create{
					Base: project.Base{
						HoursPerDay: 8,
						DaysPerWeek: 5,
						StartOn:     ptr.Time(app.ExampleTime()),
						DueOn:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
						IsPublic:    false,
					},
					Name: "My New Project",
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.Project{
					Task: project.Task{
						Name: "My New Project",
					},
					Base: project.Base{
						HoursPerDay: 8,
						DaysPerWeek: 5,
						StartOn:     ptr.Time(app.ExampleTime()),
						DueOn:       ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
						IsPublic:    false,
					},
					IsArchived: false,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				// args := a.(*project.Create)
				return nil
			},
		},
	}
	aliasMaxLen = 50
)

func OnActivate(tlbx app.Tlbx, me *user.User) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	// _, err := tx.Exec(`INSERT INTO accounts WHERE id=?`, me)
	// PanicOn(err)
	tx.Commit()
}

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	// _, err := tx.Exec(`DELETE FROM accounts WHERE id=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM times WHERE account=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM tasks WHERE account=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM projects WHERE account=?`, me)
	// PanicOn(err)
	// _, err = tx.Exec(`DELETE FROM projects WHERE account=?`, me)
	// PanicOn(err)
	tx.Commit()
}
