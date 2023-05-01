package eventstream_test

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/suite"

// 	"go.breu.io/nozl/internal/eventstream"
// 	"go.breu.io/nozl/internal/service"
// 	"go.breu.io/nozl/internal/shared"
// 	"go.breu.io/nozl/internal/testutils"
// )

// var (
// 	st shared.Stack
// )

// type (
// 	EventstreamTestSuite struct {
// 		suite.Suite
// 		token  string
// 		apiKey string
// 		svc    *service.Service
// 	}
// )

// func (ets *EventstreamTestSuite) SetupSuite() {
// 	st = shared.Stack{}

// 	st.SetupStack()
// 	ets.GenerateData()
// }

// func (ets *EventstreamTestSuite) TearDownSuite() {
// 	st.TeardownStack()
// }

// func (ets *EventstreamTestSuite) GenerateData() {
// 	url := st.GetContainerEndpoint("nozl")
// 	token, err := testutils.RegisterUser(fmt.Sprintf("%s/signup", url))

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	ets.token = token
// 	tnt := testutils.RegisterTenant(fmt.Sprintf("%s/dashboard/tenant", url), token)
// 	ets.apiKey = tnt.APIKey

// 	svc := testutils.RegisterService(fmt.Sprintf("%s/dashboard/service", url), token)
// 	ets.svc = svc
// }

// func (ets *EventstreamTestSuite) Test_00_SendSuccessfulTwilioSms() {
// 	url := fmt.Sprintf("%s/sdk/message", st.GetContainerEndpoint("nozl"))
// 	svcID := ets.svc.ID
// 	operationID := ""
// 	destination := "+3214930471"
// 	userID := uuid.New().String()

// 	body := eventstream.NewBody(userID, "unit test message", destination)
// 	msg := eventstream.NewMessage(svcID, operationID, *body)
// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", ets.apiKey)
// 	jsonData, err := json.Marshal(msg)

// 	var js map[string]string

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	resp, err := testutils.HTTPRequest(url, "POST", headers, jsonData)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &js); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	ets.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	ets.Assert().Equal(fmt.Sprintf("message with ID %s allowed by filter limiter", msg.ID), js["message"])

// }

// func (ets *EventstreamTestSuite) Test_01_RejectedTwilioSms() {
// 	url := fmt.Sprintf("%s/sdk/message", st.GetContainerEndpoint("nozl"))
// 	svcID := ets.svc.ID
// 	operationID := ""
// 	destination := "+3214930471"
// 	userID := uuid.New().String()
// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", ets.apiKey)

// 	for idx := 0; idx < 2; idx++ {
// 		flag := "allowed"

// 		if idx%2 == 1 {
// 			flag = "rejected"
// 		}

// 		body := eventstream.NewBody(userID, fmt.Sprintf("unit test message %d", idx), destination)
// 		msg := eventstream.NewMessage(svcID, operationID, *body)
// 		jsonData, err := json.Marshal(msg)

// 		var js map[string]string

// 		if err != nil {
// 			shared.Logger.Error(err.Error())
// 			return
// 		}

// 		resp, err := testutils.HTTPRequest(url, "POST", headers, jsonData)

// 		if err != nil {
// 			shared.Logger.Error(err.Error())
// 		}

// 		defer resp.Body.Close()

// 		respBody, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			shared.Logger.Error(err.Error())
// 		}

// 		if err := json.Unmarshal(respBody, &js); err != nil {
// 			shared.Logger.Error(err.Error())
// 		}

// 		ets.Assert().Equal(http.StatusOK, resp.StatusCode)
// 		ets.Assert().Equal(fmt.Sprintf("message with ID %s %s by filter limiter", msg.ID, flag), js["message"])
// 	}
// }

// func run(ets *EventstreamTestSuite, filterLimit int, userID string, flag string) {
// 	url := fmt.Sprintf("%s/sdk/message", st.GetContainerEndpoint("nozl"))
// 	limitUrl := fmt.Sprintf("%s/dashboard/filter", st.GetContainerEndpoint("nozl"))
// 	svcID := ets.svc.ID
// 	operationID := ""
// 	destination := "+3214930471"
// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", ets.apiKey)

// 	testutils.SetFilterLimit(limitUrl, ets.token, filterLimit)

// 	body := eventstream.NewBody(userID, fmt.Sprintf("unit test message"), destination)
// 	msg := eventstream.NewMessage(svcID, operationID, *body)
// 	jsonData, err := json.Marshal(msg)

// 	var js map[string]string

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	resp, err := testutils.HTTPRequest(url, "POST", headers, jsonData)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &js); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	ets.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	ets.Assert().Equal(fmt.Sprintf("message with ID %s %s by filter limiter", msg.ID, flag), js["message"])
// }

// func runLoop(ets *EventstreamTestSuite, loopLim int, filLim int, userID string, flag string) {
// 	for i := 0; i < loopLim; i++ {
// 		run(ets, filLim, userID, flag)
// 	}
// }

// func (ets *EventstreamTestSuite) Test_02_LoadTestTwilioSmsApiSingleUser() {

// 	load := []struct {
// 		filterLimit int
// 		msgCount    int
// 		userID      string
// 	}{
// 		{
// 			filterLimit: 2,
// 			msgCount:    3,
// 			userID:      uuid.New().String(),
// 		},
// 		{
// 			filterLimit: 4,
// 			msgCount:    6,
// 			userID:      uuid.New().String(),
// 		}, {
// 			filterLimit: 7,
// 			msgCount:    10,
// 			userID:      uuid.New().String(),
// 		},
// 	}

// 	runLoop(ets, 2, load[0].filterLimit, load[0].userID, "allowed")
// 	runLoop(ets, 2, load[0].filterLimit, load[0].userID, "rejected")

// 	runLoop(ets, 4, load[1].filterLimit, load[1].userID, "allowed")
// 	runLoop(ets, 2, load[1].filterLimit, load[1].userID, "rejected")

// 	runLoop(ets, 7, load[2].filterLimit, load[2].userID, "allowed")
// 	runLoop(ets, 2, load[2].filterLimit, load[2].userID, "rejected")
// }

// func (ets *EventstreamTestSuite) Test_03_LoadTestTwilioSmsApiMultipleUser() {

// 	load := []struct {
// 		filterLimit int
// 		msgCount    int
// 		user1ID     string
// 		user2ID     string
// 		user3ID     string
// 	}{
// 		{
// 			filterLimit: 2,
// 			msgCount:    3,
// 			user1ID:     uuid.New().String(),
// 			user2ID:     uuid.New().String(),
// 			user3ID:     uuid.New().String(),
// 		},
// 		{
// 			filterLimit: 4,
// 			msgCount:    6,
// 			user1ID:     uuid.New().String(),
// 			user2ID:     uuid.New().String(),
// 			user3ID:     uuid.New().String(),
// 		}, {
// 			filterLimit: 7,
// 			msgCount:    10,
// 			user1ID:     uuid.New().String(),
// 			user2ID:     uuid.New().String(),
// 			user3ID:     uuid.New().String(),
// 		},
// 	}

// 	runLoop(ets, 2, load[0].filterLimit, load[0].user1ID, "allowed")
// 	runLoop(ets, 2, load[0].filterLimit, load[0].user1ID, "rejected")
// 	runLoop(ets, 2, load[0].filterLimit, load[0].user2ID, "allowed")
// 	runLoop(ets, 2, load[0].filterLimit, load[0].user2ID, "rejected")
// 	runLoop(ets, 2, load[0].filterLimit, load[0].user3ID, "allowed")
// 	runLoop(ets, 2, load[0].filterLimit, load[0].user3ID, "rejected")

// 	runLoop(ets, 4, load[1].filterLimit, load[1].user1ID, "allowed")
// 	runLoop(ets, 2, load[1].filterLimit, load[1].user1ID, "rejected")
// 	runLoop(ets, 4, load[1].filterLimit, load[1].user2ID, "allowed")
// 	runLoop(ets, 2, load[1].filterLimit, load[1].user2ID, "rejected")
// 	runLoop(ets, 4, load[1].filterLimit, load[1].user3ID, "allowed")
// 	runLoop(ets, 2, load[1].filterLimit, load[1].user3ID, "rejected")

// 	runLoop(ets, 7, load[2].filterLimit, load[2].user1ID, "allowed")
// 	runLoop(ets, 2, load[2].filterLimit, load[2].user1ID, "rejected")
// 	runLoop(ets, 7, load[2].filterLimit, load[2].user2ID, "allowed")
// 	runLoop(ets, 2, load[2].filterLimit, load[2].user2ID, "rejected")
// 	runLoop(ets, 7, load[2].filterLimit, load[2].user3ID, "allowed")
// 	runLoop(ets, 2, load[2].filterLimit, load[2].user3ID, "rejected")
// }

// func TestHandler(t *testing.T) {
// 	suite.Run(t, new(EventstreamTestSuite))
// }
