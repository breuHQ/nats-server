package schema

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type SchemaDetails struct {
	ServiceID string
	FileName  string
	FileID    string
}

func NewSchemaDetails(serviceID string, fileName string) *SchemaDetails {
	return &SchemaDetails{
		ServiceID: serviceID,
		FileName:  fileName,
		FileID:    uuid.New().String(),
	}
}

func AddSchemaDetailsToKVStoreHelper(schemaFile *SchemaDetails) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaDetailsKV)

	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return err
	}

	jsonPayload, _ := json.Marshal(schemaFile)
	kv.Put(schemaFile.FileID, jsonPayload)
	return nil
}

func AddSchemaDetailstoKVStore(serviceID string, fileName string) (*SchemaDetails, error) {
	schemaDetails := NewSchemaDetails(serviceID, fileName)
	if err := AddSchemaDetailsToKVStoreHelper(schemaDetails); err != nil {
		shared.Logger.Error("Failed to add schemaDetails to KV store!")
		return nil, err
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
