package schema

import (
	"bytes"
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
	BaseUrl     string
	Path        string
	HttpMethod  string
	PathDetails *openapi3.Operation
	File        SchemaFile
}

func newSchema(pathKey string, httpMethod string, pathDetails *openapi3.Operation, baseUrl string, schemaFile SchemaFile) Schema {
	return Schema{
		BaseUrl:     baseUrl,
		Path:        pathKey,
		HttpMethod:  httpMethod,
		PathDetails: pathDetails,
		File:        schemaFile,
	}
}

func traverseSchemaMapForRefs(pathSchemaProps *openapi3.SchemaRef) {
	items := pathSchemaProps.Value.Items
	properties := pathSchemaProps.Value.Properties

	if items != nil {
		items.Ref = ""
		traverseSchemaMapForRefs(items)
	}

	if pathSchemaProps.Value.Properties != nil {
		for _, prop := range properties {
			prop.Ref = ""
			traverseSchemaMapForRefs(prop)
		}
	}
}

func setSchemaRefToNull(operation *openapi3.Operation) {
	if operation.RequestBody != nil {
		var contentType string
		for key, _ := range operation.RequestBody.Value.Content {
			contentType = key
		}

		reqBodySchema := operation.RequestBody.Value.Content[contentType].Schema

		// Go through request body schema's properties
		// and items recursively, and make there Ref empty
		traverseSchemaMapForRefs(reqBodySchema)
	}
}

func ParseOpenApiV3Schema(serviceID string, specFile []byte, fileName string, updateOperations bool) error {
	doc, err := openapi3.NewLoader().LoadFromData(specFile)

	if err != nil {
		shared.Logger.Error("Failed to ready open API spec!")
		return err
	}

	var schemaFile SchemaFile

	if !updateOperations {
		schemaFileNew, err := AddSchemaFile(serviceID, fileName)
		if err != nil {
			shared.Logger.Error("Failed to add schemaDetails to KV store!")
			return err
		}
		schemaFile = schemaFileNew
	} else {
		schemaFileNew, err := GetSchemaFile(serviceID, fileName)
		if err != nil {
			shared.Logger.Error("Failed to Get schemaDetails from KV store!")
			return err
		}
		schemaFileNew.UpdatedAt = GetDate()
		schemaFile = schemaFileNew

		kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)
		if err != nil {
			shared.Logger.Error("Failed to retreive schemaDetails KV store!")
			return err
		}

		schemaFileJson, err := json.Marshal(schemaFile)
		if err != nil {
			shared.Logger.Error("Failed to marshal schemaDetails!")
			return err
		}

		val, err := kv.Get(schemaFile.FileID)
		if err != nil {
			shared.Logger.Error("Failed to retreive schemaDetails KV store!")
			return err
		}

		revision := val.Revision()
		kv.Update(schemaFile.FileID, schemaFileJson, revision)
	}

	for pathKey, pathValue := range doc.Paths {
		baseUrl := ""

		// If the path doesn't have it's own server URL
		// use other URL specified at start of API spec file
		if pathValue.Servers != nil {
			baseUrl = pathValue.Servers[0].URL
		} else {
			baseUrl = doc.Servers[0].URL
		}
		if pathValue.Get != nil {
			setSchemaRefToNull(pathValue.Get)
			AddSchemaToKVStore(serviceID, pathKey, "GET", pathValue.Get, baseUrl, schemaFile, updateOperations)
		}
		if pathValue.Post != nil {
			setSchemaRefToNull(pathValue.Post)
			AddSchemaToKVStore(serviceID, pathKey, "POST", pathValue.Post, baseUrl, schemaFile, updateOperations)
		}
		if pathValue.Put != nil {
			setSchemaRefToNull(pathValue.Put)
			AddSchemaToKVStore(serviceID, pathKey, "PUT", pathValue.Put, baseUrl, schemaFile, updateOperations)
		}
		if pathValue.Patch != nil {
			setSchemaRefToNull(pathValue.Patch)
			AddSchemaToKVStore(serviceID, pathKey, "PATCH", pathValue.Patch, baseUrl, schemaFile, updateOperations)
		}
		if pathValue.Delete != nil {
			setSchemaRefToNull(pathValue.Delete)
			AddSchemaToKVStore(serviceID, pathKey, "DELETE", pathValue.Delete, baseUrl, schemaFile, updateOperations)
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
			if !stringInSlice(key, RequiredParams) {
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

func AddSchemaToKVStore(serviceID string, pathKey string, httpMethod string, operation *openapi3.Operation, baseUrl string, schemaFile SchemaFile, updateOperations bool) {
	schemaKv, _ := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	MakeOptionalFieldsNullable(operation)
	currSchema := newSchema(pathKey, httpMethod, operation, baseUrl, schemaFile)
	operationID := operation.OperationID
	jsonPayload, _ := json.Marshal(currSchema)

	if updateOperations {
		kv, err := schemaKv.Get(fmt.Sprintf("%s_%s", serviceID, operationID))
		if err != nil {
			shared.Logger.Error(err.Error())
			return
		}

		revision := kv.Revision()
		schemaKv.Update(fmt.Sprintf("%s_%s", serviceID, operationID), jsonPayload, revision)
	} else {
		schemaKv.Put(fmt.Sprintf("%s_%s", serviceID, operationID), jsonPayload)
	}
}

func ValidateOpenAPIV3Schema(msg *eventstream.Message) error {
	schemaValid, err := GetMsgRefSchema(msg)
	if err != nil {
		shared.Logger.Error(err.Error())
		return err
	}

	//if schemaValid.HttpMethod != "GET" && schemaValid.HttpMethod != "DELETE" {
	headers := make(map[string]string)
	if schemaValid.PathDetails.RequestBody != nil {
		for key, _ := range schemaValid.PathDetails.RequestBody.Value.Content {
			headers["Content-Type"] = key
		}
	}

	payload := GetJsonPayloadFromMsg(msg) // TODO: add category based check
	pathParams := GetPathParamsFromMsg(msg)
	queryParams := GetQueryParamsFromMsg(msg)
	httpReq, err := http.NewRequest(schemaValid.HttpMethod, schemaValid.Path, payload)

	for key, val := range headers {
		httpReq.Header.Add(key, val)
	}

	input := &openapi3filter.RequestValidationInput{
		Request:     httpReq,
		PathParams:  pathParams,
		QueryParams: queryParams,
	}

	ctx := context.Background()
	for _, param := range schemaValid.PathDetails.Parameters {
		err = openapi3filter.ValidateParameter(ctx, input, param.Value)
		if err != nil {
			return err
		}
	}

	if schemaValid.PathDetails.RequestBody != nil {
		err = openapi3filter.ValidateRequestBody(ctx, input, schemaValid.PathDetails.RequestBody.Value)
	}

	return err
}

func GetPathParamsFromMsg(msg *eventstream.Message) map[string]string {
	pathParams := make(map[string]string)
	for key, val := range msg.PathParams {
		pathParams[key] = val.(string)
	}
	return pathParams
}

func GetQueryParamsFromMsg(msg *eventstream.Message) url.Values {
	urlValues := url.Values{}
	for key, val := range msg.QueryParams {
		urlValues.Set(key, val.(string))
	}
	return urlValues
}

func GetPayloadFromMsg(msg *eventstream.Message) io.Reader {
	formValues := url.Values{}
	for key, val := range msg.ReqBody {
		formValues.Set(key, val.(string))
	}

	formData := formValues.Encode()
	payload := strings.NewReader(formData)
	return payload
}

func GetJsonPayloadFromMsg(msg *eventstream.Message) io.Reader {
	payload, _ := json.Marshal(msg.ReqBody)
	reader := bytes.NewReader(payload)

	return reader
}

func GetMsgRefSchema(msg *eventstream.Message) (*Schema, error) {
	schemaKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	entry, err := schemaKv.Get(fmt.Sprintf("%s_%s", msg.ServiceID, msg.OperationID))
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
