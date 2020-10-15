package taskeps

import (
	"net/http"
	"strings"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/epsutil"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new task",
			Path:         (&task.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Create{
					Host:            app.ExampleID(),
					Project:         app.ExampleID(),
					Parent:          app.ExampleID(),
					PreviousSibling: ptr.ID(app.ExampleID()),
					Name:            "do it",
					Description:     ptr.String("do the thing you're supposed to do"),
					IsParallel:      true,
					User:            ptr.ID(app.ExampleID()),
					EstimatedTime:   40,
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleTask
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Create)
				me := me.Get(tlbx)
				validate.Str("name", args.Name, tlbx, nameMinLen, nameMaxLen)
				if args.Description != nil && *args.Description == "" {
					args.Description = nil
				}
				if args.Description != nil {
					validate.Str("description", *args.Description, tlbx, descriptionMinLen, descriptionMaxLen)
				}
				epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				t := &task.Task{
					ID:                  tlbx.NewID(),
					Parent:              &args.Parent,
					FirstChild:          nil,
					NextSibling:         nil,
					User:                args.User,
					Name:                args.Name,
					Description:         args.Description,
					CreatedBy:           me,
					CreatedOn:           NowMilli(),
					MinimumTime:         args.EstimatedTime,
					EstimatedTime:       args.EstimatedTime,
					LoggedTime:          0,
					EstimatedSubTime:    0,
					LoggedSubTime:       0,
					EstimatedExpense:    args.EstimatedExpense,
					LoggedExpense:       0,
					EstimatedSubExpense: 0,
					LoggedSubExpense:    0,
					FileCount:           0,
					FileSize:            0,
					FileSubCount:        0,
					FileSubSize:         0,
					ChildCount:          0,
					DescendantCount:     0,
					IsParallel:          args.IsParallel,
				}
				if args.User != nil && !args.User.Equal(me) {
					// if Im assigning to someone that isnt me,
					// validate that user has write access to this
					// project
					epsutil.MustHaveAccess(tlbx, args.Host, args.Project, args.User, cnsts.RoleWriter)
				}
				srv := service.Get(tlbx)
				tx := srv.Data().Begin()
				defer tx.Rollback()
				// lock project, required for any action that will change aggregate values nad/or tree structure
				epsutil.MustLockProject(tlbx, tx, args.Host, args.Project)
				// get correct next sibling value from either previousSibling if
				// specified or parent.FirstChild otherwise. Then update previousSiblings nextSibling value
				// or parents firstChild value depending on the scenario.
				var previousSibling *task.Task
				if args.PreviousSibling != nil {
					previousSibling = getOne(tlbx, tx, args.Host, args.Project, *args.PreviousSibling)
					app.ReturnIf(previousSibling == nil, http.StatusNotFound, "previousSibling not found")
					t.NextSibling = previousSibling.NextSibling
					previousSibling.NextSibling = &t.ID
					// point previous sibling at new task
					_, err := tx.Exec(`UPDATE tasks SET nextSibling=? WHERE host=? AND project=? AND id=?`, t.ID, args.Host, args.Project, previousSibling.ID)
					PanicOn(err)
				} else {
					// else newTask is being inserted as firstChild, so set any current firstChild
					// as newTask's NextSibling
					// get parent for updating child/descendant counts and firstChild if required
					parent := getOne(tlbx, tx, args.Host, args.Project, args.Parent)
					app.ReturnIf(parent == nil, http.StatusNotFound, "parent not found")
					t.NextSibling = parent.FirstChild
					// increment parents child and descendant counters and firstChild pointer incase that was changed
					_, err := tx.Exec(`UPDATE tasks SET firstChild=? WHERE host=? AND project=? AND id=?`, t.ID, args.Host, args.Project, t.Parent)
					PanicOn(err)
				}
				// insert new task
				_, err := tx.Exec(Strf(`INSERT INTO tasks (host, project, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, sql_task_columns), args.Host, args.Project, t.ID, t.Parent, t.FirstChild, t.NextSibling, t.User, t.Name, t.Description, t.CreatedBy, t.CreatedOn, t.MinimumTime, t.EstimatedTime, t.LoggedTime, t.EstimatedSubTime, t.LoggedSubTime, t.EstimatedExpense, t.LoggedExpense, t.EstimatedSubExpense, t.LoggedSubExpense, t.FileCount, t.FileSize, t.FileSubCount, t.FileSubSize, t.ChildCount, t.DescendantCount, t.IsParallel)
				PanicOn(err)
				// at this point the tree structure has been updated so all tasks are pointing to the correct new positions
				// all that remains to do is update aggregate values
				setAncestralChainAggregateValuesFromTask(tlbx, tx, args.Host, args.Project, args.Parent)
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Update a task",
			Path:         (&task.Update{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Update{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Update{
					Host:             app.ExampleID(),
					Project:          app.ExampleID(),
					ID:               app.ExampleID(),
					Parent:           &field.ID{V: app.ExampleID()},
					PreviousSibling:  &field.IDPtr{V: ptr.ID(app.ExampleID())},
					Name:             &field.String{V: "new name"},
					Description:      &field.StringPtr{V: ptr.String("new description")},
					IsParallel:       &field.Bool{V: true},
					User:             &field.IDPtr{V: ptr.ID(app.ExampleID())},
					EstimatedTime:    &field.UInt64{V: 123},
					EstimatedExpense: &field.UInt64{V: 123},
				}
			},
			GetExampleResponse: func() interface{} {
				return exampleTask
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Update)
				me := me.Get(tlbx)
				if args.Parent == nil &&
					args.PreviousSibling == nil &&
					args.Name == nil &&
					args.Description == nil &&
					args.IsParallel == nil &&
					args.User == nil &&
					args.EstimatedTime == nil &&
					args.EstimatedExpense == nil {
					// nothing to update
					return nil
				}
				if args.ID.Equal(args.Project) {
					app.ReturnIf(!me.Equal(args.Host), http.StatusForbidden, "only the host may edit the project root node")
					app.ReturnIf(args.User != nil, http.StatusForbidden, "user value is not settable on the project root node")
					app.ReturnIf(args.Parent != nil, http.StatusForbidden, "parent value is not settable on the project root node")
					app.ReturnIf(args.PreviousSibling != nil, http.StatusForbidden, "previousSibling value is not settable on the project root node")
				} else {
					epsutil.IMustHaveAccess(tlbx, args.Host, args.Project, cnsts.RoleWriter)
				}
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				treeUpdateRequired := false
				simpleUpdateRequired := false
				if args.Parent != nil ||
					args.PreviousSibling != nil ||
					args.EstimatedExpense != nil ||
					args.EstimatedTime != nil ||
					args.IsParallel != nil {
					// if moving the task or setting an aggregate value effecting property
					// we must lock
					epsutil.MustLockProject(tlbx, tx, args.Host, args.Project)
				}
				t := getOne(tlbx, tx, args.Host, args.Project, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "task not found")
				var newParent, newPreviousSibling, currentParent, currentPreviousSibling *task.Task
				if args.Parent != nil {
					if !args.Parent.V.Equal(*t.Parent) {
						var newNextSibling *ID
						// validate new parent exists
						newParent = getOne(tlbx, tx, args.Host, args.Project, args.Parent.V)
						app.ReturnIf(newParent == nil, http.StatusNotFound, "parent not found")
						// we must ensure we dont allow recursive loops
						row := tx.QueryRow(Strf("%s SELECT COUNT(*)=1 FROM ancestors a WHERE id=?", sql_ancestors_cte), args.Host, args.Project, args.Parent.V, args.Host, args.Project, args.ID)
						ancestorLoopDetected := false
						PanicOn(row.Scan(&ancestorLoopDetected))
						app.BadReqIf(ancestorLoopDetected, "ancestor loop detected, invalid parent value")
						if args.PreviousSibling != nil && args.PreviousSibling.V != nil {
							newPreviousSibling = getOne(tlbx, tx, args.Host, args.Project, *args.PreviousSibling.V)
							app.ReturnIf(newPreviousSibling == nil, http.StatusNotFound, "previousSibling not found")
							app.BadReqIf(!newPreviousSibling.Parent.Equal(args.Parent.V), "previousSiblings parent does not match the specified parent arg")
							newNextSibling = newPreviousSibling.NextSibling
							newPreviousSibling.NextSibling = &t.ID
						} else {
							newNextSibling = newParent.FirstChild
							newParent.FirstChild = &t.ID
						}
						// need to reconnect currentPreviousSibling with current nextSibling
						currentPreviousSibling = getPreviousSibling(tlbx, tx, args.Host, args.Project, args.ID)
						// need to get current parent for ancestor value updates
						currentParent = getOne(tlbx, tx, args.Host, args.Project, *t.Parent)
						if currentPreviousSibling != nil {
							currentPreviousSibling.NextSibling = t.NextSibling
						} else {
							// need to update currentParent firstChild as t is it
							currentParent.FirstChild = t.NextSibling
						}
						t.Parent = &newParent.ID
						t.NextSibling = newNextSibling
						treeUpdateRequired = true
					} else {
						// if args.Parent is set and it's equal to the current value
						// i.e. no change is being made then let's just nil it out
						args.Parent = nil
					}
				}
				if args.Parent == nil && args.PreviousSibling != nil {
					// we now know we are doing a purely horizontal move, i.e. not changing parent node
					// get current previousSibling
					currentPreviousSibling = getPreviousSibling(tlbx, tx, args.Host, args.Project, args.ID)
					if !((currentPreviousSibling == nil && args.PreviousSibling.V == nil) ||
						(currentPreviousSibling != nil && args.PreviousSibling.V != nil &&
							currentPreviousSibling.ID.Equal(*args.PreviousSibling.V))) {
						var newNextSibling *ID
						// here we know that an actual change is being attempted
						if args.PreviousSibling.V != nil {
							// moving to a non first child position
							newPreviousSibling = getOne(tlbx, tx, args.Host, args.Project, *args.PreviousSibling.V)
							app.ReturnIf(newPreviousSibling == nil, http.StatusNotFound, "previousSibling not found")
							app.BadReqIf(!newPreviousSibling.Parent.Equal(*t.Parent), "previousSiblings parent does not match the current tasks parent")
							newNextSibling = newPreviousSibling.NextSibling
							newPreviousSibling.NextSibling = &t.ID
						} else {
							// moving to the first child position
							currentParent = getOne(tlbx, tx, args.Host, args.Project, *t.Parent)
							PanicIf(currentParent == nil, "currentParent not found")
							newNextSibling = currentParent.FirstChild
							currentParent.FirstChild = &t.ID
						}
						// need to reconnect currentPreviousSibling with current nextSibling
						if currentPreviousSibling != nil {
							currentPreviousSibling.NextSibling = t.NextSibling
						}
						t.NextSibling = newNextSibling
						treeUpdateRequired = true
					} else {
						// here we know no change is being made so nil out currentPreviousSibling to save a sql update call
						currentPreviousSibling = nil
						// and nil this out to remove redundant data from activity entry
						args.PreviousSibling = nil
					}
				}
				// at this point all the moving has been done
				if args.Name != nil && t.Name != args.Name.V {
					validate.Str("name", args.Name.V, tlbx, nameMinLen, nameMaxLen)
					t.Name = args.Name.V
					simpleUpdateRequired = true
				}
				if args.Description != nil && args.Description.V != nil && *args.Description.V == "" {
					args.Description.V = nil
				}
				if args.Description != nil &&
					((args.Description.V == nil && t.Description != nil) ||
						(args.Description.V != nil && t.Description == nil) ||
						(args.Description.V != nil && t.Description != nil && t.Description != args.Description.V)) {
					if args.Description.V != nil {
						validate.Str("description", *args.Description.V, tlbx, descriptionMinLen, descriptionMaxLen)
					}
					t.Description = args.Description.V
					simpleUpdateRequired = true
				}
				if args.IsParallel != nil && t.IsParallel != args.IsParallel.V {
					t.IsParallel = args.IsParallel.V
					treeUpdateRequired = true
				}
				if args.User != nil &&
					((args.User.V == nil && t.User != nil) ||
						(args.User.V != nil && t.User == nil) ||
						(args.User.V != nil && t.User != nil && !t.User.Equal(*args.User.V))) {
					if args.User.V != nil && !args.User.V.Equal(me) {
						epsutil.MustHaveAccess(tlbx, args.Host, args.Project, args.User.V, cnsts.RoleWriter)
					}
					t.User = args.User.V
					simpleUpdateRequired = true
				}
				if args.EstimatedTime != nil && t.EstimatedTime != args.EstimatedTime.V {
					t.EstimatedTime = args.EstimatedTime.V
					treeUpdateRequired = true
				}
				if args.EstimatedExpense != nil && t.EstimatedExpense != args.EstimatedExpense.V {
					t.EstimatedExpense = args.EstimatedExpense.V
					treeUpdateRequired = true
				}
				update := func(ts ...*task.Task) {
					for _, t := range ts {
						if t != nil {
							_, err := tx.Exec(`UPDATE tasks SET parent=?, firstChild=?, nextSibling=?, name=?, description=?, isParallel=?, user=?, estimatedTime=?, estimatedExpense=? WHERE host=? AND project=? AND id=?`, t.Parent, t.FirstChild, t.NextSibling, t.Name, t.Description, t.IsParallel, t.User, t.EstimatedTime, t.EstimatedExpense, args.Host, args.Project, t.ID)
							PanicOn(err)
						}
					}
				}
				if simpleUpdateRequired || treeUpdateRequired {
					update(t, currentParent, currentPreviousSibling, newParent, newPreviousSibling)
					epsutil.LogActivity(tlbx, tx, args.Host, args.Project, args.ID, cnsts.TypeTask, cnsts.ActionUpdated, &t.Name, args)
				}
				if treeUpdateRequired {
					if currentParent != nil {
						// if we moved parent we must recalculate aggregate values on the old parents ancestral chain
						setAncestralChainAggregateValuesFromTask(tlbx, tx, args.Host, args.Project, currentParent.ID)
					}
					setAncestralChainAggregateValuesFromTask(tlbx, tx, args.Host, args.Project, t.ID)
				}
				tx.Commit()
				return t
			},
		},
		{
			Description:  "Delete tasks",
			Path:         (&task.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &task.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &task.Delete{
					Host:    app.ExampleID(),
					Project: app.ExampleID(),
					ID:      app.ExampleID(),
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*task.Delete)
				app.BadReqIf(args.ID.Equal(args.Project), "use project delete endpoint to delete a project node")
				me := me.Get(tlbx)
				tx := service.Get(tlbx).Data().Begin()
				defer tx.Rollback()
				role := epsutil.MustGetRole(tlbx, tx, args.Host, args.Project, me)
				app.ReturnIf(role == cnsts.RoleReader, http.StatusForbidden, "you don't have permission to delete a task")
				epsutil.MustLockProject(tlbx, tx, args.Host, args.Project)
				// at this point we need to get the tasks
				t := getOne(tlbx, tx, args.Host, args.Project, args.ID)
				app.ReturnIf(t == nil, http.StatusNotFound, "task not found")
				app.ReturnIf(role == cnsts.RoleWriter && (t.DescendantCount > 5 || t.CreatedOn.Before(Now().Add(-1*time.Hour))), http.StatusForbidden, "an admin must delete this task")
				// need to check for mariadb 10.6 to see if CTEs work in update/delete queries so we can skip this intermediate step altogether
				queryArgs := make([]interface{}, 0, 3+t.DescendantCount)
				queryArgs = append(queryArgs, args.Host, args.Project)
				PanicOn(tx.Query(func(rows isql.Rows) {
					i := ID{}
					PanicOn(rows.Scan(&i))
					queryArgs = append(queryArgs, i)
				}, `WITH RECURSIVE descendants (id) AS (SELECT id FROM tasks WHERE host=? AND project=? AND id=? UNION SELECT t.id FROM tasks t, descendants d WHERE t.host=? AND t.project=? AND t.parent=d.id) CYCLE id restrict SELECT id FROM descendants`, args.Host, args.Project, args.ID, args.Host, args.Project))
				_, err := tx.Exec(Strf(`DELETE FROM tasks WHERE host=? AND project=? %s`, sql.InCondition(true, `id`, len(queryArgs)-2)), queryArgs...)
				PanicOn(err)
				tx.Commit()
				return nil
			},
		},
	}

	nameMinLen        = 1
	nameMaxLen        = 250
	descriptionMinLen = 1
	descriptionMaxLen = 1250
	exampleTask       = &task.Task{
		ID:                  app.ExampleID(),
		Parent:              ptr.ID(app.ExampleID()),
		FirstChild:          ptr.ID(app.ExampleID()),
		NextSibling:         ptr.ID(app.ExampleID()),
		User:                ptr.ID(app.ExampleID()),
		Name:                "do it",
		Description:         ptr.String("do that thing you're supposed to do"),
		CreatedBy:           app.ExampleID(),
		CreatedOn:           app.ExampleTime(),
		MinimumTime:         100,
		EstimatedTime:       100,
		LoggedTime:          100,
		EstimatedSubTime:    100,
		LoggedSubTime:       100,
		EstimatedExpense:    100,
		LoggedExpense:       100,
		EstimatedSubExpense: 100,
		LoggedSubExpense:    100,
		FileCount:           100,
		FileSize:            100,
		FileSubCount:        100,
		FileSubSize:         100,
		ChildCount:          100,
		DescendantCount:     100,
		IsParallel:          true,
	}
)

func setAncestralChainAggregateValuesFromTask(tlbx app.Tlbx, tx service.Tx, host, project, task ID) {
	ancestorChain := make(IDs, 0, 20)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			i := ID{}
			PanicOn(rows.Scan(&i))
			ancestorChain = append(ancestorChain, i)
		}
	}, `CALL setAncestralChainAggregateValuesFromTask(?, ?, ?)`, host, project, task))
}

func getAncestors(tlbx app.Tlbx, tx service.Tx, host, project, ofTask ID, after *ID, limit int) []*task.Task {
	ancestors := make([]*task.Task, 0, 20)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			t, err := scan(rows)
			PanicOn(err)
			ancestors = append(ancestors, t)
		}
	}, Strf(`%s SELECT %s FROM tasks t JOIN ancestors a ON t.id = a.id WHERE t.host=? AND t.project=? AND a.n <> 0 ORDER BY a.n ASC`, sql_ancestors_cte, sql_task_columns_prefixed), host, project, ofTask, host, project, host, project))
	return ancestors
}

func getOne(tlbx app.Tlbx, tx service.Tx, host, project, id ID) *task.Task {
	ts := get(tlbx, tx, host, project, IDs{id})
	if len(ts) == 0 {
		return nil
	}
	return ts[0]
}

func get(tlbx app.Tlbx, tx service.Tx, host, project ID, ids IDs) []*task.Task {
	if len(ids) == 0 {
		return nil
	}
	ts := make([]*task.Task, 0, len(ids))
	queryArgs := make([]interface{}, 0, len(ids)+2)
	queryArgs = append(queryArgs, host, project)
	queryArgs = append(queryArgs, ids.ToIs()...)
	PanicOn(tx.Query(func(rows isql.Rows) {
		for rows.Next() {
			t, err := scan(rows)
			PanicOn(err)
			ts = append(ts, t)
		}
	}, Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.project=? %s`, sql_task_columns_prefixed, sql.InCondition(true, `t.id`, len(ids))), queryArgs...))
	return ts
}

func getPreviousSibling(tlbx app.Tlbx, tx service.Tx, host, project, nextSibling ID) *task.Task {
	row := tx.QueryRow(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.project=? AND t.nextSibling=?`, sql_task_columns_prefixed), host, project, nextSibling)
	t, err := scan(row)
	sql.PanicIfIsntNoRows(err)
	return t
}

type taskScanner interface {
	Scan(dest ...interface{}) error
}

func scan(ts taskScanner) (*task.Task, error) {
	t := &task.Task{}
	err := ts.Scan(
		&t.ID,
		&t.Parent,
		&t.FirstChild,
		&t.NextSibling,
		&t.User,
		&t.Name,
		&t.Description,
		&t.CreatedBy,
		&t.CreatedOn,
		&t.MinimumTime,
		&t.EstimatedTime,
		&t.LoggedTime,
		&t.EstimatedSubTime,
		&t.LoggedSubTime,
		&t.EstimatedExpense,
		&t.LoggedExpense,
		&t.EstimatedSubExpense,
		&t.LoggedSubExpense,
		&t.FileCount,
		&t.FileSize,
		&t.FileSubCount,
		&t.FileSubSize,
		&t.ChildCount,
		&t.DescendantCount,
		&t.IsParallel)
	if t.ID.IsZero() {
		t = nil
	}
	return t, err
}

var (
	sql_task_columns_prefixed = `t.id, t.parent, t.firstChild, t.nextSibling, t.user, t.name, t.description, t.createdBy, t.createdOn, t.minimumTime, t.estimatedTime, t.loggedTime, t.estimatedSubTime, t.loggedSubTime, t.estimatedExpense, t.loggedExpense, t.estimatedSubExpense, t.loggedSubExpense, t.fileCount, t.fileSize, t.fileSubCount, t.fileSubSize, t.childCount, t.descendantCount, t.isParallel`
	sql_task_columns          = strings.ReplaceAll(sql_task_columns_prefixed, `t.`, ``)
	sql_ancestors_cte         = `WITH RECURSIVE ancestors (n, id, parent) AS (SELECT 0, id, parent FROM tasks WHERE host=? AND project=? AND id=? UNION SELECT a.n + 1, t.id, t.parent FROM tasks t, ancestors a WHERE t.host=? AND t.project=? AND t.id = a.parent) CYCLE id RESTRICT`
)
