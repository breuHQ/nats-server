package nozl

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats-server/v2/nozl/auth"
	"github.com/nats-io/nats-server/v2/nozl/core"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/schema"
	"github.com/nats-io/nats-server/v2/nozl/service"
	"github.com/nats-io/nats-server/v2/nozl/shared"
	"github.com/nats-io/nats-server/v2/nozl/tenant"
)

func PreSetupNozl(natsPort int) {
	shared.InitializeLogger()
	eventstream.Eventstream.SetupUrl(natsPort)
	eventstream.Eventstream.InitializeNats()
}

func SetupNozl(backendPort string) {
	// Hack around ingress-gce not supporting rewrite-target.
	// LINK: https://app.shortcut.com/breu/story/2859/helm-chart-for-nozl#activity-2871
	prefix := getRoutingPrefix()

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover()) // NOTE: This will prevent echo from crashing if a panic occurs.

	e.GET(prefix+"/healthz", healthz)
	e.POST(prefix+"/signup", auth.SignUpHandler)
	e.POST(prefix+"/login", auth.LoginHandler)

	tenantGroup := e.Group(prefix + "/sdk")
	{
		tenantGroup.Use(middleware.KeyAuth(auth.GetKeyAuthConfig))

		tenantGroup.POST("/message", core.SendMessageHandler)
	}

	backendAPIGroup := e.Group(prefix + "/dashboard")
	{
		backendAPIGroup.Use(echojwt.WithConfig(auth.GetConfig()))

		backendAPIGroup.GET("/limiter", core.GetMainLimiterConf)
		backendAPIGroup.POST("/limiter", core.SetMainLimiterRate)

		backendAPIGroup.GET("/filter", core.GetFilterConf)
		backendAPIGroup.POST("/filter", core.SetFilterConf)

		backendAPIGroup.POST("/service", service.CreateServiceHandler)
		backendAPIGroup.GET("/service", service.GetServiceAllHandler)
		backendAPIGroup.GET("/service/:id", service.GetServiceHandler)
		backendAPIGroup.DELETE("/service/:id", service.DeleteServiceHandler)

		backendAPIGroup.POST("/tenant", tenant.CreateTenantHandler)
		backendAPIGroup.GET("/tenant", tenant.GetTenantAllHandler)
		backendAPIGroup.GET("/tenant/:id", tenant.GetTenantHandler)
		backendAPIGroup.POST("/tenant/:id", tenant.RefreshAPIKeyHandler)
		backendAPIGroup.DELETE("/tenant/:id", tenant.DeleteTenantHandler)
		backendAPIGroup.DELETE("/tenant", tenant.DeleteAllTenantsHandler)

		backendAPIGroup.GET("/message", core.GetAllMsgWaitListHandler)
		backendAPIGroup.GET("/message/log", core.GetMsgLogHandler)
		backendAPIGroup.GET("/message/:message_id", core.ForceSendMsgHandler)
		backendAPIGroup.DELETE("/message/:message_id", core.DeleteMsgHandler)
		backendAPIGroup.DELETE("/message/log", core.DeleteMsgLogHandler)

		backendAPIGroup.POST("/schema/upload", schema.UploadOpenApiSpecHandler)
		backendAPIGroup.DELETE("/schema/:file_id", schema.DeleteOpenApiSpecHandler)
		backendAPIGroup.DELETE("/schema", schema.DeleteAllOpenApiSpecHandler)
		backendAPIGroup.GET("/schema", schema.GetAllOpenApiSpecHandler)

		backendAPIGroup.GET("", auth.RestrictedHandler)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", backendPort)))
}

// TODO: make sure that connection to all services it needs to connect to are working properly.
func healthz(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func getRoutingPrefix() string {
	prefix, ok := os.LookupEnv("ROUTING_PREFIX")
	if !ok {
		prefix = ""
	}

	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	if prefix != "" && strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}

	return prefix
}
