package test

import (
	"net"
	"net/http"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
	"github.com/0xor1/tlbx/pkg/web/server"
	"github.com/tomasen/realip"
)

const (
	baseHref    = "http://localhost:%d"
	pwd         = "1aA$_t;3"
	emailSuffix = "@test.localhost"
)

type Rig interface {
	// http
	Port() int
	NewClient() *app.Client
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
	Store() store.Client
	// cleanup
	CleanUp()
}

type User interface {
	Client() *app.Client
	ID() ID
	Email() string
	Pwd() string
}

type testUser struct {
	client *app.Client
	id     ID
	email  string
	pwd    string
}

func (u *testUser) Client() *app.Client {
	return u.client
}

func (u *testUser) ID() ID {
	return u.id
}

func (u *testUser) Email() string {
	return u.email
}

func (u *testUser) Pwd() string {
	return u.pwd
}

type rig struct {
	port    int
	ali     *testUser
	bob     *testUser
	cat     *testUser
	dan     *testUser
	log     log.Log
	cache   iredis.Pool
	user    isql.ReplicaSet
	pwd     isql.ReplicaSet
	data    isql.ReplicaSet
	email   email.Client
	store   store.Client
	useAuth bool
}

func (r *rig) Port() int {
	return r.port
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

func (r *rig) Store() store.Client {
	return r.store
}

func (r *rig) NewClient() *app.Client {
	return app.NewClient(Sprintf(baseHref, r.Port()), &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 200}})
}

func NewRig(
	config *config.Config,
	eps []*app.Endpoint,
	useUsers bool,
	onActivate func(app.Tlbx, *user.User),
	onDelete func(app.Tlbx, ID),
	enableAliases bool,
	onSetAlias func(app.Tlbx, ID, *string) error,
	enableAvatars bool,
	onSetAvatar func(app.Tlbx, ID, bool) error,
) Rig {
	listener, err := net.Listen("tcp", ":0")
	PanicOn(err)
	r := &rig{
		port:    listener.Addr().(*net.TCPAddr).Port,
		log:     config.Log,
		cache:   config.Cache,
		email:   config.Email,
		store:   config.Store,
		user:    config.User,
		pwd:     config.Pwd,
		data:    config.Data,
		useAuth: useUsers,
	}

	if useUsers {
		eps = append(
			eps,
			usereps.New(
				config.FromEmail,
				config.ActivateFmtLink,
				config.ConfirmChangeEmailFmtLink,
				onActivate,
				onDelete,
				enableAliases,
				onSetAlias,
				enableAvatars,
				config.AvatarBucket,
				config.AvatarPrefix,
				onSetAvatar,
				config.Store)...)
	}
	go app.Run(func(c *app.Config) {
		c.TlbxMwares = app.TlbxMwares{
			session.BasicMware(config.SessionAuthKey64s, config.SessionEncrKey32s, config.IsLocal),
			ratelimit.Mware(func(c *ratelimit.Config) {
				c.KeyGen = func(tlbx app.Tlbx) string {
					var key string
					if me.Exists(tlbx) {
						key = me.Get(tlbx).String()
					}
					return Sprintf("rate-limiter-%s-%s", realip.RealIP(tlbx.Req()), key)
				}
				c.Pool = r.cache
				c.PerMinute = 1000000 // when running batch tests 120 rate limit is easily exceeded
			}),
			service.Mware(r.cache, r.user, r.pwd, r.data, r.email, r.store),
		}
		c.Endpoints = eps
		c.Serve = func(h http.HandlerFunc) {
			server.Run(func(servC *server.Config) {
				servC.AppListener = listener
				servC.Log = c.Log
				servC.Handler = h
			})
		}
	})

	time.Sleep(20 * time.Millisecond)
	r.ali = r.createUser("ali", emailSuffix, pwd)
	r.bob = r.createUser("bob", emailSuffix, pwd)
	r.cat = r.createUser("cat", emailSuffix, pwd)
	r.dan = r.createUser("dan", emailSuffix, pwd)
	return r
}

func (r *rig) CleanUp() {
	if r.useAuth {
		(&user.Delete{
			Pwd: r.Ali().Pwd(),
		}).MustDo(r.Ali().Client())
		(&user.Delete{
			Pwd: r.Bob().Pwd(),
		}).MustDo(r.Bob().Client())
		(&user.Delete{
			Pwd: r.Cat().Pwd(),
		}).MustDo(r.Cat().Client())
		(&user.Delete{
			Pwd: r.Dan().Pwd(),
		}).MustDo(r.Dan().Client())
	}
}

func (r *rig) createUser(alias, emailSuffix, pwd string) *testUser {
	email := Sprintf("%s%s%d", alias, emailSuffix, r.port)
	c := r.NewClient()
	if r.useAuth {
		(&user.Register{
			Alias:      ptr.String(alias),
			Email:      email,
			Pwd:        pwd,
			ConfirmPwd: pwd,
		}).MustDo(c)

		var code string
		row := r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
		PanicOn(row.Scan(&code))

		(&user.Activate{
			Email: email,
			Code:  code,
		}).MustDo(c)

		id := (&user.Login{
			Email: email,
			Pwd:   pwd,
		}).MustDo(c).ID

		return &testUser{
			client: c,
			id:     id,
			email:  email,
			pwd:    pwd,
		}
	}
	return &testUser{client: c}
}