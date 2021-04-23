package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth"
	"github.com/0xor1/tlbx/pkg/web/app/user/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/user/social"
	"github.com/0xor1/tlbx/pkg/web/app/user/social/socialeps"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
)

const (
	baseHref    = "http://localhost"
	pwd         = "1aA$_t;3"
	emailSuffix = "@test.localhost"
)

type Rig interface {
	// root http server handler
	RootHandler() http.HandlerFunc
	// unique
	Unique() string
	// http
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
	rootHandler http.HandlerFunc
	unique      string
	register    func(Rig, string, *auth.Register)
	ali         *testUser
	bob         *testUser
	cat         *testUser
	dan         *testUser
	log         log.Log
	rateLimit   iredis.Pool
	cache       iredis.Pool
	user        isql.ReplicaSet
	data        isql.ReplicaSet
	email       email.Client
	store       store.Client
	fcm         fcm.Client
	useAuth     bool
}

func (r *rig) RootHandler() http.HandlerFunc {
	return r.rootHandler
}

func (r *rig) Unique() string {
	return r.unique
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

func (r *rig) RateLimit() iredis.Pool {
	return r.rateLimit
}

func (r *rig) Cache() iredis.Pool {
	return r.cache
}

func (r *rig) User() isql.ReplicaSet {
	return r.user
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

func (r *rig) FCM() fcm.Client {
	return r.fcm
}

func (r *rig) NewClient() *app.Client {
	return app.NewClient(baseHref, r)
}

func (r *rig) Do(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	r.rootHandler(rec, req)
	return rec.Result(), nil
}

func NewUserRig(
	config *config.Config,
	eps []*app.Endpoint,
	rateLimitMware func(iredis.Pool, ...int) func(app.Tlbx),
	buckets []string,
	useSocial bool,
	register func(Rig, string, *auth.Register),
	configUser ...func(*usereps.Config),
) Rig {
	userEps, configAuth := usereps.AuthExtensions(configUser...)
	eps = app.JoinEps(eps, userEps)
	if useSocial {
		buckets = append(buckets, socialeps.AvatarBucket)
		if register == nil {
			register = func(r Rig, name string, reg *auth.Register) {
				reg.AppData = &social.RegisterAppData{
					Handle: name + r.Unique(),
					Alias:  name,
				}
			}
		}
	}
	return NewRig(
		config,
		app.JoinEps(eps, userEps),
		rateLimitMware,
		buckets,
		true,
		register,
		configAuth,
	)
}

func NewRig(
	config *config.Config,
	eps []*app.Endpoint,
	rateLimitMware func(iredis.Pool, ...int) func(app.Tlbx),
	buckets []string,
	useAuth bool,
	register func(Rig, string, *auth.Register),
	configAuth ...func(*autheps.Config),
) Rig {
	r := &rig{
		unique:    Strf("%d", os.Getpid()),
		register:  register,
		log:       config.Log,
		rateLimit: config.Redis.RateLimit,
		cache:     config.Redis.Cache,
		email:     config.Email,
		store:     config.Store,
		fcm:       config.FCM,
		user:      config.SQL.Users,
		data:      config.SQL.Data,
		useAuth:   useAuth,
	}

	for _, bucket := range buckets {
		r.store.MustCreateBucket(bucket, "private")
	}

	if useAuth {
		eps = app.JoinEps(
			eps,
			autheps.New(
				config.App.FromEmail,
				config.App.ActivateFmtLink,
				config.App.ConfirmChangeEmailFmtLink,
				configAuth...))
	}
	Go(func() {
		app.Run(func(c *app.Config) {
			c.ProvideApiDocs = false
			c.TlbxSetup = app.TlbxMwares{
				session.BasicMware(
					config.Web.Session.AuthKey64s,
					config.Web.Session.EncrKey32s,
					config.Web.Session.Secure),
				rateLimitMware(r.rateLimit, 1000000),
				service.Mware(r.cache, r.user, r.data, r.email, r.store, r.fcm),
			}
			c.Endpoints = eps
			c.Serve = func(h http.HandlerFunc) {
				r.rootHandler = h
			}
		})
	}, r.log.ErrorOn)

	// sleep to ensure r.rootHandler has been passed to rig struct
	time.Sleep(100 * time.Millisecond)
	r.ali = r.createUser("ali", emailSuffix, pwd)
	r.bob = r.createUser("bob", emailSuffix, pwd)
	r.cat = r.createUser("cat", emailSuffix, pwd)
	r.dan = r.createUser("dan", emailSuffix, pwd)
	return r
}

func (r *rig) CleanUp() {
	if r.useAuth {
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
}

func (r *rig) createUser(handle, emailSuffix, pwd string) *testUser {
	email := Strf("%s%s%s", handle, emailSuffix, r.unique)
	c := r.NewClient()
	if r.useAuth {
		reg := &auth.Register{
			Email: email,
			Pwd:   pwd,
		}
		if r.register != nil {
			r.register(r, handle, reg)
		}
		(reg).MustDo(c)

		var code string
		row := r.User().Primary().QueryRow(`SELECT activateCode FROM auths WHERE email=?`, email)
		PanicOn(row.Scan(&code))

		(&auth.Activate{
			Email: email,
			Code:  code,
		}).MustDo(c)

		id := (&auth.Login{
			Email: email,
			Pwd:   pwd,
		}).MustDo(c)

		return &testUser{
			client: c,
			id:     id,
			email:  email,
			pwd:    pwd,
		}
	}
	return &testUser{client: c}
}
