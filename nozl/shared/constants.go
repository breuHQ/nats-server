package shared

import (
	"os"
)

func getEnvs() map[string]string {
	jwtSecret, _ := os.LookupEnv("JWT_SECRET")

	return map[string]string{
		"jwt_secret": jwtSecret,
	}
}

const (
	ServiceKV         string = "Service"
	TenantKV          string = "Tenant"
	TenantAPIKV       string = "TenanatApiKey"
	UserKV            string = "User"
	MsgWaitListKV     string = "MsgWaitList"
	MsgLogKV          string = "MsgLog"
	SenderPhoneNumber string = "+19034598701"
)

var (
	envs                  = getEnvs()
	JWTSecret             = envs["jwt_secret"]
	UserTokenRate         = 1
	UserBucketSize        = 1
	MainLimiterRate       = 1
	MainLimiterBucketSize = 1
)
