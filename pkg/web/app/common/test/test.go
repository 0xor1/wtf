package test

import (
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth"
	"github.com/0xor1/wtf/pkg/web/app/common/auth/autheps"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

const (
	baseHref    = "http://localhost:8080"
	pwd         = "1aA$_t;3"
	emailSuffix = "@test.localhost"
)

type Rig interface {
	// log
	Log() log.Log
	// users
	Ali() User
	Bob() User
	Cat() User
	Dan() User
	// services
	Cache() iredis.Pool
	User() isql.ReplicaSet
	Pwd() isql.ReplicaSet
	Data() isql.ReplicaSet
	Email() email.Client
	Store() store.LocalClient
	// cleanup
	CleanUp()
}

type User interface {
	Client() *app.Client
	ID() ID
	Email() string
	Pwd() string
}

type user struct {
	client *app.Client
	id     ID
	email  string
	pwd    string
}

func (u *user) Client() *app.Client {
	return u.client
}

func (u *user) ID() ID {
	return u.id
}

func (u *user) Email() string {
	return u.email
}

func (u *user) Pwd() string {
	return u.pwd
}

type rig struct {
	ali   *user
	bob   *user
	cat   *user
	dan   *user
	log   log.Log
	cache iredis.Pool
	user  isql.ReplicaSet
	pwd   isql.ReplicaSet
	data  isql.ReplicaSet
	email email.Client
	store store.LocalClient
}

func (r *rig) Log() log.Log {
	return r.log
}

func (r *rig) Ali() User {
	return r.ali
}

func (r *rig) Bob() User {
	return r.bob
}

func (r *rig) Cat() User {
	return r.cat
}

func (r *rig) Dan() User {
	return r.dan
}

func (r *rig) Cache() iredis.Pool {
	return r.cache
}

func (r *rig) User() isql.ReplicaSet {
	return r.user
}

func (r *rig) Pwd() isql.ReplicaSet {
	return r.pwd
}

func (r *rig) Data() isql.ReplicaSet {
	return r.data
}

func (r *rig) Email() email.Client {
	return r.email
}

func (r *rig) Store() store.LocalClient {
	return r.store
}

func NewClient() *app.Client {
	return app.NewClient(baseHref)
}

func NewRig(eps []*app.Endpoint, onDelete func(app.Toolbox, ID)) Rig {
	l := log.New()
	r := &rig{
		log:   l,
		cache: iredis.CreatePool("localhost:6379"),
		email: email.NewLocalClient(l),
		store: store.NewLocalClient("tmpTestStore"),
	}

	var err error
	r.user, err = isql.NewReplicaSet("users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true&loc=UTC&multiStatements=true")
	PanicOn(err)
	r.pwd, err = isql.NewReplicaSet("pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	PanicOn(err)
	r.data, err = isql.NewReplicaSet("data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true&loc=UTC&multiStatements=true")
	PanicOn(err)

	go app.Run(func(c *app.Config) {
		c.ToolboxMware = service.Mware(r.cache, r.user, r.pwd, r.data, r.email, r.store)
		c.RateLimiterPool = r.cache
		c.Endpoints = append(eps, autheps.New(onDelete, "test@test.localhost", "http://localhost:8080")...)
	})

	time.Sleep(20 * time.Millisecond)
	r.ali = r.createUser("ali"+emailSuffix, pwd)
	r.bob = r.createUser("bob"+emailSuffix, pwd)
	r.cat = r.createUser("cat"+emailSuffix, pwd)
	r.dan = r.createUser("dan"+emailSuffix, pwd)

	return r
}

func (r *rig) CleanUp() {
	r.Store().MustDeleteStore()
	(&auth.Delete{
		Pwd: r.Ali().Pwd(),
	}).MustDo(r.Ali().Client())
	(&auth.Delete{
		Pwd: r.Bob().Pwd(),
	}).MustDo(r.Bob().Client())
	(&auth.Delete{
		Pwd: r.Cat().Pwd(),
	}).MustDo(r.Cat().Client())
	(&auth.Delete{
		Pwd: r.Dan().Pwd(),
	}).MustDo(r.Dan().Client())
}

func (r *rig) createUser(email, pwd string) *user {
	c := NewClient()

	(&auth.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).MustDo(c)

	var code string
	row := r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&auth.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	id := (&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c).Me

	return &user{
		client: c,
		id:     id,
		email:  email,
		pwd:    pwd,
	}
}