package tenant_test

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"testing"

// 	"github.com/go-faker/faker/v4"
// 	"github.com/stretchr/testify/suite"

// 	"go.breu.io/nozl/internal/shared"
// 	"go.breu.io/nozl/internal/tenant"
// 	"go.breu.io/nozl/internal/testutils"
// )

// type (
// 	RequestData struct {
// 		ID     string `json:"id"`
// 		Name   string `json:"name"`
// 		APIKey string `json:"api_key"`
// 	}

// 	Endpoint struct {
// 		SignUp  string
// 		Create  string
// 		Get     string
// 		GetAll  string
// 		Refresh string
// 		Delete  string
// 	}

// 	ServerHandlerTestSuite struct {
// 		suite.Suite
// 		request     *tenant.Tenant
// 		response    *tenant.Tenant
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

// func (s *ServerHandlerTestSuite) SetupSuite() {
// 	initTestContainers()

// 	s.SetupRequestData()
// 	s.PopulateEndpointAll()
// 	s.GetUserToken()
// }

// func (s *ServerHandlerTestSuite) PopulateEndpointAll() {
// 	prefix := st.GetContainerEndpoint("nozl")
// 	apiGroup := "dashboard"
// 	s.endpointAll = new(Endpoint)
// 	s.endpointAll.SignUp = fmt.Sprintf("%s/signup", prefix)
// 	s.endpointAll.Create = fmt.Sprintf("%s/%s/tenant", prefix, apiGroup)
// 	s.endpointAll.Get = fmt.Sprintf("%s/%s/tenant/%s", prefix, apiGroup, s.request.ID)
// 	s.endpointAll.GetAll = fmt.Sprintf("%s/%s/tenant", prefix, apiGroup)
// 	s.endpointAll.Refresh = fmt.Sprintf("%s/%s/tenant/%s", prefix, apiGroup, s.request.ID)
// 	s.endpointAll.Delete = fmt.Sprintf("%s/%s/tenant/%s", prefix, apiGroup, s.request.ID)
// }

// func (s *ServerHandlerTestSuite) TearDownSuite() {
// 	st.TeardownStack()
// }

// func (s *ServerHandlerTestSuite) TearDownTest() {
// 	s.response = new(tenant.Tenant)
// }

// func (s *ServerHandlerTestSuite) GetUserToken() {
// 	s.token, _ = testutils.RegisterUser(s.endpointAll.SignUp)
// }

// func (s *ServerHandlerTestSuite) SetupRequestData() {
// 	err := faker.FakeData(&s.request)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}
// }

// func (s *ServerHandlerTestSuite) Test_01_CreateTenant() {
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

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &s.response); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusCreated, resp.StatusCode)
// 	s.Assert().Equal(s.request.ID, s.response.ID)
// 	s.Assert().Equal(s.request.Name, s.response.Name)
// 	s.Assert().Equal(s.request.APIKey, s.response.APIKey)
// }

// func (s *ServerHandlerTestSuite) Test_02_GetTenant() {
// 	url := s.endpointAll.Get

// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)

// 	resp, err := testutils.HTTPRequest(url, "GET", headers, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}
// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &s.response); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	s.Assert().Equal(s.request.ID, s.response.ID)
// 	s.Assert().Equal(s.request.Name, s.response.Name)
// 	s.Assert().Equal(s.request.APIKey, s.response.APIKey)

// }

// func (s *ServerHandlerTestSuite) Test_03_GetAllTenant() {
// 	url := s.endpointAll.GetAll

// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)

// 	resp, err := testutils.HTTPRequest(url, "GET", headers, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}
// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	var tenantAll []RequestData
// 	if err := json.Unmarshal(respBody, &tenantAll); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	s.Assert().Equal(s.request.ID, tenantAll[0].ID)
// 	s.Assert().Equal(s.request.Name, tenantAll[0].Name)
// 	s.Assert().Equal(s.request.APIKey, tenantAll[0].APIKey)

// }

// func (s *ServerHandlerTestSuite) Test_04_RefreshTenantAPIKey() {
// 	url := s.endpointAll.Refresh

// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)

// 	resp, err := testutils.HTTPRequest(url, "POST", headers, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}
// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &s.response); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)
// 	s.Assert().Equal(s.request.ID, s.response.ID)
// 	s.Assert().Equal(s.request.Name, s.response.Name)
// 	s.Assert().NotEqual(s.request.APIKey, s.response.APIKey)

// }

// func (s *ServerHandlerTestSuite) Test_05_DeleteTenant() {
// 	url := s.endpointAll.Delete

// 	headers := make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)
// 	resp, err := testutils.HTTPRequest(url, "DELETE", headers, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}
// 	defer resp.Body.Close()

// 	s.Assert().Equal(http.StatusOK, resp.StatusCode)

// 	url = s.endpointAll.Get
// 	headers = make(map[string]string)
// 	headers["Authorization"] = fmt.Sprintf("Bearer %s", s.token)

// 	resp, err = testutils.HTTPRequest(url, "GET", headers, nil)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}
// 	defer resp.Body.Close()

// 	s.Assert().Equal(http.StatusNotFound, resp.StatusCode)

// }

// func TestHandler(t *testing.T) {
// 	suite.Run(t, new(ServerHandlerTestSuite))
// }
