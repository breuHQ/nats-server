package core

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/rate"
	"github.com/nats-io/nats-server/v2/nozl/shared"
	"github.com/nats-io/nats.go"
)

func GetMainLimiterConf(ctx echo.Context) error {
	confMap := eventstream.GetMultValIntKVstore(shared.ConfigKV, []string{shared.MainLimiterRate, shared.MainLimiterBucketSize})
	rateLimit := confMap[shared.MainLimiterRate]
	bucketSize := confMap[shared.MainLimiterBucketSize]

	return ctx.JSON(http.StatusOK, echo.Map{
		"limit":       rateLimit,
		"bucket_size": bucketSize,
	})
}

func SetMainLimiterRate(ctx echo.Context) error {
	payload := make(map[string]string)
	if err := json.NewDecoder(ctx.Request().Body).Decode(&payload); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	limit, err := strconv.Atoi(payload["limit"])
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to convert limit to integer",
		})
	}

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MainLimiterKV)
	if err != nil {
		return err
	}

	allKey, err := kv.Keys()
	if err != nil {
		return err
	}

	for _, key := range allKey {
		updateLimitInLimiter(kv, key, limit)
	}

	kv, err = eventstream.Eventstream.RetreiveKeyValStore(shared.ConfigKV)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	kv.Put(shared.MainLimiterRate, []byte(payload["limit"]))
	kv.Put(shared.MainLimiterBucketSize, []byte(payload["limit"]))

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Main Limit updated",
	})
}

func GetFilterConf(ctx echo.Context) error {
	confMap := eventstream.GetMultValIntKVstore(shared.ConfigKV, []string{shared.UserTokenRate, shared.UserBucketSize})
	rateLimit := confMap[shared.UserTokenRate]
	bucketSize := confMap[shared.UserBucketSize]

	return ctx.JSON(http.StatusOK, echo.Map{
		"limit":       rateLimit,
		"bucket_size": bucketSize,
	})
}

func SetFilterConf(ctx echo.Context) error {
	payload := make(map[string]string)
	if err := json.NewDecoder(ctx.Request().Body).Decode(&payload); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	limit, err := strconv.Atoi(payload["limit"])
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to convert limit to integer",
		})
	}

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ConfigKV)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	kv.Put(shared.UserTokenRate, []byte(payload["limit"]))
	kv.Put(shared.UserBucketSize, []byte(payload["limit"]))

	if err := UpdateCurrFilterConf(limit); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to update previously stored filters",
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Filter configuration updated",
	})
}

func UpdateCurrFilterConf(limit int) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.FilterLimiterKV)
	if err != nil {
		return err
	}

	allKey, err := kv.Keys()
	if err != nil {
		return err
	}

	for _, key := range allKey {
		updateLimitInLimiter(kv, key, limit)
	}

	return nil
}

func updateLimitInLimiter(kv nats.KeyValue, key string, limit int) {
	flRaw, _ := kv.Get(key)

	fl := &rate.Limiter{}
	err := json.Unmarshal(flRaw.Value(), fl)

	fl.SetLimit(rate.Limit(limit))
	fl.SetBurst(limit)

	flUpdated, err := json.Marshal(fl)
	_, err = kv.Put(key, flUpdated)
	if err != nil {
		shared.Logger.Error(err.Error())
	}
}
