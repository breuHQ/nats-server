package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/server/backend/eventstream"
	"github.com/nats-io/nats-server/v2/server/backend/shared"
)

func CreateServiceHandler(ctx echo.Context) error {
	newService := NewService("", "", "")

	if err := ctx.Bind(&newService); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)

	if err != nil {
		if kv, err = eventstream.Eventstream.CreateKeyValStore(shared.ServiceKV, ""); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to create new bucket for service",
			})
		}
	}

	keys, err := kv.Keys()
	if len(keys) == 0 {
		shared.Logger.Info("Key array is empty")
	} else {
		if err != nil {
			shared.Logger.Error(fmt.Sprintf("Key array is empty: %s", err.Error()))

			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to retreive keys from the Key Value store",
			})
		}

		for _, key := range keys {
			value, err := kv.Get(key)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, echo.Map{
					"message": "Unable to retreive value from the Key Value store",
				})
			}

			existingService := new(Service)
			if err := json.Unmarshal(value.Value(), &existingService); err != nil {
				shared.Logger.Error(err.Error())
			}

			if existingService.AccountSID == newService.AccountSID {
				return ctx.JSON(http.StatusConflict, echo.Map{
					"message": "Service already exists",
				})
			}
		}
	}

	jsonPayload, _ := json.Marshal(newService)

	if _, err = kv.Put(newService.ID, jsonPayload); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to save service information",
		})
	}

	return ctx.JSON(http.StatusCreated, newService)
}

func GetServiceHandler(ctx echo.Context) error {
	id := ctx.Param("id")

	if kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV); err == nil {
		if entry, err := kv.Get(id); err == nil {
			payload := new(Service)
			if err := json.Unmarshal(entry.Value(), &payload); err != nil {
				shared.Logger.Error(err.Error())
			}

			return ctx.JSON(http.StatusOK, payload)
		}

		return ctx.JSON(http.StatusNotFound, echo.Map{
			"message": "service not found",
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "bucket not found",
	})
}

func GetServiceAllHandler(ctx echo.Context) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Service Key Value store",
		})
	}

	serviceAllList := []Service{}
	
	allKeys, err := kv.Keys()
	if err != nil {
		return ctx.JSON(http.StatusOK, serviceAllList)
	}


	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to retreive value corresponding to a key from the Key Value store",
			})
		}

		payload := new(Service)
		if err := json.Unmarshal(value.Value(), &payload); err != nil {
			shared.Logger.Error(err.Error())
		}

		serviceAllList = append(serviceAllList, *payload)
	}

	return ctx.JSON(http.StatusOK, serviceAllList)
}

func DeleteServiceHandler(ctx echo.Context) error {
	id := ctx.Param("id")

	if kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV); err == nil {
		if _, err := kv.Get(id); err != nil {
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"message": "Service does not exist.",
			})
		}

		if err := kv.Delete(id); err != nil {
			return ctx.JSON(http.StatusConflict, echo.Map{
				"message": "Error Deleting",
			})
		}

		return ctx.JSON(http.StatusOK, echo.Map{
			"message": fmt.Sprintf("Service deleted with ID: %s", id),
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "bucket not found",
	})
}
