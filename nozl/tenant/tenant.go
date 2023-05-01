package tenant

import (
	"fmt"

	"github.com/google/uuid"
)

type (
	Tenant struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		APIKey string `json:"api_key"`
	}
)

func NewTenant(name string) *Tenant {
	tnt := &Tenant{
		ID:   uuid.New().String(),
		Name: name,
	}
	tnt.generateAPIKey()

	return tnt
}

func (t *Tenant) generateAPIKey() {
	t.APIKey = uuid.New().String()
}

func (t *Tenant) GetDetails() string {
	return fmt.Sprintf(
		"ID: %s, Name: %s",
		t.ID, t.Name,
	)
}

func (t *Tenant) GetID() string {
	return t.ID
}

func (t *Tenant) GetAPIKey() string {
	return t.APIKey
}
