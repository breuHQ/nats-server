package testutils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/go-faker/faker/v4"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/schema"
	"github.com/nats-io/nats-server/v2/nozl/service"
	"github.com/nats-io/nats-server/v2/nozl/shared"
	"github.com/nats-io/nats-server/v2/nozl/tenant"
	"github.com/nats-io/nats-server/v2/nozl/user"
)

func genFakeUserData() (*user.User, error) {

	var u *user.User
	err := faker.FakeData(&u)

	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	return u, nil
}

// Creates and saves a user in data store and
// returns an Authorization Token
func RegisterUser(url string) (string, error) {
	payload, err := genFakeUserData()

	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headers := make(map[string]string)

	resp, err := HTTPRequest(url, "POST", headers, jsonData)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := map[string]string{}
	if err := json.Unmarshal(respBody, &body); err != nil {
		return "", err
	}
	return body["token"], nil
}

// Creates and saves a tenant with fake name in data store and
// returns a Tenant struct
func RegisterTenant(url string, authToken string) *tenant.Tenant {
	tnt := tenant.NewTenant(faker.Name())

	jsonData, err := json.Marshal(tnt)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", authToken)
	resp, err := HTTPRequest(url, "POST", headers, jsonData)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	var js *tenant.Tenant

	if err := json.Unmarshal(respBody, &js); err != nil {
		shared.Logger.Error(err.Error())
	}

	return js
}

func RegisterService(url string, authToken string) *service.Service {
	svcName:= "twilio"
	svcAccountSID := "ACe4c8ac5725c5c02c75aec71f53cc69e4"
	svcAuthToken := "3ffa94991aa2a5c0246800cd1f1a5616"
	svc := service.NewService(svcName, svcAccountSID, svcAuthToken)
	jsonData, err := json.Marshal(svc)

	if err != nil {
		shared.Logger.Error(err.Error())
		return nil
	}

	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", authToken) 
	resp, err := HTTPRequest(url, "POST", headers, jsonData)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		shared.Logger.Error(err.Error())
	}

	var js *service.Service

	if err := json.Unmarshal(respBody, &js); err != nil {
		shared.Logger.Error(err.Error())
	}

	return js
}

func SetFilterLimit(url string, authToken string, filterLimit int) {
	requestPayload := struct {
		Limit string `json:"limit"`
	} {
		Limit: fmt.Sprintf("%d", filterLimit),
	}
	jsonData, _ := json.Marshal(&requestPayload)
	headers := make(map[string]string)

	headers["Authorization"] = fmt.Sprintf("Bearer %s", authToken) 

	_, err := HTTPRequest(url, "POST", headers, jsonData)

	if err != nil {
		shared.Logger.Error(err.Error())
		return
	}

}

func HTTPRequest(url string, method string, headers map[string]string, jsonData []byte) (*http.Response, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	headers["Content-Type"] = "application/json"

	for key, val := range headers {
		request.Header.Add(key, val)
	}

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		shared.Logger.Error(err.Error())
		return nil, err
	}

	return response, nil
}

func GenericHTTPRequest(svc *service.Service, msg *eventstream.Message) {
	sid := svc.AccountSID
	authToken := svc.AuthToken
	authKey := basicAuth(sid, authToken)
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

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}