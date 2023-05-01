package service

import (
	"github.com/google/uuid"
)

type (
	Service struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		AccountSID string `json:"account_sid"`
		AuthToken  string `json:"auth_token"`
	}
)

func NewService(name string, accountSID string, authToken string) *Service {
	return &Service{
		ID:         uuid.New().String(),
		Name:       name,
		AccountSID: accountSID,
		AuthToken:  authToken,
	}
}
