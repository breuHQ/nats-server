package schema

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type SchemaFile struct {
	ServiceID   string
	FileName    string
	FileID      string
	CreatedAt string
	UpdatedAt string
}

func GetDate() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func NewSchemaFile(serviceID string, fileName string) SchemaFile {
	return SchemaFile{
		ServiceID:   serviceID,
		FileName:    fileName,
		FileID:      uuid.New().String(),
		CreatedAt: GetDate(),
		UpdatedAt: GetDate(),
	}
}

func addSchemaFileToKVStore(schemaFile SchemaFile) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)

	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return err
	}

	jsonPayload, _ := json.Marshal(schemaFile)
	kv.Put(schemaFile.FileID, jsonPayload)
	return nil
}

func AddSchemaFile(serviceID string, fileName string) (SchemaFile, error) {
	schemaDetails := NewSchemaFile(serviceID, fileName)
	if err := addSchemaFileToKVStore(schemaDetails); err != nil {
		shared.Logger.Error("Failed to add schemaDetails to KV store!")
		return schemaDetails, err
	}
	return schemaDetails, nil
}

func GetSchemaFile(serviceID string, fileName string) (SchemaFile, error) {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)
	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return SchemaFile{}, err
	}

	allKeys, err := kv.Keys()
	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return SchemaFile{}, err
	}

	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			shared.Logger.Error("Failed to retreive value from KV store!")
			return SchemaFile{}, err
		}

		var schemaFile SchemaFile
		if err := json.Unmarshal(value.Value(), &schemaFile); err != nil {
			shared.Logger.Error("Failed to unmarshal value from KV store!")
			return SchemaFile{}, err
		}

		if schemaFile.FileName == fileName && schemaFile.ServiceID == serviceID {
			return schemaFile, nil
		}
	}

	return SchemaFile{}, nil
}
