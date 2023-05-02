package schema

import (
	// "encoding/json"
	"encoding/json"
	// "fmt"

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

func ParseOpenApiV3Schema() error {
	doc, err := openapi3.NewLoader().LoadFromFile("/home/tam/Documents/codes/breu/nats-server/twilio_api_v2010.json")

	if err != nil {
		shared.Logger.Error("Failed to ready open API spec!")
		return err
	}

	schemaKv, _ := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)
	schemaList := map[string]Schema{}

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

	_, err = schemaKv.Put("service_id", jsonPayload)

	if err != nil {
		shared.Logger.Error("Failed to save schema in data store.")
		return err
	}

	// if err = schemaKv.Get("service_id"); err != nil {
	// 	schemaKv.Put("service_id", jsonPayload)
	// }

	return nil
}
