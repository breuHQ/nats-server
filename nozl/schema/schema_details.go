package schema

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type SchemaFile struct {
	ServiceID string
	FileName  string
	FileID    string
}

func NewSchemaFile(serviceID string, fileName string) SchemaFile {
	return SchemaFile{
		ServiceID: serviceID,
		FileName:  fileName,
		FileID:    uuid.New().String(),
	}
}

func AddSchemaFileToKVStoreHelper(schemaFile SchemaFile) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)

	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return err
	}

	jsonPayload, _ := json.Marshal(schemaFile)
	kv.Put(schemaFile.FileID, jsonPayload)
	return nil
}

func AddSchemaFiletoKVStore(serviceID string, fileName string) (SchemaFile, error) {
	schemaDetails := NewSchemaFile(serviceID, fileName)
	if err := AddSchemaFileToKVStoreHelper(schemaDetails); err != nil {
		shared.Logger.Error("Failed to add schemaDetails to KV store!")
		return schemaDetails, err
	}
	return schemaDetails, nil

	// Checking if schemaDetails exists in KV store

	// if kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaDetailsKV); err == nil {
	// 	if entry, err := kv.Get(schemaDetails.FileID); err == nil {
	// 		payload := NewSchemaFile("", "")
	// 		if err := json.Unmarshal(entry.Value(), &payload); err != nil {
	// 			shared.Logger.Error(err.Error())
	// 		}

	// 		return schemaDetails, nil

	// 	} else {
	// 		fmt.Print("Failed to retreive schemaDetails from KV store!")
	// 		return nil, err
	// 	}
	// }
	// return nil, nil
}
