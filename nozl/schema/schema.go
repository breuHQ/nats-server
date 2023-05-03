package schema

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type Schema struct {
	Path        string
	HttpMethod  string
	PathDetails *openapi3.Operation
}

func addSchema(pathKey string, httpMethod string, pathDetails *openapi3.Operation) Schema {
	return Schema{
		Path:        pathKey,
		HttpMethod:  httpMethod,
		PathDetails: pathDetails,
	}
}

func ParseOpenApiV3Schema(serviceID string, specFile []byte) error {
	doc, err := openapi3.NewLoader().LoadFromData(specFile)

	if err != nil {
		shared.Logger.Error("Failed to ready open API spec!")
		return err
	}

	var schemaList map[string]Schema
	schemaKv, _ := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	entry, err := schemaKv.Get(serviceID)

	if err != nil {
		schemaList = make(map[string]Schema)
	} else {
		err := json.Unmarshal(entry.Value(), &schemaList)
		if err != nil {
			return err
		}
	}

	for pathKey, pathValue := range doc.Paths {

		if pathValue.Get != nil {
			schemaList[pathValue.Get.OperationID] = addSchema(pathKey, "get", pathValue.Get)
		} else if pathValue.Post != nil {
			schemaList[pathValue.Post.OperationID] = addSchema(pathKey, "post", pathValue.Post)
		} else if pathValue.Put != nil {
			schemaList[pathValue.Put.OperationID] = addSchema(pathKey, "put", pathValue.Put)
		} else if pathValue.Patch != nil {
			schemaList[pathValue.Patch.OperationID] = addSchema(pathKey, "patch", pathValue.Patch)
		} else if pathValue.Delete != nil {
			schemaList[pathValue.Delete.OperationID] = addSchema(pathKey, "delete", pathValue.Delete)
		}
	}

	jsonPayload, _ := json.Marshal(schemaList)

	_, err = schemaKv.Put(serviceID, jsonPayload)

	if err != nil {
		shared.Logger.Error("Failed to save schema in data store.")
		return err
	}

	return nil
}

func ValidateOpenAPIV3Schema(msg *eventstream.Message) (bool, error) {
	schemaValid, err := GetMsgValidSchema(msg)
	if err != nil {
		shared.Logger.Error(err.Error())
		return false, err
	}
	msgBody := msg.Body
	fmt.Println(schemaValid)
	fmt.Println(msgBody)

	return false, nil
}

func GetMsgValidSchema(msg *eventstream.Message) (Schema, error) {
	schemaKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return Schema{}, err
	}

	entry, err := schemaKv.Get(msg.ServiceID)
	if err != nil {
		shared.Logger.Error(err.Error())
		return Schema{}, err
	}

	var schemaList map[string]Schema
	err = json.Unmarshal(entry.Value(), &schemaList)
	if err != nil {
		shared.Logger.Error(err.Error())
		return Schema{}, err
	}

	return schemaList[msg.OperationID], nil
}
