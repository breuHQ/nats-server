package service_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"

// 	"io/ioutil"
// 	"net/http"

// 	"testing"

// 	"github.com/go-faker/faker/v4"
// 	"github.com/stretchr/testify/suite"
// 	"go.breu.io/nozl/internal/service"
// 	"go.breu.io/nozl/internal/shared"
// 	"go.breu.io/nozl/internal/testutils"
// )

// type (
// 	Endpoint struct {
// 		SignUp string
// 		Create string
// 		Get    string
// 		GetAll string
// 		Delete string
// 	}

// 	serviceTestSuite struct {
// 		suite.Suite
// 		request     *service.Service
// 		response    *service.Service
// 		requestStr  *bytes.Reader
// 		endpointAll *Endpoint
// 		token       string
// 	}
// )

// var (
// 	st shared.Stack
// )

// func initTestContainers() {
// 	st = shared.Stack{}
// 	st.SetupStack()

// }

// func (s *serviceTestSuite) SetupSuite() {
// 	initTestContainers()
// 	s.SetupRequestData()
// 	s.PopulateEndpointAll()
// 	s.GetUserToken()
// }


// func (s *serviceTestSuite) PopulateEndpointAll() {
// 	prefix := st.GetContainerEndpoint("nozl")
// 	apiGroup := "dashboard"
// 	s.endpointAll = new(Endpoint)
// 	s.endpointAll.SignUp = fmt.Sprintf("%s/signup", prefix)
// 	s.endpointAll.Create = fmt.Sprintf("%s/%s/service", prefix, apiGroup)
// 	s.endpointAll.Get = fmt.Sprintf("%s/%s/service/%s", prefix, apiGroup, s.request.ID)
// 	s.endpointAll.GetAll = fmt.Sprintf("%s/%s/service", prefix, apiGroup)
// 	s.endpointAll.Delete = fmt.Sprintf("%s/%s/service/%s", prefix, apiGroup, s.request.ID)
// }

// func (s *serviceTestSuite) TearDownSuite() {
// 	st.TeardownStack()
// }

// func (s *serviceTestSuite) GetUserToken() {
// 	s.token, _ = testutils.RegisterUser(s.endpointAll.SignUp)
// }

// func (s *serviceTestSuite) SetupRequestData() {
// 	s.request = service.NewService(faker.Name(), faker.Name(), faker.Name())
// }

// func (s *serviceTestSuite) Test_01_CreateTenant() {
// 	url := s.endpointAll.Create

// 	jsonData, err := json.Marshal(s.request)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)

// 	resp, err := testutils.HTTPRequest(url, "POST", headers, jsonData)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(body, &s.response); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusCreated, resp.StatusCode)
// 	s.Assert().Equal(s.request.Name, s.response.Name)
// 	s.Assert().Equal(s.request.AccountSID, s.response.AccountSID)
// 	s.Assert().Equal(s.request.AuthToken, s.response.AuthToken)
// }

// func (s *serviceTestSuite) Test_02_GetService() {
// 	url := s.endpointAll.Get

// 	resp, err := testutils.HTTPRequest(url, "GET", map[string]string{"Authorization": fmt.Sprintf("Bearer %s", s.token)}, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(body, &s.response); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	s.Assert().Equal(s.request.Name, s.response.Name)
// 	s.Assert().Equal(s.request.AccountSID, s.response.AccountSID)
// 	s.Assert().Equal(s.request.AuthToken, s.response.AuthToken)
// }

// func (s *serviceTestSuite) Test_03_GetAllService() {
// 	url := s.endpointAll.GetAll

// 	resp, err := testutils.HTTPRequest(url, "GET", map[string]string{"Authorization": fmt.Sprintf("Bearer %s", s.token)}, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	var serviceAll []service.Service
// 	if err := json.Unmarshal(body, &serviceAll); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	if len(serviceAll) > 0 {
// 		s.Assert().Equal(s.request.AccountSID, serviceAll[0].AccountSID)
// 		s.Assert().Equal(s.request.Name, serviceAll[0].Name)
// 		s.Assert().Equal(s.request.AuthToken, serviceAll[0].AuthToken)
// 	}

// }

// func (s *serviceTestSuite) Test_04_DeleteService() {
// 	url := s.endpointAll.Delete

// 	resp, err := testutils.HTTPRequest(url, "DELETE", map[string]string{"Authorization": fmt.Sprintf("Bearer %s", s.token)}, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(body, &s.response); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)

// 	url = s.endpointAll.Get
// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)

// 	resp, err = testutils.HTTPRequest(url, "GET", headers, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	defer resp.Body.Close()

// 	s.Assert().Equal(http.StatusNotFound, resp.StatusCode)
// }

// func TestServiceTestSuite(t *testing.T) {
// 	suite.Run(t, new(serviceTestSuite))
// }
