package fcm

import (
	"context"
	"time"

	"firebase.google.com/go/messaging"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
)

type tlbxKey struct {
	name string
}

type Client interface {
	fcm.Client
	AsyncSend(topic IDs, data map[string]string, timeout time.Duration)
	RawAsyncSend(fcmType string, tokens []string, data map[string]string, timeout time.Duration)
}

func Mware(userSqlName, name string, fcm fcm.Client) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{
			tlbx:        tlbx,
			userSqlName: userSqlName,
			name:        name,
			fcm:         fcm,
		})
	}
}

func Get(tlbx app.Tlbx, name string) Client {
	return tlbx.Get(tlbxKey{name}).(Client)
}

type client struct {
	tlbx        app.Tlbx
	userSqlName string
	name        string
	fcm         fcm.Client
}

func (c *client) Send(ctx context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	var err error
	var res *messaging.BatchResponse
	n := len(m.Tokens)
	c.do(func() {
		res, err = c.fcm.Send(ctx, m)
	}, Strf("%s %d", "SEND", n))
	return res, err
}

func (c *client) MustSend(ctx context.Context, m *messaging.MulticastMessage) *messaging.BatchResponse {
	res, err := c.Send(ctx, m)
	PanicOn(err)
	return res
}

func (c *client) do(do func(), action string) {
	start := NowUnixMilli()
	do()
	c.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "FCM",
		Name:   c.name,
		Action: action,
	})
}

func (c *client) AsyncSend(topic IDs, data map[string]string, timeout time.Duration) {
	app.BadReqIf(len(topic) == 0 || len(topic) > 5, "topic must be 1-5 ids long")
	tokens := make([]string, 0, 20)
	sql.Get(c.tlbx, c.userSqlName).Query(func(rows isql.Rows) {
		for rows.Next() {
			token := ""
			PanicOn(rows.Scan(&token))
			tokens = append(tokens, token)
		}
	}, `SELECT DISTINCT f.token FROM fcmTokens f JOIN users u ON f.user=u.id WHERE topic=? AND u.fcmEnabled=1`, topic.StrJoin("_"))
	c.RawAsyncSend("data", tokens, data, timeout)
}

var clientHeaderName = "X-Fcm-Client"
var fcmTypeName = "X-Fcm-Type"

// this should only be used by AsyncSend and in fcmeps
func (c *client) RawAsyncSend(fcmType string, tokens []string, data map[string]string, timeout time.Duration) {
	PanicIf(fcmType == "", "fcmType must be none empty string")
	_, fcmTypeExists := data[fcmTypeName]
	PanicIf(fcmTypeExists, fcmTypeName+" is a reserved fcm push property for internal api use")
	data[fcmTypeName] = fcmType
	if len(tokens) == 0 {
		return
	}
	_, clientExists := data[clientHeaderName]
	PanicIf(clientExists, clientHeaderName+" is a reserved fcm push property for internal api use")
	client := c.tlbx.Req().Header.Get(clientHeaderName)
	if client != "" {
		data[clientHeaderName] = client
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	log := c.tlbx.Log()
	Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		log.Info("doing async call to fcm service with %d tokens", len(tokens))
		res := c.MustSend(ctx, &messaging.MulticastMessage{
			Tokens: tokens,
			Data:   data,
		})
		log.Info("FCM success: %d, fail: %d", res.SuccessCount, res.FailureCount)
	}, log.ErrorOn)
}
