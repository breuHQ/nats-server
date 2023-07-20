package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
		kv.Put(shared.UserTokenRate, []byte(shared.TokenRateDefault))
	}

	_, err = kv.Get(shared.UserBucketSize)
	if err != nil {
		kv.Put(shared.UserBucketSize, []byte(shared.BucketSizeDefault))
	}

	_, err = kv.Get(shared.MainLimiterRate)
	if err != nil {
		kv.Put(shared.MainLimiterRate, []byte(shared.TokenRateDefault))
	}

	_, err = kv.Get(shared.MainLimiterBucketSize)
	if err != nil {
		kv.Put(shared.MainLimiterBucketSize, []byte(shared.BucketSizeDefault))
	}
}

func (c *core) initKVStore(bucketName string, bucketDescription string, replicationFactor int) {
	kv, err := eventstream.Eventstream.CreateKeyValStore(bucketName, bucketDescription, replicationFactor)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	c.KVStore = kv
}

// Sends message on the Filter subject.
func (c *core) Send(msg *eventstream.Message) {
	eventstream.Eventstream.PublishEncodedMessage("Filter", msg)
}

// TODO: Get limiter values from API.
func (c *core) filterLimiterAllow(msg *eventstream.Message, filterOnID string) bool {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.FilterLimiterKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return false
	}

	flRaw, err := kv.Get(filterOnID)
	if err != nil {
		c.RegisterFilter(kv, filterOnID)
		return true
	}

	fl := &rate.Limiter{}
	err = json.Unmarshal(flRaw.Value(), fl)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	allow := fl.Allow()
	flUpdated, err := json.Marshal(fl)
	if err != nil {
		shared.Logger.Error(err.Error())
	}
	_, err = kv.Put(filterOnID, flUpdated)
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
	if err != nil {
		shared.Logger.Error(err.Error())
	}

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

func (c *core) RegisterFilter(kv nats.KeyValue, filterOnID string) {
	confKeyAll := []string{shared.UserTokenRate, shared.UserBucketSize}
	confMap := eventstream.GetMultValKVstore(shared.ConfigKV, confKeyAll)
	TokenRate, _ := strconv.Atoi(string(confMap[shared.UserTokenRate]))
	BucketSize, _ := strconv.Atoi(string(confMap[shared.UserBucketSize]))

	newFilter := rate.NewLimiter(rate.Limit(TokenRate), BucketSize)
	newFilter.Allow()
	newFilterRaw, err := json.Marshal(newFilter)
	if err != nil {
		shared.Logger.Error(err.Error())
		return
	}

	_, err = kv.Put(filterOnID, newFilterRaw)
	if err != nil {
		shared.Logger.Error(err.Error())
		return
	}
}

func (c *core) handleLimiter(msg *eventstream.Message) {
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

func (c *core) filterLimitChecker(msg *eventstream.Message, filterOnIDStr string) {
	if c.filterLimiterAllow(msg, filterOnIDStr) {
		shared.Logger.Info(msg.ID,
			zap.String("Subject", "Filter"),
			zap.String("Status", "Allowed"),
		)
		go c.handleLimiter(msg)
		allowMessage(true, "ok")
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

		allowMessage(false, "Rate Limit Exceeded")
	}
}

func (c *core) handleFilter(msg *eventstream.Message) {
	err := schema.ValidateOpenAPIV3Schema(msg)
	if err != nil {
		// TODO: Decide later if this message should be sent to dead letter queue
		allowMessage(false, string(err.Error()))
		shared.Logger.Error(err.Error())
		return
	}
	serv, err := c.getServiceFromMsg(msg)
	if err != nil {
		allowMessage(false, string(err.Error()))
		shared.Logger.Error(err.Error())
		return
	}
	filterOnID, exists := msg.ReqBody[serv.FilterOn]
	if !exists {
		allowMessage(false, fmt.Sprintf("Incorrect filter_on field name. Correct filter_on field name for this service: %s", serv.FilterOn))
		return
	}
	filterOnIDStr := filterOnID.(string)
	c.filterLimitChecker(msg, filterOnIDStr)
}

func (c *core) sendToService(msg *eventstream.Message) error {
	shared.Logger.Info("Message sent to service")

	serv, err := c.getServiceFromMsg(msg)
	if err != nil {
		return err
	}

	t := service.TwilioHTTP{}
	v := service.VonageHTTP{}
	sg := service.SendGridHTTP{}
	var result []byte

	switch serv.Category {
	case service.Twilio:
		result, err = t.GenericHTTPRequest(serv, msg)
	case service.Vonage:
		result, err = v.GenericHTTPRequest(serv, msg)
	case service.SendGrid:
		result, err = sg.GenericHTTPRequest(serv, msg)
	}

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
		return errors.New("failed to retreive msgLog KV store")
	}

	msg.SentAt = time.Now().Format("2006-01-02 15:04:05")

	jsonPayload, _ := json.Marshal(msg)
	if _, err = kv.Put(msg.ID, jsonPayload); err != nil {
		return err
	}

	return nil
}

func (c *core) getServiceFromMsg(msg *eventstream.Message) (*service.Service, error) {
	currService := service.NewService("", map[string]string{}, "", "")

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, errors.New("failed to retreive service KV store")
	}

	entry, err := kv.Get(msg.ServiceID)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, errors.New("serviceID is incorrect")
	}

	if err := json.Unmarshal(entry.Value(), &currService); err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	return currService, err
}
