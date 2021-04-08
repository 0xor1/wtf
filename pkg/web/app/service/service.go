package service

import (
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	emailmw "github.com/0xor1/tlbx/pkg/web/app/service/email"
	fcmmw "github.com/0xor1/tlbx/pkg/web/app/service/fcm"
	"github.com/0xor1/tlbx/pkg/web/app/service/redis"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	storemw "github.com/0xor1/tlbx/pkg/web/app/service/store"
)

const (
	cache     = "cache"
	sqlAuth   = "auth"
	sqlData   = "data"
	emailName = "email"
	storeName = "store"
	fcmName   = "fcm"
)

type Layer interface {
	Cache() redis.Pool
	Auth() sql.Client
	Data() sql.Client
	Email() email.Client
	Store() store.Client
	FCM() fcmmw.Client
}

func Mware(pool iredis.Pool, auth, data isql.ReplicaSet, email email.Client, store store.Client, fcm fcm.Client) func(app.Tlbx) {
	mwares := []func(app.Tlbx){
		redis.Mware(cache, pool),
		sql.Mware(sqlAuth, auth),
		sql.Mware(sqlData, data),
		emailmw.Mware(emailName, email),
		storemw.Mware(storeName, store),
		fcmmw.Mware(sqlAuth, fcmName, fcm),
	}
	return func(tlbx app.Tlbx) {
		for _, mw := range mwares {
			mw(tlbx)
		}
	}
}

type layer struct {
	tlbx app.Tlbx
}

func Get(tlbx app.Tlbx) Layer {
	return &layer{tlbx}
}

func (l *layer) Cache() redis.Pool {
	return redis.Get(l.tlbx, cache)
}

func (l *layer) Auth() sql.Client {
	return sql.Get(l.tlbx, sqlAuth)
}

func (l *layer) Data() sql.Client {
	return sql.Get(l.tlbx, sqlData)
}

func (l *layer) Email() email.Client {
	return emailmw.Get(l.tlbx, emailName)
}

func (l *layer) Store() store.Client {
	return storemw.Get(l.tlbx, storeName)
}

func (l *layer) FCM() fcmmw.Client {
	return fcmmw.Get(l.tlbx, fcmName)
}
