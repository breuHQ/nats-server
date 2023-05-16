package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	File        SchemaFile
}

func addSchema(pathKey string, httpMethod string, pathDetails *openapi3.Operation, schemaFile SchemaFile) Schema {
	return Schema{
		Path:        pathKey,
		HttpMethod:  httpMethod,
		PathDetails: pathDetails,
		File:        schemaFile,
	}
}

func ParseOpenApiV3Schema(serviceID string, specFile []byte, fileName string) error {
	doc, err := openapi3.NewLoader().LoadFromData(specFile)

	if err != nil {
		shared.Logger.Error("Failed to ready open API spec!")
		return err
	}

	schemaFile, err := AddSchemaFiletoKVStore(serviceID, fileName)
	if err != nil {
		shared.Logger.Error("Failed to add schemaDetails to KV store!")
		return err
	}

	for pathKey, pathValue := range doc.Paths {

		if pathValue.Get != nil {
			AddSchemaToKVStore(serviceID, pathKey, "GET", pathValue.Get, schemaFile)
		}
		if pathValue.Post != nil {
			AddSchemaToKVStore(serviceID, pathKey, "POST", pathValue.Post, schemaFile)
		}
		if pathValue.Put != nil {
			AddSchemaToKVStore(serviceID, pathKey, "PUT", pathValue.Put, schemaFile)
		}
		if pathValue.Patch != nil {
			AddSchemaToKVStore(serviceID, pathKey, "PATCH", pathValue.Patch, schemaFile)
		}
		if pathValue.Delete != nil {
			AddSchemaToKVStore(serviceID, pathKey, "DELETE", pathValue.Delete, schemaFile)
		}
	}
	return nil
}

func MakeOptionalFieldsNullable(operation *openapi3.Operation) error {
	if operation.RequestBody != nil {
		var contentType string
		for key, _ := range operation.RequestBody.Value.Content {
			contentType = key
		}

		ReqBodySchema := operation.RequestBody.Value.Content[contentType].Schema.Value
		RequiredParams := ReqBodySchema.Required
		ReqBodyParams := ReqBodySchema.Properties

		for key, val := range ReqBodyParams {
			MakeRefsEmpty(val) // This is being done because when marshalling SchemaRef, it only marshals field "Ref"
			if stringInSlice(key, RequiredParams) == false {
				val.Value.Nullable = true
			}
		}
	}

	return nil
}

func MakeRefsEmpty(schemaRef *openapi3.SchemaRef) error {
	schemaRef.Ref = ""
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

func AddSchemaToKVStore(serviceID string, pathKey string, httpMethod string, operation *openapi3.Operation, schemaFile SchemaFile) {
	schemaKv, _ := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	MakeOptionalFieldsNullable(operation)
	currSchema := addSchema(pathKey, httpMethod, operation, schemaFile)
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

	headers := make(map[string]string)
	for key, _ := range schemaValid.PathDetails.RequestBody.Value.Content {
		headers["Content-Type"] = key
	}

	payload := GetPayloadFromMsg(msg, headers["Content-Type"])
	httpReq, err := http.NewRequest(schemaValid.HttpMethod, schemaValid.Path, payload)

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

func GetPayloadFromMsg(msg *eventstream.Message, contentType string) io.Reader {
	formValues := url.Values{}
	for key, val := range msg.ReqBody {
		formValues.Set(key, val.(string))
	}

	formData := formValues.Encode()
	payload := strings.NewReader(formData)
	return payload
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
