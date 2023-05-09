package eventstream

import (
	"time"

	"github.com/google/uuid"
)

type (
	ReqBody map[string]interface{}

	Message struct {
		ID          string  `json:"id"`
		ServiceID   string  `json:"service_id"`
		OperationID string  `json:"operation_id"`
		TenantID    string  `json:"tenant_id"`
		CreatedAt   string  `json:"created_at"`
		SentAt      string  `json:"sent_at"`
		ReqBody     ReqBody `json:"req_body"`
	}

	Body struct {
		UserID      string `json:"user_id"`
		Payload     string `json:"payload"`
		Destination string `json:"destination"`
	}
)

func NewMessage(serviceID string, operationID string, body ReqBody) *Message {
	return &Message{
		ID:          uuid.New().String(),
		ServiceID:   serviceID,
		OperationID: operationID,
		ReqBody:     body,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
}

func NewBody(userID string, payload string, dest string) *Body {
	return &Body{
		UserID:      userID,
		Payload:     payload,
		Destination: dest,
	}
}
