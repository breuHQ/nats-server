package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
			AddSchemaToKVStore(serviceID, pathKey, "GET", pathValue.Get)
		}
		if pathValue.Post != nil {
			AddSchemaToKVStore(serviceID, pathKey, "POST", pathValue.Post)
		}
		if pathValue.Put != nil {
			AddSchemaToKVStore(serviceID, pathKey, "PUT", pathValue.Put)
		}
		if pathValue.Patch != nil {
			AddSchemaToKVStore(serviceID, pathKey, "PATCH", pathValue.Patch)
		}
		if pathValue.Delete != nil {
			AddSchemaToKVStore(serviceID, pathKey, "DELETE", pathValue.Delete)
		}
	}

	return nil
}

func MakeOptionalFieldsNullable(operation *openapi3.Operation) error {
	if operation.RequestBody != nil {
		// TODO: This needs to be undated to also handle other headers
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

	formValues := url.Values{}
	for key, val := range msg.ReqBody {
		formValues.Set(key, val.(string))
	}

	formData := formValues.Encode()
	payload := strings.NewReader(formData)

	httpReq, err := http.NewRequest(schemaValid.HttpMethod, schemaValid.Path, payload)
	headers := make(map[string]string)
	// TODO: Header should be dynamically set
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	for key, val := range headers {
		httpReq.Header.Add(key, val)
	}

	input := &openapi3filter.RequestValidationInput{
		Request: httpReq,
	}

	ctx := context.Background()
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
