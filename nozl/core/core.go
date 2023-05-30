package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/rate"
	"github.com/nats-io/nats-server/v2/nozl/schema"
	"github.com/nats-io/nats-server/v2/nozl/service"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type (
	core struct {
		KVStore nats.KeyValue
	}

	Subject string
)

const (
	Filter      Subject = "Filter"
	MainLimiter Subject = "MainLimiter"
)

var (
	ErrTenantDoesNotExist = errors.New("tenant does not exist")
	ErrFilterDoesNotExist = errors.New("filter does not exist")
)

var (
	Core = &core{}
)

func (s Subject) String() string {
	return string(s)
}

func (c *core) InitSubscriptions() {
	c.initFilter()
	c.initMainLimiter()
}

func (c *core) InitStores(replicationFactor int) {
	c.initKVStore(shared.ServiceKV, "", replicationFactor)
	c.initKVStore(shared.TenantKV, "", replicationFactor)
	c.initKVStore(shared.UserKV, "", replicationFactor)
	c.initKVStore(shared.MsgWaitListKV, "", replicationFactor)
	c.initKVStore(shared.MsgLogKV, "", replicationFactor)
	c.initKVStore(shared.TenantAPIKV, "", replicationFactor)
	c.initKVStore(shared.SchemaKV, "", replicationFactor)
	c.initKVStore(shared.SchemaFileKV, "", replicationFactor)
	c.initKVStore(shared.FilterLimiterKV, "", replicationFactor)
	c.initKVStore(shared.MainLimiterKV, "", replicationFactor)
	c.initKVStore(shared.ConfigKV, "", replicationFactor)
}

func (c *core) InitConf() {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ConfigKV)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	_, err = kv.Get(shared.UserTokenRate)
	if err != nil {
		kv.Put(shared.UserTokenRate, []byte("1"))
	}

	_, err = kv.Get(shared.UserBucketSize)
	if err != nil {
		kv.Put(shared.UserBucketSize, []byte("1"))
	}

	_, err = kv.Get(shared.MainLimiterRate)
	if err != nil {
		kv.Put(shared.MainLimiterRate, []byte("1"))
	}

	_, err = kv.Get(shared.MainLimiterBucketSize)
	if err != nil {
		kv.Put(shared.MainLimiterBucketSize, []byte("1"))
	}
}

func (c *core) initKVStore(bucketName string, bucketDescription string, replicationFactor int) {
	kv, err := eventstream.Eventstream.CreateKeyValStore(bucketName, bucketDescription, replicationFactor)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	c.KVStore = kv
}

func (c *core) initFilter() {
	_, err := eventstream.Eventstream.EnCon.Subscribe(Filter.String(), handleFilter(c))
	if err != nil {
		shared.Logger.Error("Error subscribing to Filter subject", zap.Error(err))
	}
}

func (c *core) initMainLimiter() {
	_, err := eventstream.Eventstream.EnCon.Subscribe(MainLimiter.String(), handleLimiter(c))
	if err != nil {
		shared.Logger.Error("Error subscribing to MainLimiter subject", zap.Error(err))
	}
}

// Sends message on the Filter subject.
func (c *core) Send(msg *eventstream.Message) {
	eventstream.Eventstream.PublishEncodedMessage("Filter", msg)
}

// TODO: Get limiter values from API.
func (c *core) filterLimiterAllow(msg *eventstream.Message) bool {
	userID := msg.ReqBody["user_id"].(string)
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.FilterLimiterKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return false
	}

	flRaw, err := kv.Get(userID)
	if err != nil {
		c.RegisterFilter(kv, userID)
		return true
	}

	fl := &rate.Limiter{}
	err = json.Unmarshal(flRaw.Value(), fl)

	allow := fl.Allow()
	flUpdated, err := json.Marshal(fl)
	_, err = kv.Put(userID, flUpdated)
	if err != nil {
		shared.Logger.Error(err.Error())
		return false
	}

	return allow
}

func (c *core) mainLimiterWait(msg *eventstream.Message) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MainLimiterKV)
	if err != nil {
		return err
	}

	mlRaw, err := kv.Get(msg.ServiceID)
	if err != nil {
		return err
	}

	ml := &rate.Limiter{}
	err = json.Unmarshal(mlRaw.Value(), ml)

	if err := ml.Wait(context.Background()); err != nil {
		return err
	}

	mlUpdated, err := json.Marshal(ml)
	if err != nil {
		return err
	}

	_, err = kv.Put(msg.ServiceID, mlUpdated)
	if err != nil {
		return err
	}

	return nil
}

func (c *core) RegisterFilter(kv nats.KeyValue, userID string) {
	confKeyAll := []string{shared.UserTokenRate, shared.UserBucketSize}
	confMap := eventstream.GetMultValIntKVstore(shared.ConfigKV, confKeyAll)
	TokenRate := confMap[shared.UserTokenRate]
	BucketSize := confMap[shared.UserBucketSize]

	newFilter := rate.NewLimiter(rate.Limit(TokenRate), BucketSize)
	newFilter.Allow()
	newFilterRaw, err := json.Marshal(newFilter)
	if err != nil {
		shared.Logger.Error(err.Error())
		return
	}

	_, err = kv.Put(userID, newFilterRaw)
	if err != nil {
		shared.Logger.Error(err.Error())
		return
	}
	return
}

func handleLimiter(c *core) nats.Handler {
	return func(msg *eventstream.Message) {

		if err := c.mainLimiterWait(msg); err != nil {
			shared.Logger.Error(err.Error())
		}

		if err := c.sendToService(msg); err != nil {
			shared.Logger.Error(err.Error())
		}

		shared.Logger.Info(msg.ID,
			zap.String("Subject", "MainLimiter"),
			zap.String("Status", "Allowed"),
		)
	}
}

func handleFilter(c *core) nats.Handler {
	return func(msg *eventstream.Message) {
		err := schema.ValidateOpenAPIV3Schema(msg)
		if err != nil {
			// TODO: Decide later if this message should be sent to dead letter queue
			eventstream.MessageFilterAllow <- &eventstream.MessageFilterStatus{
				Allow:  false,
				Reason: string(err.Error()), // TODO: return schema specific error
			}
			shared.Logger.Error(err.Error())
			return
		}
		if c.filterLimiterAllow(msg) {
			shared.Logger.Info(msg.ID,
				zap.String("Subject", "Filter"),
				zap.String("Status", "Allowed"),
			)
			eventstream.Eventstream.PublishEncodedMessage("MainLimiter", msg)
			eventstream.MessageFilterAllow <- &eventstream.MessageFilterStatus{
				Allow:  true,
				Reason: "ok",
			}
		} else {
			queuePayload, _ := json.Marshal(msg)

			kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MsgWaitListKV)
			if err != nil {
				shared.Logger.Error(err.Error())
			}

			if _, err := kv.Put(msg.ID, queuePayload); err != nil {
				shared.Logger.Error(err.Error())
			}

			shared.Logger.Info(msg.ID,
				zap.String("Subject", "Filter"),
				zap.String("Status", "Rejected"),
			)

			eventstream.MessageFilterAllow <- &eventstream.MessageFilterStatus{
				Allow:  false,
				Reason: "Rate Limit Exceeded",
			}
		}
	}
}

func (c *core) sendToService(msg *eventstream.Message) error {
	shared.Logger.Info("Message sent to service")

	serv, err := c.getServiceFromMsg(msg)
	if err != nil {
		return err
	}

	t := service.Twilio{}
	result, err := t.GenericHTTPRequest(serv, msg)

	if err != nil {
		return err
	}

	eventstream.ServiceResponse <- result

	err = c.LogSentMessage(msg)

	return err
}

func (c *core) LogSentMessage(msg *eventstream.Message) error {
	// fmt.Println("Logging sent message")
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MsgLogKV)
	if err != nil {
		return err
	}

	msg.SentAt = time.Now().Format("2006-01-02 15:04:05")

	jsonPayload, _ := json.Marshal(msg)
	if _, err = kv.Put(msg.ID, jsonPayload); err != nil {
		return err
	}

	return nil
}

func (c *core) getServiceFromMsg(msg *eventstream.Message) (*service.Service, error) {
	currService := service.NewService("", "", "")

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	entry, err := kv.Get(msg.ServiceID)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	if err := json.Unmarshal(entry.Value(), &currService); err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	return currService, err
}
