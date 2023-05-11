package schema

import (
	"encoding/json"

	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type SchemaFile struct {
	serviceID string
	FileName  string
	FileID    string
}

func NewSchemaFile(serviceID string, fileName string, fileID string) *SchemaFile {
	return &SchemaFile{
		serviceID: serviceID,
		FileName:  fileName,
		FileID:    fileID,
	}
}

func AddSchemaFileToKVStore(schemaFile *SchemaFile) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)
	if err != nil {
		shared.Logger.Error("Failed to retreive schema file KV store!")
		return err
	}

	jsonPayload, _ := json.Marshal(schemaFile)
	kv.Put(schemaFile.FileID, jsonPayload)
	return nil
}
