package shared

import (
	"os"
	"time"
)

func getEnvs() map[string]string {
	jwtSecret, _ := os.LookupEnv("JWT_SECRET")

	return map[string]string{
		"jwt_secret": jwtSecret,
	}
}

func GetDate() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

const (
	ServiceKV             string = "Service"
	TenantKV              string = "Tenant"
	TenantAPIKV           string = "TenanatApiKey"
	UserKV                string = "User"
	MsgWaitListKV         string = "MsgWaitList"
	SchemaKV              string = "schema"
	SchemaFileKV          string = "SchemaFile"
	FilterLimiterKV       string = "FilterLimter"
	MainLimiterKV         string = "MainLimiter"
	ConfigKV              string = "Configuration"
	MsgLogKV              string = "MsgLog"
	SenderPhoneNumber     string = "+19034598701"
	UserTokenRate         string = "UserTokenRate"
	UserBucketSize        string = "UserBucketSize"
	MainLimiterRate       string = "MainLimiterRate"
	MainLimiterBucketSize string = "MainLimiterBucketSize"
)

var (
	envs              = getEnvs()
	JWTSecret         = envs["jwt_secret"]
	TokenRateDefault  = "1"
	BucketSizeDefault = "1"
)
