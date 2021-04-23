package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	j "github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth"
	"golang.org/x/crypto/ssh/terminal"
)

//util program to generate and manipulate perfect k-ary trees of subtasks in tw projects
func main() {
	// cmdFlags
	fs := flag.NewFlagSet("twtrees", flag.ExitOnError)
	var outputCsvFile bool
	fs.BoolVar(&outputCsvFile, "o", false, "print out a csv file of results")
	var inst string
	fs.StringVar(&inst, "i", "http://sunbeam.teamwork.localhost", "installation url base for running tests on")
	var user string
	fs.StringVar(&user, "u", "test@test.test", "user email for basic auth")
	var treesHost string
	fs.StringVar(&treesHost, "th", "https://task-trees.com", "the url host of the task trees env")
	var treesUser string
	fs.StringVar(&treesUser, "tu", "test@test.test", "user email for task trees env")
	var treeK uint
	fs.UintVar(&treeK, "k", 3, "k-ary tree k value must be >0")
	var treeH uint
	fs.UintVar(&treeH, "h", 3, "k-ary tree h value")
	var projectName string
	fs.StringVar(&projectName, "pn", "twtrees", "project name to use in tw projects")
	fs.Parse(os.Args[1:])
	if treeK < 1 {
		panic("treeK value less than 1")
	}
	Print("Enter TW Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	PanicOn(err)
	pwd := string(bytePassword)
	Println()
	Print("Enter trees Password: ")
	bytePassword, err = terminal.ReadPassword(int(syscall.Stdin))
	PanicOn(err)
	treesPwd := string(bytePassword)
	Println()
	projectName += "_" + time.Now().Format("20060102150405")
	Println("outputCsvFile =", outputCsvFile)
	Println("i =", inst)
	Println("u =", user)
	Println("p = *******")
	Println("k =", treeK)
	Println("h =", treeH)
	var totalTasksToCreate uint
	if treeK == 1 {
		totalTasksToCreate = treeH
	} else {
		totalTasksToCreate = (pow(treeK, treeH+1) - 1) / (treeK - 1)
	}
	Println("N =", totalTasksToCreate)
	Println("pn =", projectName)

	runTW(inst, user, pwd, projectName, treeK, treeH)
	runTrees(treesHost, treesUser, treesPwd, projectName, treeK, treeH)
}

func runTW(inst, user, pwd, projectName string, treeK, treeH uint) {
	rm := &twReqMaker{
		inst: inst,
		user: user,
		pwd:  pwd,
	}

	Println("starting in TW")
	Println("get my id")
	myId := rm.get("/me.json", nil).MustInt64("person", "id")
	Println("myId =", myId)

	Println("create project")
	pj := j.MustFromString(`{"project":{}}`)
	pj.MustGet("project").MustSet("name", projectName).MustSet("people", int64Str(myId)).MustSet("use-tasks", true)
	projectId := rm.post("/projects.json", pj).MustInt64("id")
	Println("projectId =", projectId)

	Println("create tasklist")
	tasklistId := rm.post(Strf("/projects/%d/tasklists.json", projectId), j.MustNew().MustSet("todo-list", "name", "twtrees")).MustInt64("TASKLISTID")
	Println("tasklistId =", tasklistId)

	Println("create root task")
	pj = j.MustFromString(`{"todo-item":{}}`)
	pj.MustGet("todo-item").MustSet("content", "0").MustSet("estimated-minutes", 60).MustSet("start-date", todayDateString()).MustSet("due-date", tomorrowDateString())
	rootTaskId := rm.post(Strf("/tasklists/%d/tasks.json", tasklistId), pj).MustInt64("id")
	Println("rootTaskId =", rootTaskId)

	start := time.Now()
	twCreatePerfectKaryTree(rm, tasklistId, rootTaskId, 0, 0, treeK, treeH)
	Println()
	Println("time to create tree (excluding root node)", time.Now().Sub(start))

	start = time.Now()
	pj = j.MustFromString(`{"todo-item":{}}`)
	pj.MustGet("todo-item").MustSet("start-date", tomorrowDateString()).MustSet("due-date", dayAfterTomorrowDateString()).MustSet("push-subtasks", true).MustSet("push-dependents", true).MustSet("use-defaults", false)
	rm.put(Strf("/tasks/%d.json", rootTaskId), pj)
	Println("time to push start/due dates", time.Now().Sub(start))
	Println("finished in TW")
}

func runTrees(host, email, pwd, projectName string, treeK, treeH uint) {
	c := app.NewClient(host)
	me := (&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c)

	Println("starting in Trees")

	Println("create project")
	p := (&project.Create{
		Name: projectName,
	}).MustDo(c)
	Println("projectId =", p.ID.String())

	start := time.Now()
	treesCreatePerfectKaryTree(me, c, p.ID, p.ID, 0, 0, treeK, treeH)
	Println()
	Println("time to create tree (excluding root node)", time.Now().Sub(start))
	Println("finished in Trees")
}

func pow(x, y uint) uint {
	val := x
	for i := uint(0); i < y-1; i++ {
		x *= val
	}
	return x
}

func todayDateString() string {
	return time.Now().Format("20060102")
}

func tomorrowDateString() string {
	return time.Now().Add(time.Hour * 24).Format("20060102")
}

func dayAfterTomorrowDateString() string {
	return time.Now().Add(time.Hour * 48).Format("20060102")
}

type twReqMaker struct {
	inst      string
	user      string
	pwd       string
	projectId string
}

func (r *twReqMaker) do(method, path string, body *j.Json) *j.Json {
	var re io.Reader
	if body != nil {
		re = body.MustToReader()
	}
	req, e := http.NewRequest(method, r.inst+path, re)
	panicIf(e)
	req.SetBasicAuth(r.user, r.pwd)
	req.Header.Set("twProjectsVer", "twtrees")
	resp, e := http.DefaultClient.Do(req)
	panicIf(e)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	return j.MustFromReadCloser(resp.Body)
}

func (r *twReqMaker) post(path string, body *j.Json) *j.Json {
	return r.do("POST", path, body)
}

func (r *twReqMaker) put(path string, body *j.Json) *j.Json {
	return r.do("PUT", path, body)
}

func (r *twReqMaker) get(path string, body *j.Json) *j.Json {
	return r.do("GET", path, body)
}

func twCreatePerfectKaryTree(rm *twReqMaker, tasklistId, parentTaskId, lastUsedNameIdx int64, currentDepth, k, h uint) int64 {
	if currentDepth >= h {
		return lastUsedNameIdx
	}
	for i := uint(0); i < k; i++ {
		lastUsedNameIdx++
		Printf("\rcreating node %d", lastUsedNameIdx)
		pj := j.MustFromString(`{"todo-item":{}}`)
		pj.MustGet("todo-item").MustSet("content", int64Str(lastUsedNameIdx)).MustSet("estimated-minutes", 60).MustSet("start-date", todayDateString()).MustSet("due-date", tomorrowDateString()).MustSet("parentTaskId", parentTaskId)
		taskId := rm.post(Strf("/tasklists/%d/tasks.json", tasklistId), pj).MustInt64("id")
		lastUsedNameIdx = twCreatePerfectKaryTree(rm, tasklistId, taskId, lastUsedNameIdx, currentDepth+1, k, h)
	}
	return lastUsedNameIdx
}

func treesCreatePerfectKaryTree(me ID, c *app.Client, projectId, parentId ID, lastUsedNameIdx int64, currentDepth, k, h uint) int64 {
	if currentDepth >= h {
		return lastUsedNameIdx
	}
	var previousSiblingId *ID
	for i := uint(0); i < k; i++ {
		lastUsedNameIdx++
		Printf("\rcreating node %d", lastUsedNameIdx)
		isParallel := true
		est := uint64(60)
		if i%2 == 0 {
			isParallel = false
		}
		t := (&task.Create{
			Host:       me,
			Project:    projectId,
			Parent:     parentId,
			PrevSib:    previousSiblingId,
			Name:       int64Str(lastUsedNameIdx),
			IsParallel: isParallel,
			User:       &me,
			TimeEst:    est,
			CostEst:    est * 2,
		}).MustDo(c)
		previousSiblingId = &t.Task.ID
		lastUsedNameIdx = treesCreatePerfectKaryTree(me, c, projectId, t.Task.ID, lastUsedNameIdx, currentDepth+1, k, h)
	}
	return lastUsedNameIdx
}

func int64Str(i int64) string {
	return Strf("%d", i)
}

func panicIf(e error) {
	if e != nil {
		panic(e)
	}
}
