package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
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

	for pathKey, pathValue := range doc.Paths {

		if pathValue.Get != nil {
			AddSchemaToKVStore(serviceID, pathKey, "get", pathValue.Get)
		}
		if pathValue.Post != nil {
			AddSchemaToKVStore(serviceID, pathKey, "post", pathValue.Post)
		}
		if pathValue.Put != nil {
			AddSchemaToKVStore(serviceID, pathKey, "put", pathValue.Put)
		}
		if pathValue.Patch != nil {
			AddSchemaToKVStore(serviceID, pathKey, "patch", pathValue.Patch)
		}
		if pathValue.Delete != nil {
			AddSchemaToKVStore(serviceID, pathKey, "delete", pathValue.Delete)
		}
	}

	return nil
}

func MakeOptionalFieldsNullable(operation *openapi3.Operation) error {
	if operation.RequestBody != nil {
		ReqBodySchema := operation.RequestBody.Value.Content["application/x-www-form-urlencoded"].Schema.Value
		RequiredParams := ReqBodySchema.Required
		ReqBodyParams := ReqBodySchema.Properties

		for key, val := range ReqBodyParams {
			if stringInSlice(key, RequiredParams) == false {
				val.Value.Nullable = true
			}
		}
	}

	return nil
}

func stringInSlice(refStr string, list []string) bool {
	for _, currStr := range list {
		if currStr == refStr {
			return true
		}
	}
	return false
}

func AddSchemaToKVStore(serviceID string, pathKey string, httpMethod string, operation *openapi3.Operation) {
	schemaKv, _ := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	MakeOptionalFieldsNullable(operation)
	currSchema := addSchema(pathKey, httpMethod, operation)
	operationID := operation.OperationID
	jsonPayload, _ := json.Marshal(currSchema)
	schemaKv.Put(fmt.Sprintf("%s-%s", serviceID, operationID), jsonPayload)
}

func ValidateOpenAPIV3Schema(msg *eventstream.Message) error {
	schemaValid, err := GetMsgRefSchema(msg)
	if err != nil {
		shared.Logger.Error(err.Error())
		return err
	}
	openapi3.SchemaErrorDetailsDisabled = true

	msgBody := msg.Body
	fmt.Println(schemaValid)
	fmt.Println(msgBody)

	ctx := context.Background()
	jsonData, _ := json.Marshal(&msgBody)
	formData := strings.NewReader(string(jsonData))

	httpReq, err := http.NewRequest("POST", schemaValid.Path, formData)
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	for key, val := range headers {
		httpReq.Header.Add(key, val)
	}

	input := &openapi3filter.RequestValidationInput{
		Request: httpReq,
	}

	err = openapi3filter.ValidateRequestBody(ctx, input, schemaValid.PathDetails.RequestBody.Value)

	return err
}

func GetMsgRefSchema(msg *eventstream.Message) (*Schema, error) {
	schemaKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	entry, err := schemaKv.Get(fmt.Sprintf("%s-%s", msg.ServiceID, msg.OperationID))
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	var schemaCurr *Schema
	err = json.Unmarshal(entry.Value(), &schemaCurr)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	return schemaCurr, nil
}
