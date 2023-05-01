package eventstream

import (
	"time"

	"github.com/google/uuid"
)

type (
	Message struct {
		ID          string `json:"id"`
		ServiceID   string `json:"service_id"`
		OperationID string `json:"operation_id"`
		TenantID    string `json:"tenant_id"`
		CreatedAt   string `json:"created_at"`
		SentAt      string `json:"sent_at"`
		Body        Body   `json:"body"`
	}

	Body struct {
		UserID      string `json:"user_id"`
		Payload     string `json:"payload"`
		Destination string `json:"destination"`
	}
)

func NewMessage(serviceID string, operationID string, body Body) *Message {
	return &Message{
		ID:          uuid.New().String(),
		ServiceID:   serviceID,
		OperationID: operationID,
		Body:        body,
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
