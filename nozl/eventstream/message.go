package eventstream

import (
	"time"

	"github.com/google/uuid"
)

type (
	ReqBody   map[string]interface{}
	URLParams map[string]interface{}

	Message struct {
		ID          string    `json:"id"`
		ServiceID   string    `json:"service_id"`
		OperationID string    `json:"operation_id"`
		TenantID    string    `json:"tenant_id"`
		CreatedAt   string    `json:"created_at"`
		SentAt      string    `json:"sent_at"`
		ReqBody     ReqBody   `json:"req_body"`
		URLParams   URLParams `json:"url_params"`
	}

	MessageFilterStatus struct {
		Allow  bool   `json:"allow"`
		Reason string `json:"reason"`
	}
)

func NewMessage(serviceID string, operationID string, body ReqBody, urlParams URLParams) *Message {
	return &Message{
		ID:          uuid.New().String(),
		ServiceID:   serviceID,
		OperationID: operationID,
		ReqBody:     body,
		URLParams:   urlParams,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
}
