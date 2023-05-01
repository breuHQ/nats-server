package core

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"

// 	"go.breu.io/nozl/internal/eventstream"
// 	"go.breu.io/nozl/internal/shared"
// )

// const (
// 	ServiceIDTest string = "AC9f560ea30baaaf8013e4e44284eb6769"
// 	TestMsgMail   string = "test message 123 Mail"
// 	TestMsgNews   string = "test message 123 News"
// 	DestNum       string = "+923214930471"
// )

// var (
// 	st shared.Stack
// )

// func TestMain(m *testing.M) {
// 	st = shared.Stack{}

// 	st.SetupStack()

// 	exitCode := m.Run()

// 	st.TeardownStack()
// 	os.Exit(exitCode)
// }

// // TODO: Complete Test Case.
// func TestCore(t *testing.T) {
// 	shared.InitializeLogger()
// 	eventstream.Eventstream.ReadEnv()
// 	eventstream.Eventstream.InitializeNats()

// 	Core.InitializeCore(shared.MainLimiterRate, shared.MainLimiterBucketSize)

// 	Core.Init()

// 	AUserID := uuid.New().String()
// 	BUserID := uuid.New().String()

// 	bodyMsgMail := eventstream.NewBody(AUserID, TestMsgMail, DestNum)
// 	newMsgMail := eventstream.NewMessage(ServiceIDTest, "", *bodyMsgMail)

// 	bodyMsgNews := eventstream.NewBody(BUserID, TestMsgMail, DestNum)
// 	newMsgNews := eventstream.NewMessage(ServiceIDTest, "", *bodyMsgNews)

// 	Core.Send(newMsgMail)
// 	time.Sleep(1 * time.Second)
// 	Core.Send(newMsgMail)
// 	Core.Send(newMsgNews)
// 	Core.Send(newMsgNews)

// 	time.Sleep(9 * time.Second)
// 	t.Log("Incomplete test case") // TODO: Complete this.
// }

// func TestHealthCheck(t *testing.T) {
// 	ep := st.GetContainerEndpoint("nozl")
// 	resp, err := http.Get(ep + "/healthz")

// 	if err != nil {
// 		shared.Logger.Info(err.Error())
// 		return
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		shared.Logger.Info(err.Error())
// 		return
// 	}

// 	var js map[string]interface{}

// 	_ = json.Unmarshal(body, &js)

// 	defer resp.Body.Close()

// 	assert.Equal(t, "ok", js["status"])
// }
