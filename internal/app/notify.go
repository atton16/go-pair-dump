package app

import (
	"context"
	"log"

	"github.com/atton16/go-pair-dump/internal/services"
	jsoniter "github.com/json-iterator/go"
)

type PairdumpStatus string
type PairdumpScope string
type PairdumpStatusMessage struct {
	Status  PairdumpStatus `json:"status"`
	Scope   PairdumpScope  `json:"scope,omitempty"`
	Message string         `json:"message,omitempty"`
}

const (
	StatusStart PairdumpStatus = "start"
	StatusDone  PairdumpStatus = "done"
	StatusError PairdumpStatus = "error"
)

const (
	AppGetSymbols  PairdumpScope = "app.GetSymbols"
	AppGetKlines   PairdumpScope = "app.GetKlines"
	AppEnsureIndex PairdumpScope = "app.EnsureIndex"
	AppBulkWrite   PairdumpScope = "app.BulkWrite"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (status PairdumpStatus) MarshalBinary() ([]byte, error) {
	return []byte(status), nil
}

func (scope PairdumpScope) MarshalBinary() ([]byte, error) {
	return []byte(scope), nil
}

func NotifyOK(ctx context.Context, status PairdumpStatus) {
	var config = services.GetConfig()
	var rd = services.GetRedis()
	if config.Notification.Enable {
		m, _ := json.Marshal(PairdumpStatusMessage{Status: status})
		log.Printf("notification: NotifyOK -> %s\n", m)
		result, err := rd.Publish(ctx, config.Notification.Channel, m)
		log.Printf("notification: NotifyOK -> result=%d, error=%v", result, err)
	}
}

func NotifyError(ctx context.Context, scope PairdumpScope, err error) {
	var config = services.GetConfig()
	var rd = services.GetRedis()
	if config.Notification.Enable {
		m, _ := json.Marshal(PairdumpStatusMessage{Status: StatusError, Scope: scope, Message: err.Error()})
		log.Printf("notification: NotifyError -> %s\n", m)
		result, err := rd.Publish(ctx, config.Notification.Channel, m)
		log.Printf("notification: NotifyError -> result=%d, error=%v", result, err)
	}
}
