package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/schema"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

type HTTPClient interface {
	GenericHTTPRequest(svc *Service, msg *eventstream.Message) ([]byte, error)
}

type TwilioHTTP struct{}
type SendGridHTTP struct{}
type VonageHTTP struct{}

// TODO: Extract common logic from all 3 Generic HTTP methods

func (sg *SendGridHTTP) GenericHTTPRequest(svc *Service, msg *eventstream.Message) ([]byte, error) {
	schemaValid, err := schema.GetMsgRefSchema(msg)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	ep := schemaValid.BaseUrl + schemaValid.Path
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + svc.AuthDetails["api_key"]
	var payload io.Reader

	if schemaValid.PathDetails.RequestBody != nil {
		for key, _ := range schemaValid.PathDetails.RequestBody.Value.Content {
			headers["Content-Type"] = key
		}
		payload = schema.GetJsonPayloadFromMsg(msg)
	} else {
		payload = bytes.NewBuffer([]byte{})
	}

	httpReq, err := http.NewRequest(schemaValid.HttpMethod, ep, payload)

	for key, val := range headers {
		httpReq.Header.Add(key, val)
	}

	client := &http.Client{}

	response, err := client.Do(httpReq)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	var js map[string]interface{}

	if err := json.Unmarshal(respBody, &js); err != nil {
		shared.Logger.Error(err.Error())
	}

	shared.Logger.Info(response.Status)
	shared.Logger.Info(string(respBody))

	serviceResponse := make(map[string]interface{})
	serviceResponse["status"] = response.Status

	for key, val := range js {
		serviceResponse[key] = val
	}

	return json.Marshal(serviceResponse)
}

func (t *TwilioHTTP) GenericHTTPRequest(svc *Service, msg *eventstream.Message) ([]byte, error) {
	sid := svc.AuthDetails["account_sid"]
	authToken := svc.AuthDetails["auth_token"]
	authKey := t.basicAuth(sid, authToken)
	rgx := regexp.MustCompile("{AccountSid}")

	schemaValid, err := schema.GetMsgRefSchema(msg)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	updatedPath := rgx.ReplaceAllString(schemaValid.Path, sid)
	ep := schemaValid.BaseUrl + updatedPath

	headers := make(map[string]string)
	headers["Authorization"] = "Basic " + authKey
	var payload io.Reader

	if schemaValid.PathDetails.RequestBody != nil {
		for key, _ := range schemaValid.PathDetails.RequestBody.Value.Content {
			headers["Content-Type"] = key
		}
		payload = schema.GetXURLFormEncodedPayloadFromMsg(msg)
	} else {
		payload = bytes.NewBuffer([]byte{})
	}

	httpReq, err := http.NewRequest(schemaValid.HttpMethod, ep, payload)

	for key, val := range headers {
		httpReq.Header.Add(key, val)
	}

	client := &http.Client{}

	response, err := client.Do(httpReq)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	var js map[string]interface{}

	if err := json.Unmarshal(respBody, &js); err != nil {
		shared.Logger.Error(err.Error())
	}

	shared.Logger.Info(response.Status)
	shared.Logger.Info(string(respBody))

	serviceResponse := make(map[string]interface{})
	serviceResponse["status"] = response.Status

	for key, val := range js {
		serviceResponse[key] = val
	}

	return json.Marshal(serviceResponse)
}

func (t *TwilioHTTP) basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (v *VonageHTTP) GenericHTTPRequest(svc *Service, msg *eventstream.Message) ([]byte, error) {
	schemaValid, err := schema.GetMsgRefSchema(msg)
	rgx := regexp.MustCompile("{format}")
	updatedPath := rgx.ReplaceAllString(schemaValid.Path, "json") // TODO: get this value from schema directly

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	ep := schemaValid.BaseUrl + updatedPath
	headers := make(map[string]string)
	var payload io.Reader

	if schemaValid.PathDetails.RequestBody != nil {
		for key, _ := range schemaValid.PathDetails.RequestBody.Value.Content {
			headers["Content-Type"] = key
		}
		payload = schema.GetXURLFormEncodedPayloadFromMsg(msg)
	} else {
		payload = bytes.NewBuffer([]byte{})
	}

	httpReq, err := http.NewRequest(schemaValid.HttpMethod, ep, payload)

	for key, val := range headers {
		httpReq.Header.Add(key, val)
	}

	client := &http.Client{}

	response, err := client.Do(httpReq)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	var js map[string]interface{}

	if err := json.Unmarshal(respBody, &js); err != nil {
		shared.Logger.Error(err.Error())
	}

	shared.Logger.Info(response.Status)
	shared.Logger.Info(string(respBody))

	serviceResponse := make(map[string]interface{})
	serviceResponse["status"] = response.Status

	for key, val := range js {
		serviceResponse[key] = val
	}

	return json.Marshal(serviceResponse)
}
