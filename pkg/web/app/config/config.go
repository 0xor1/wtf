package config

import (
	"context"
	"encoding/base64"
	"time"

	firebase "firebase.google.com/go"
	"github.com/0xor1/tlbx/pkg/config"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	sp "github.com/SparkPost/gosparkpost"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"google.golang.org/api/option"
)

type Config struct {
	Version string
	Log     log.Log
	Web     struct {
		StaticDir               string
		ContentSecurityPolicies []string
		RateLimit               int
		Session                 struct {
			Secure     bool
			AuthKey64s [][]byte
			EncrKey32s [][]byte
		}
	}
	App struct {
		FromEmail                 string
		ActivateFmtLink           string
		ConfirmChangeEmailFmtLink string
	}
	Redis struct {
		RateLimit iredis.Pool
		Cache     iredis.Pool
	}
	SQL struct {
		Users isql.ReplicaSet
		Data  isql.ReplicaSet
	}
	Email email.Client
	Store store.Client
	FCM   fcm.Client
}

func GetBase(file ...string) *config.Config {
	c := config.New(file...)
	c.SetDefault("version", "dev")
	c.SetDefault("log.type", "local")
	c.SetDefault("web.staticDir", "client/dist")
	c.SetDefault("web.contentSecurityPolicies", []string{})
	c.SetDefault("web.rateLimit", 300)
	// session cookie store
	c.SetDefault("web.session.secure", false)
	c.SetDefault("web.session.authKey64s", []string{
		"Va3ZMfhH4qSfolDHLU7oPal599DMcL93A80rV2KLM_om_HBFFUbodZKOHAGDYg4LCvjYKaicodNmwLXROKVgcA",
		"WK_2RgRx6vjfWVkpiwOCB1fvv1yklnltstBjYlQGfRsl6LyVV4mkt6UamUylmkwC8MEgb9bSGr1FYgM2Zk20Ug",
	})
	c.SetDefault("web.session.encrKey32s", []string{
		"3ICuYRUelY-4Fhak0Iw0_5CW24bJvxFWM0jAA78IIp8",
		"u80sYkgbBav52fJXbENYhN3Iyof7WhuLHHMaS_rmUQw",
	})
	c.SetDefault("app.fromEmail", "test@test.localhost")
	c.SetDefault("app.activateFmtLink", "http://localhost:8081/#/activate?email=%s&code=%s")
	c.SetDefault("app.confirmChangeEmailFmtLink", "http://localhost:8081/#/confirmChangeEmail?me=%s&code=%s")
	c.SetDefault("redis.rateLimit", "localhost:6379")
	c.SetDefault("redis.cache", "localhost:6379")
	c.SetDefault("sql.users.primary", "users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("sql.users.slaves", []string{})
	c.SetDefault("sql.data.primary", "data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("sql.data.slaves", []string{})
	c.SetDefault("sql.connMaxLifetime", 5*time.Second)
	c.SetDefault("sql.maxIdleConns", 50)
	c.SetDefault("sql.maxOpenConns", 100)
	c.SetDefault("email.type", "local")
	c.SetDefault("email.apikey", "")
	c.SetDefault("aws.region", "local")
	c.SetDefault("aws.ses.creds.id", "localtest")
	c.SetDefault("aws.ses.creds.secret", "localtest")
	c.SetDefault("aws.s3.endpoint", "http://localhost:9000")
	c.SetDefault("aws.s3.creds.id", "localtest")
	c.SetDefault("aws.s3.creds.secret", "localtest")
	c.SetDefault("fcm.serviceAccountKeyFile", "")

	return c
}

func GetProcessed(c *config.Config) *Config {
	res := &Config{}

	res.Version = c.GetString("version")

	switch c.GetString("log.type") {
	case "local":
		res.Log = log.New()
	default:
		PanicIf(true, "unsupported log type %s", c.GetString("log.type"))
	}

	res.Web.StaticDir = c.GetString("web.staticDir")
	res.Web.ContentSecurityPolicies = c.GetStringSlice("web.contentSecurityPolicies")
	res.Web.RateLimit = c.GetInt("web.rateLimit")
	res.Web.Session.Secure = c.GetBool("web.session.secure")
	authKey64s := c.GetStringSlice("web.session.authKey64s")
	encrKey32s := c.GetStringSlice("web.session.encrKey32s")
	for i := range authKey64s {
		authBytes, err := base64.RawURLEncoding.DecodeString(authKey64s[i])
		PanicOn(err)
		PanicIf(len(authBytes) != 64, "sessionAuthBytes length is not 64")
		res.Web.Session.AuthKey64s = append(res.Web.Session.AuthKey64s, authBytes)
		encrBytes, err := base64.RawURLEncoding.DecodeString(encrKey32s[i])
		PanicOn(err)
		PanicIf(len(encrBytes) != 32, "sessionEncrBytes length is not 32")
		res.Web.Session.EncrKey32s = append(res.Web.Session.EncrKey32s, encrBytes)
	}

	res.App.FromEmail = c.GetString("app.fromEmail")
	res.App.ActivateFmtLink = c.GetString("app.activateFmtLink")
	res.App.ConfirmChangeEmailFmtLink = c.GetString("app.confirmChangeEmailFmtLink")

	res.Redis.RateLimit = iredis.CreatePool(c.GetString("redis.rateLimit"))
	res.Redis.Cache = iredis.CreatePool(c.GetString("redis.cache"))

	sqlMaxLifetime := c.GetDuration("sql.connMaxLifetime")
	sqlMaxIdleConns := c.GetInt("sql.maxIdleConns")
	sqlMaxOpenConns := c.GetInt("sql.maxOpenConns")

	var err error
	res.SQL.Users, err = isql.NewReplicaSet(c.GetString("sql.users.primary"), c.GetStringSlice("sql.users.slaves")...)
	PanicOn(err)
	res.SQL.Users.SetConnMaxLifetime(sqlMaxLifetime)
	res.SQL.Users.SetMaxIdleConns(sqlMaxIdleConns)
	res.SQL.Users.SetMaxOpenConns(sqlMaxOpenConns)

	res.SQL.Data, err = isql.NewReplicaSet(c.GetString("sql.data.primary"), c.GetStringSlice("sql.data.slaves")...)
	PanicOn(err)
	res.SQL.Data.SetConnMaxLifetime(sqlMaxLifetime)
	res.SQL.Data.SetMaxIdleConns(sqlMaxIdleConns)
	res.SQL.Data.SetMaxOpenConns(sqlMaxOpenConns)

	switch c.GetString("email.type") {
	case "local":
		res.Email = email.NewLocalClient(res.Log)
	case "sparkpost":
		spClient := &sp.Client{}
		PanicOn(spClient.Init(&sp.Config{
			BaseUrl:    "https://api.eu.sparkpost.com",
			ApiKey:     c.GetString("email.apikey"),
			ApiVersion: 1,
		}))
		res.Email = email.NewSparkPostClient(spClient)
	case "ses":
		res.Email = email.NewSESClient(
			ses.New(
				session.New(
					&aws.Config{
						Region:      ptr.String(c.GetString("aws.region")),
						Credentials: credentials.NewStaticCredentials(c.GetString("aws.ses.creds.id"), c.GetString("aws.ses.creds.secret"), ""),
					})))
	default:
		PanicIf(true, "unsupported email type %s", c.GetString("email.type"))
	}

	res.Store = store.New(
		s3.New(
			session.New(
				&aws.Config{
					Region:           ptr.String(c.GetString("aws.region")),
					Endpoint:         ptr.String(c.GetString("aws.s3.endpoint")),
					Credentials:      credentials.NewStaticCredentials(c.GetString("aws.s3.creds.id"), c.GetString("aws.s3.creds.secret"), ""),
					DisableSSL:       ptr.Bool(true),
					S3ForcePathStyle: ptr.Bool(true),
				})))

	if c.GetString("fcm.serviceAccountKeyFile") != "" {
		opt := option.WithCredentialsFile(c.GetString("fcm.serviceAccountKeyFile"))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		PanicOn(err)
		client, err := app.Messaging(context.Background())
		PanicOn(err)
		res.FCM = fcm.NewClient(client)
	} else {
		res.FCM = fcm.NewNopClient(res.Log)
	}

	return res
}
