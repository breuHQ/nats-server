package service

import (
	"github.com/google/uuid"
)

const (
	Twilio   string = "Twilio"
	SendGrid string = "SendGrid"
	Vonage   string = "Vonage"
	Custom   string = "Custom"
)

type (
	// TODO: Fix JSON validation issues
	Service struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		FilterOn    string            `json:"filter_on"`
		AuthDetails map[string]string `json:"auth_details"`
		Category    string            `json:"category"`
	}
)

func NewService(name string, authDetails map[string]string, filterOn string, category string) *Service {
	return &Service{
		ID:          uuid.New().String(),
		Name:        name,
		FilterOn:    filterOn,
		AuthDetails: authDetails,
		Category:    category,
	}
}
