package fcm

import (
	"context"

	"firebase.google.com/go/messaging"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/log"
)

type Client interface {
	Send(ctx context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error)
	MustSend(ctx context.Context, m *messaging.MulticastMessage) *messaging.BatchResponse
}

func NewClient(fcm *messaging.Client) Client {
	return &client{
		fcm: fcm,
	}
}

type client struct {
	fcm *messaging.Client
}

func (c *client) Send(ctx context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	allTs := m.Tokens
	res := &messaging.BatchResponse{}
	// can't send more than 500 at a time so send batches if over that limit
	Println("yolo1", m.Tokens, m.Data, m)
	for len(allTs) > 0 {
		if len(allTs) > 500 {
			m.Tokens = allTs[:500]
			allTs = allTs[500:]
		} else {
			m.Tokens = allTs
			allTs = nil
		}
		Println("yolo2", m.Tokens, m.Data, m)
		curRes, err := c.fcm.SendMulticast(ctx, m)
		if curRes != nil {
			res.FailureCount += curRes.FailureCount
			res.SuccessCount += curRes.SuccessCount
			res.Responses = append(res.Responses, curRes.Responses...)
		}
		if err != nil {
			return res, ToError(err)
		}
	}
	return res, nil
}

func (c *client) MustSend(ctx context.Context, m *messaging.MulticastMessage) *messaging.BatchResponse {
	res, err := c.Send(ctx, m)
	PanicOn(err)
	return res
}

func NewNoopClient(l log.Log) Client {
	return &noopClient{
		log: l,
	}
}

type noopClient struct {
	log log.Log
}

func (c *noopClient) Send(ctx context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	c.log.Warning("noop fcm client called for %d tokens", len(m.Tokens))
	return &messaging.BatchResponse{}, nil
}

func (c *noopClient) MustSend(ctx context.Context, m *messaging.MulticastMessage) *messaging.BatchResponse {
	res, err := c.Send(ctx, m)
	PanicOn(err)
	return res
}
