package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/filter"
	"github.com/nats-io/nats-server/v2/nozl/rate"
	"github.com/nats-io/nats-server/v2/nozl/schema"
	"github.com/nats-io/nats-server/v2/nozl/service"
	"github.com/nats-io/nats-server/v2/nozl/shared"
	"github.com/nats-io/nats-server/v2/nozl/tenant"
)

type (
	core struct {
		Tenants     Tenants
		Filters     Filters
		MainLimiter *rate.Limiter
		KVStore     nats.KeyValue
	}

	Tenants map[uuid.UUID]tenant.Tenant
	Filters map[string]filter.Filter
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

func (filters Filters) GetByID(id string) (filter.Filter, error) {
	if fil, exists := filters[id]; exists {
		return fil, nil
	}

	return nil, ErrFilterDoesNotExist
}

func (t Tenants) GetByID(id uuid.UUID) (*tenant.Tenant, error) {
	if tnt, exists := t[id]; exists {
		return &tnt, nil
	}

	return nil, ErrTenantDoesNotExist
}

// Initializes Core Module.
func (c *core) InitializeCore(tokenRate int, initialBucketSize int) {
	r := rate.Limit(tokenRate)
	rl := rate.NewLimiter(r, initialBucketSize)
	c.Tenants = make(map[uuid.UUID]tenant.Tenant)
	c.Filters = make(map[string]filter.Filter)
	c.MainLimiter = rl
}

// Initializes the Filter Service and the Main Limiter Service.
func (c *core) Init() {
	//TODO: Store user filters in nats KV store also
	c.initFilter()
	c.initMainLimiter()
	c.initKVStore(shared.ServiceKV, "")
	c.initKVStore(shared.TenantKV, "")
	c.initKVStore(shared.UserKV, "")
	c.initKVStore(shared.MsgWaitListKV, "")
	c.initKVStore(shared.MsgLogKV, "")
	c.initKVStore(shared.TenantAPIKV, "")
	c.initKVStore(shared.SchemaKV, "")
}

func (c *core) initKVStore(bucketName string, bucketDescription string) {
	kv, err := eventstream.Eventstream.CreateKeyValStore(bucketName, bucketDescription)

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
	if val, exists := c.Filters[userID]; exists {
		return val.Allow()
	}

	c.RegisterFilter(userID, shared.UserTokenRate, shared.UserBucketSize)

	return c.Filters[userID].Allow()
}

func (c *core) RegisterFilter(userID string, tokenRate int, bucketSize int) uuid.UUID {
	newFilter := filter.NewFilter(tokenRate, bucketSize, userID)
	c.Filters[userID] = newFilter

	return newFilter.GetID()
}

// Get Tenant details by providing its id.
func (c *core) GetTenantDetailsByID(id uuid.UUID) (string, error) {
	t, err := c.Tenants.GetByID(id)
	if err != nil {
		return "", err
	}

	return t.GetDetails(), nil
}

func handleLimiter(c *core) nats.Handler {
	return func(msg *eventstream.Message) {
		if err := c.MainLimiter.Wait(context.Background()); err != nil {
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
				Reason: string(err.Error()),
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
