package core

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/server/backend/rate"
	"github.com/nats-io/nats-server/v2/server/backend/shared"
)

func GetMainLimiterRate(ctx echo.Context) error {
	rateLimit := Core.MainLimiter.Limit()

	return ctx.JSON(http.StatusOK, echo.Map{
		"limit": rateLimit,
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

	Core.MainLimiter.SetLimit(rate.Limit(limit))
	Core.MainLimiter.SetBurst(limit)
	shared.MainLimiterRate = limit
	shared.MainLimiterBucketSize = limit

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Main Limit updated",
	})
}

func GetFilterConf(ctx echo.Context) error {
	rateLimit := shared.UserTokenRate
	bucketSize := shared.UserBucketSize

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

	shared.UserTokenRate = limit
	shared.UserBucketSize = limit

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
	for _, fil := range Core.Filters {
		fil.SetFilterLimit(limit)
		fil.SetFilterBucketSize(limit)
	}

	return nil
}
