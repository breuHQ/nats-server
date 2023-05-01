package auth

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"testing"

// 	"github.com/go-faker/faker/v4"
// 	"github.com/stretchr/testify/suite"

// 	"go.breu.io/nozl/internal/shared"
// 	"go.breu.io/nozl/internal/testutils"
// )

// var (
// 	st shared.Stack
// )

// type (
// 	AuthTestSuite struct {
// 		suite.Suite
// 		username string
// 		password string
// 	}
// )

// func (ath *AuthTestSuite) SetupSuite() {
// 	st = shared.Stack{}

// 	st.SetupStack()
// 	ath.username = faker.Username()
// 	ath.password = faker.Password()
// }

// func (ath *AuthTestSuite) TearDownSuite() {
// 	st.TeardownStack()
// }

// func (ath *AuthTestSuite) Test_01_Signup() {
// 	url := fmt.Sprintf("%s/signup", st.GetContainerEndpoint("nozl"))
// 	requestPayload := make(map[string]string)
// 	requestPayload["username"] = ath.username
// 	requestPayload["password"] = ath.password
// 	headers := make(map[string]string)
	
// 	jsonBody, err := json.Marshal(requestPayload)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	response, err := testutils.HTTPRequest(url, "POST", headers, jsonBody)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	var js map[string]string

// 	defer response.Body.Close()

// 	respBody, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &js); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	ath.Assert().Equal(http.StatusOK, response.StatusCode)
// 	ath.Assert().NotEqual("", js["token"])
// }

// func (ath *AuthTestSuite) Test_02_Login() {
// 	url := fmt.Sprintf("%s/login", st.GetContainerEndpoint("nozl"))
// 	requestPayload := make(map[string]string)
// 	requestPayload["username"] = ath.username
// 	requestPayload["password"] = ath.password
// 	headers := make(map[string]string)
	
// 	jsonBody, err := json.Marshal(requestPayload)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	response, err := testutils.HTTPRequest(url, "POST", headers, jsonBody)

// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 		return
// 	}

// 	var js map[string]string

// 	defer response.Body.Close()

// 	respBody, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	if err := json.Unmarshal(respBody, &js); err != nil {
// 		shared.Logger.Error(err.Error())
// 	}

// 	ath.Assert().Equal(http.StatusOK, response.StatusCode)
// 	ath.Assert().NotEqual("", js["token"])
// }

// func TestHandler(t *testing.T) {
// 	suite.Run(t, new(AuthTestSuite))
// }