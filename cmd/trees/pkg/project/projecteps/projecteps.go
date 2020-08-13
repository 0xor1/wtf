package projecteps

import (
	"bytes"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/consts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
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
						CurrencyCode: "USD",
						HoursPerDay:  8,
						DaysPerWeek:  5,
						IsPublic:     false,
					},
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Create{
					Base: project.Base{
						CurrencyCode: "USD",
						HoursPerDay:  8,
						DaysPerWeek:  5,
						StartOn:      ptr.Time(app.ExampleTime()),
						DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
						IsPublic:     false,
					},
					Name: "My New Project",
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.Project{
					Task: task.Task{
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
				args := a.(*project.Create)
				me := me.Get(tlbx)
				validate.Str("name", args.Name, tlbx, 1, nameMaxLen)
				app.BadReqIf(args.HoursPerDay < 1 || args.HoursPerDay > 24, "invalid hoursPerDay must be > 0 and <= 24")
				app.BadReqIf(args.DaysPerWeek < 1 || args.DaysPerWeek > 7, "invalid daysPerWeek must be > 0 and <= 7")
				app.BadReqIf(args.StartOn != nil && args.DueOn != nil && args.StartOn.After(*args.Base.DueOn), "invalid startOn must be before dueOn")
				p := &project.Project{
					Task: task.Task{
						ID:         tlbx.NewID(),
						Name:       args.Name,
						CreatedBy:  me,
						CreatedOn:  NowMilli(),
						IsParallel: true,
					},
					Base:       args.Base,
					IsArchived: false,
				}
				srv := service.Get(tlbx)

				u := &user.User{}
				row := srv.User().QueryRow(`SELECT id, handle, alias, hasAvatar FROM users WHERE id=?`, me)
				PanicOn(row.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar))

				tx := srv.Data().Begin()
				defer tx.Rollback()
				tx.Exec(`INSERT INTO projectLocks (host, id) VALUES (?, ?)`, me, p.ID)
				tx.Exec(`INSERT INTO projectUsers (host, project, id, handle, alias, hasAvatar, isActive, estimatedTime, loggedTime, estimatedExpense, loggedExpense, fileCount, fileSize, role) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, me, p.ID, me, u.Handle, u.Alias, u.HasAvatar, true, 0, 0, 0, 0, 0, 0, consts.RoleAdmin)
				tx.Exec(`INSERT INTO projects (host, id, isArchived, name, createdOn, currencyCode, hoursPerDay, daysPerWeek, startOn, dueOn, isPublic) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, me, p.ID, p.IsArchived, p.Name, p.CreatedOn, p.CurrencyCode, p.HoursPerDay, p.DaysPerWeek, p.StartOn, p.DueOn, p.IsPublic)
				tx.Exec(`INSERT INTO tasks (host, project, id, parent, firstChild, nextSibling, user, name, description, isParallel, createdBy, createdOn, minimumRemainingTime, estimatedTime, loggedTime, estimatedSubTime, loggedSubTime, estimatedExpense, loggedExpense, estimatedSubExpense, loggedSubExpense, fileCount, fileSize, subFileCount, subFileSize, childCount, descendantCount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, me, p.ID, p.ID, p.Parent, p.FirstChild, p.NextSibling, p.User, p.Name, p.Description, p.IsParallel, p.CreatedBy, p.CreatedOn, p.MinimumRemainingTime, p.EstimatedTime, p.LoggedTime, p.EstimatedSubTime, p.LoggedSubTime, p.EstimatedExpense, p.LoggedExpense, p.EstimatedSubExpense, p.LoggedSubExpense, p.FileCount, p.FileSize, p.SubFileCount, p.SubFileSize, p.ChildCount, p.DescendantCount)
				tx.Exec(`INSERT INTO projectActivities(host, project, occurredOn, user, item, itemType, itemHasBeenDeleted, action, itemName, extraInfo) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, me, p.ID, NowMilli(), me, p.ID, consts.TypeProject, false, consts.ActionCreated, p.Name, nil)
				tx.Commit()
				return p
			},
		},
		{
			Description:  "Get a project set",
			Path:         (&project.Get{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &project.Get{
					IsArchived: false,
					Sort:       consts.SortCreatedOn,
					Asc:        ptr.Bool(true),
					Limit:      ptr.Int(100),
				}
			},
			GetExampleArgs: func() interface{} {
				return &project.Get{
					IsArchived:     false,
					NameStartsWith: ptr.String("My Proj"),
					CreatedOnMin:   ptr.Time(app.ExampleTime()),
					CreatedOnMax:   ptr.Time(app.ExampleTime()),
					After:          ptr.ID(app.ExampleID()),
					Sort:           consts.SortName,
					Asc:            ptr.Bool(true),
					Limit:          ptr.Int(50),
				}
			},
			GetExampleResponse: func() interface{} {
				return &project.GetRes{
					Set: []*project.Project{
						exampleList,
					},
					More: true,
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				return getSet(tlbx, a.(*project.Get))
			},
		},
	}

	nameMaxLen  = 250
	aliasMaxLen = 50
	exampleList = &project.Project{
		Task: task.Task{
			ID:        app.ExampleID(),
			Name:      "My Project",
			CreatedOn: app.ExampleTime(),
		},
		Base: project.Base{
			CurrencyCode: "USD",
			HoursPerDay:  8,
			DaysPerWeek:  5,
			StartOn:      ptr.Time(app.ExampleTime()),
			DueOn:        ptr.Time(app.ExampleTime().Add(24 * time.Hour)),
			IsPublic:     false,
		},
		IsArchived: false,
	}
)

func OnSetSocials(tlbx app.Tlbx, user *user.User) error {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	_, err := tx.Exec(`UPDATE projectUsers SET handle=?, alias=?, hasAvatar=? WHERE id=?`, user.Handle, user.Alias, user.HasAvatar, user.ID)
	PanicOn(err)
	tx.Commit()
	return nil
}

func OnDelete(tlbx app.Tlbx, me ID) {
	srv := service.Get(tlbx)
	tx := srv.Data().Begin()
	defer tx.Rollback()
	_, err := tx.Exec(`DELETE FROM projectLocks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projectUsers WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projectActivities WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM projects WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM tasks WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM times WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM expenses WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM files WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`DELETE FROM comments WHERE host=?`, me)
	PanicOn(err)
	_, err = tx.Exec(`UPDATE projectUsers set isActive=FALSE WHERE id=?`, me)
	PanicOn(err)
	srv.Store().MustDeletePrefix(consts.FileBucket, me.String())
	tx.Commit()
}

func getSet(tlbx app.Tlbx, args *project.Get) *project.GetRes {
	validate.MaxIDs(tlbx, "ids", args.IDs, 100)
	app.BadReqIf(
		args.CreatedOnMin != nil &&
			args.CreatedOnMax != nil &&
			args.CreatedOnMin.After(*args.CreatedOnMax),
		"createdOnMin must be before createdOnMax")
	app.BadReqIf(
		args.StartOnMin != nil &&
			args.StartOnMax != nil &&
			args.StartOnMin.After(*args.StartOnMax),
		"startOnMin must be before startOnMax")
	app.BadReqIf(
		args.DueOnMin != nil &&
			args.DueOnMax != nil &&
			args.DueOnMin.After(*args.DueOnMax),
		"dueOnMin must be before dueOnMax")
	app.BadReqIf(
		args.StartOnMin != nil &&
			args.DueOnMin != nil &&
			args.StartOnMin.After(*args.DueOnMax),
		"startOnMin must be before dueOnMin")
	app.BadReqIf(
		args.StartOnMax != nil &&
			args.DueOnMax != nil &&
			args.StartOnMax.After(*args.DueOnMax),
		"startOnMax must be before dueOnMax")
	limit := sql.Limit(*args.Limit, 100)
	me := me.Get(tlbx)
	srv := service.Get(tlbx)
	res := &project.GetRes{
		Set: make([]*project.Project, 0, limit),
	}
	query := bytes.NewBufferString(`SELECT p.id, p.isArchived, p.name, p.createdOn, p.currencyCode, p.hoursPerDay, p.daysPerWeek, p.startOn, p.dueOn, p.isPublic, t.parent, t.firstChild, t.nextSibling, t.user, t.name, t.description, t.createdBy, t.createdOn, t.minimumRemainingTime, t.estimatedTime, t.loggedTime, t.estimatedSubTime, t.loggedSubTime, t.estimatedExpense, t.loggedExpense, t.estimatedSubExpense, t.loggedSubExpense, t.fileCount, t.fileSize, t.subFileCount, t.subFileSize, t.childCount, t.descendantCount, t.isParallel FROM projects p JOIN tasks t ON (t.host=p.host AND t.project=p.id AND t.id=p.id) WHERE p.host=?`)
	queryArgs := make([]interface{}, 0, 14)
	queryArgs = append(queryArgs, args.Host)
	idsLen := len(args.IDs)
	if !me.Equal(args.Host) {
		query.WriteString(` AND (isPublic=TRUE OR id IN (SELECT project FROM projectUsers WHERE host=? AND isActive=true AND id=?))`)
		queryArgs = append(queryArgs, args.Host, me)
	}
	if idsLen > 0 {
		query.WriteString(sql.InCondition(true, `id`, idsLen))
		query.WriteString(sql.OrderByField(`id`, idsLen))
		queryArgs = append(queryArgs, args.IDs.ToIs()...)
		queryArgs = append(queryArgs, args.IDs.ToIs()...)
	} else {
		if ptr.StringOr(args.NameStartsWith, "") != "" {
			query.WriteString(` AND name LIKE ?`)
			queryArgs = append(queryArgs, Sprintf(`%s%%`, *args.NameStartsWith))
		}
		query.WriteString(` AND isArchived=?`)
		queryArgs = append(queryArgs, args.IsArchived)
		if args.IsPublic != nil {
			query.WriteString(` AND isPublic=?`)
			queryArgs = append(queryArgs, *args.IsPublic)
		}
		if ptr.StringOr(args.NameStartsWith, "") != "" {
			query.WriteString(` AND name LIKE ?`)
			queryArgs = append(queryArgs, Sprintf(`%s%%`, *args.NameStartsWith))
		}
		if args.CreatedOnMin != nil {
			query.WriteString(` AND createdOn >= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMin)
		}
		if args.CreatedOnMax != nil {
			query.WriteString(` AND createdOn <= ?`)
			queryArgs = append(queryArgs, *args.CreatedOnMax)
		}
		if args.StartOnMin != nil {
			query.WriteString(` AND startOn >= ?`)
			queryArgs = append(queryArgs, *args.StartOnMin)
		}
		if args.StartOnMax != nil {
			query.WriteString(` AND startOn <= ?`)
			queryArgs = append(queryArgs, *args.StartOnMax)
		}
		if args.DueOnMin != nil {
			query.WriteString(` AND dueOn >= ?`)
			queryArgs = append(queryArgs, *args.DueOnMin)
		}
		if args.DueOnMax != nil {
			query.WriteString(` AND dueOn <= ?`)
			queryArgs = append(queryArgs, *args.DueOnMax)
		}
		if args.After != nil {
			query.WriteString(Sprintf(` AND %s %s= (SELECT %s FROM projects WHERE host=? AND id=?) AND id <> ?`, args.Sort, sql.GtLtSymbol(*args.Asc), args.Sort))
			queryArgs = append(queryArgs, args.Host, *args.After, *args.After)
			if args.Sort != consts.SortCreatedOn {
				query.WriteString(Sprintf(` AND createdOn %s (SELECT createdOn FROM projects WHERE host=? AND id=?)`, sql.GtLtSymbol(*args.Asc)))
				queryArgs = append(queryArgs, args.Host, *args.After)
			}
		}
		createdOnSecondarySort := ""
		if args.Sort != consts.SortCreatedOn {
			createdOnSecondarySort = ", createdOn"
		}
		query.WriteString(Sprintf(` ORDER BY %s%s %s, id LIMIT %d`, args.Sort, createdOnSecondarySort, sql.Asc(*args.Asc), limit))
	}
	srv.Data().Query(func(rows isql.Rows) {
		for rows.Next() {
			if len(args.IDs) == 0 && len(res.Set)+1 == limit {
				res.More = true
				break
			}
			p := &project.Project{}
			PanicOn(rows.Scan(&p.ID, &p.IsArchived, &p.Name, &p.CreatedOn, &p.CurrencyCode, &p.HoursPerDay, &p.DaysPerWeek, &p.StartOn, &p.DueOn, &p.IsPublic, &p.Parent, &p.FirstChild, &p.NextSibling, &p.User, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedOn, &p.MinimumRemainingTime, &p.EstimatedTime, &p.LoggedTime, &p.EstimatedSubTime, &p.LoggedSubTime, &p.EstimatedExpense, &p.LoggedExpense, &p.EstimatedSubExpense, &p.LoggedSubExpense, &p.FileCount, &p.FileSize, &p.SubFileCount, &p.SubFileSize, &p.ChildCount, &p.DescendantCount, &p.IsParallel))
			res.Set = append(res.Set, p)
		}
	}, query.String(), queryArgs...)
	return res
}
