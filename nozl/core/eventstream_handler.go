package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

func SendMessageHandler(ctx echo.Context) error {
	var reqBody eventstream.ReqBody
	var pathParams eventstream.PathParams
	var queryParams eventstream.QueryParams
	msg := eventstream.NewMessage("", "", reqBody, pathParams, queryParams)

	if err := ctx.Bind(&msg); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	apiKey := strings.TrimPrefix(ctx.Request().Header.Get("Authorization"), "Bearer ")
	if err := IdentifyTenant(msg, apiKey); err != nil {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"message": "Unable to identify tenant",
		})
	}

	//eventstream.Eventstream.PublishEncodedMessage("Filter", msg)
	go Core.handleFilter(msg)

	msgFilterAllow := <-eventstream.MessageFilterAllow

	if msgFilterAllow.Allow {
		serviceResponse := <-eventstream.ServiceResponse

		var js map[string]interface{}

		_ = json.Unmarshal(serviceResponse, &js)

		return ctx.JSON(http.StatusOK, echo.Map{
			"message_id":    msg.ID,
			"filter_status": msgFilterAllow,
			"response_body": js,
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message_id":    msg.ID,
		"filter_status": msgFilterAllow,
	})
}

func ForceSendMsgHandler(ctx echo.Context) error {
	msgID := ctx.Param("message_id")

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MsgWaitListKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Wait Queue Key Value store",
		})
	}

	msgEntry, err := kv.Get(msgID)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"message": "Unable to find message ID",
		})
	}

	msg := &eventstream.Message{}

	err = json.Unmarshal(msgEntry.Value(), &msg)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"message": "Unable to unmarshal msg",
		})
	}

	//eventstream.Eventstream.PublishEncodedMessage("MainLimiter", msg)
	Core.handleFilter(msg)

	if err = kv.Delete(msgID); err != nil {
		shared.Logger.Info("Unable to delete sent message")
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Messaged published successfully",
	})
}

func GetAllMsgWaitListHandler(ctx echo.Context) error {
	err := retrieveAllValKVStore(shared.MsgWaitListKV, ctx)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Key Value store",
		})
	}

	return nil
}

func GetMsgLogHandler(ctx echo.Context) error {
	err := retrieveAllValKVStore(shared.MsgLogKV, ctx)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Key Value store",
		})
	}

	return err
}

func retrieveAllValKVStore(store string, ctx echo.Context) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(store)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Key Value store",
		})
	}

	allKeys, err := kv.Keys()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive keys from the Key Value store",
		})
	}

	vals := []*eventstream.Message{}

	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to retreive value corresponding to a key from the Key Value store",
			})
		}

		msg := new(eventstream.Message)
		if err := json.Unmarshal(value.Value(), &msg); err != nil {
			shared.Logger.Error(err.Error())
		}

		vals = append(vals, msg)
	}

	return ctx.JSON(http.StatusOK, vals)
}

func DeleteMsgHandler(ctx echo.Context) error {
	msgID := ctx.Param("message_id")

	if kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MsgWaitListKV); err == nil {
		if _, err := kv.Get(msgID); err != nil {
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"message": "Message does not exist",
			})
		}

		if err := kv.Delete(msgID); err != nil {
			return ctx.JSON(http.StatusConflict, echo.Map{
				"message": "Error in deleting message",
			})
		}

		return ctx.JSON(http.StatusOK, echo.Map{
			"message": fmt.Sprintf("Message deleted with ID: %s", msgID),
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Message not found",
	})
}

func DeleteMsgLogHandler(ctx echo.Context) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.MsgLogKV)
	if err != nil {
		return ctx.JSON(http.StatusConflict, echo.Map{
			"message": "Error in retreiving message log",
		})
	}

	allKeys, err := kv.Keys()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive keys from the Key Value store",
		})
	}

	for _, key := range allKeys {
		if err := kv.Delete(key); err != nil {
			return ctx.JSON(http.StatusConflict, echo.Map{
				"message": "Error in deleting message",
				"key":     key,
			})
		}
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "All messages deleted in log",
	})
}

func IdentifyTenant(msg *eventstream.Message, apiKey string) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantAPIKV)
	if err != nil {
		return err
	}

	value, err := kv.Get(apiKey)
	if err != nil {
		return err
	}

	msg.TenantID = string(value.Value())

	return nil
}
