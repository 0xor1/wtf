package testutil

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app/test"
)

// only suitable for small test trees for visual validation
// whilst writing/debugging unit tests
func GrabFullTree(r test.Rig, host, project ID) map[string]*task.Task {
	tree := map[string]*task.Task{}
	rows, err := r.Data().Primary().Query(Strf(`SELECT %s FROM tasks t WHERE t.host=? AND t.Project=?`, taskeps.Sql_task_columns_prefixed), host, project)
	if rows != nil {
		defer rows.Close()
	}
	PanicOn(err)
	for rows.Next() {
		t, err := taskeps.Scan(rows)
		PanicOn(err)
		tree[t.ID.String()] = t
	}
	return tree
}

func PrintFullTree(project ID, tree map[string]*task.Task) {
	var print func(t *task.Task, as []*task.Task)
	print = func(t *task.Task, as []*task.Task) {
		p := 0
		if t.IsParallel {
			p = 1
		}
		v := Strf(`[n: %s, p: %d, m: %d, et: %d, est: %d, lt: %d, lst: %d, ee: %d, ese: %d, le: %d, lse: %d]`, t.Name, p, t.MinimumTime, t.EstimatedTime, t.EstimatedSubTime, t.LoggedTime, t.LoggedSubTime, t.EstimatedExpense, t.EstimatedSubExpense, t.LoggedExpense, t.LoggedSubExpense)
		if len(as) > 0 {
			pre := ``
			for _, a := range as[1:] {
				if a.NextSibling != nil {
					pre += `|    `
				} else {
					pre += `     `
				}
			}
			Println(Strf(`%s|`, pre))
			Println(Strf(`%s|`, pre))
			Println(Strf(`%s|____%s`, pre, v))
		} else {
			Println(v)
		}
		if t.FirstChild != nil {
			print(tree[t.FirstChild.String()], append(as, t))
		}
		if t.NextSibling != nil {
			print(tree[t.NextSibling.String()], as)
		}
	}
	println("n: name")
	println("p: isParallel")
	println("m: minimumTime")
	println("et: estimatedTime")
	println("est: estimatedSubTime")
	println("lt: loggedTime")
	println("lst: loggedSubTime")
	println("ee: estimatedExpense")
	println("ese: estimatedSubExpense")
	println("le: loggedExpense")
	println("lse: loggedSubExpense")
	println()
	print(tree[project.String()], nil)
}