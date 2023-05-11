package service

import (
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
	GenericHTTPRequest(svc *Service, msg *eventstream.Message)
}

type Twilio struct {}

func (t *Twilio) GenericHTTPRequest(svc *Service, msg *eventstream.Message) {
	sid := svc.AccountSID
	authToken := svc.AuthToken
	authKey := t.basicAuth(sid, authToken)
	rgx := regexp.MustCompile("{AccountSid}")

	schemaValid, err := schema.GetMsgRefSchema(msg)
	updatedPath := rgx.ReplaceAllString(schemaValid.Path, sid)
	ep := "https://api.twilio.com" + updatedPath

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Basic " + authKey

	for key, _ := range schemaValid.PathDetails.RequestBody.Value.Content {
		headers["Content-Type"] = key
	}

	payload := schema.GetPayloadFromMsg(msg, headers["Content-Type"])
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


	var js map[string]string

	if err := json.Unmarshal(respBody, &js); err != nil {
		shared.Logger.Error(err.Error())
	}

	shared.Logger.Info(response.Status)
	shared.Logger.Info(string(respBody))
}

func (t *Twilio) basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}