package tenant

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

func CreateTenantHandler(ctx echo.Context) error {
	tnt := NewTenant("")
	if err := ctx.Bind(&tnt); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to find tenant bucket",
		})
	}

	jsonPayload, _ := json.Marshal(tnt)
	if _, err = kv.Put(tnt.GetID(), jsonPayload); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to save tenant information",
		})
	}

	kvAPI, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantAPIKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to find tenant api key bucket",
		})
	}

	if _, err = kvAPI.Put(tnt.GetAPIKey(), []byte(tnt.GetID())); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to save tenant information in API Key bucket",
		})
	}

	return ctx.JSON(http.StatusCreated, tnt)
}

func GetTenantHandler(ctx echo.Context) error {
	tntName := ctx.Param("id")

	if kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV); err == nil {
		if entry, err := kv.Get(tntName); err == nil {
			payload := make(map[string]string)
			if err := json.Unmarshal(entry.Value(), &payload); err != nil {
				shared.Logger.Error(err.Error())
			}

			return ctx.JSON(http.StatusOK, payload)
		}

		return ctx.JSON(http.StatusNotFound, echo.Map{
			"message": "Tenant not found",
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Tenant not found",
	})
}

func GetTenantAllHandler(ctx echo.Context) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Wait Queue Key Value store",
		})
	}

	tenantAllList := []map[string]string{}
	allKeys, err := kv.Keys()
	if err != nil {
		return ctx.JSON(http.StatusOK, tenantAllList)
	}

	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to retreive value corresponding to a key from the Key Value store",
			})
		}

		payload := make(map[string]string)
		if err := json.Unmarshal(value.Value(), &payload); err != nil {
			shared.Logger.Error(err.Error())
		}

		tenantAllList = append(tenantAllList, payload)
	}

	return ctx.JSON(http.StatusOK, tenantAllList)
}

func DeleteTenantHandler(ctx echo.Context) error {
	tntName := ctx.Param("id")

	if kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV); err == nil {
		if _, err := kv.Get(tntName); err != nil {
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"message": "Tenant does not exist",
			})
		}

		if err := kv.Delete(tntName); err != nil {
			return ctx.JSON(http.StatusConflict, echo.Map{
				"message": "Error in deleting the tenant",
			})
		}

		return ctx.JSON(http.StatusOK, echo.Map{
			"message": fmt.Sprintf("Tenant deleted with account SID: %s", tntName),
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Tenant not found",
	})
}

func RefreshAPIKeyHandler(ctx echo.Context) error {
	tntID := ctx.Param("id")

	tenantKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to find tenant bucket",
		})
	}

	tenantAPIKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantAPIKV)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to find tenant api key bucket",
		})
	}

	value, err := tenantKv.Get(tntID)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"message": "Tenant does not exist",
		})
	}

	var tnt Tenant
	if err := json.Unmarshal(value.Value(), &tnt); err != nil {
		shared.Logger.Error(err.Error())
	}

	if err := tenantAPIKv.Delete(tnt.APIKey); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to delete old tenant API token in tenant bucket",
		})
	}

	tnt.APIKey = uuid.New().String()
	jsonPayload, _ := json.Marshal(tnt)

	if _, err = tenantKv.Put(tnt.GetID(), jsonPayload); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to save new tenant token in tenant bucket",
		})
	}

	if _, err = tenantAPIKv.Put(tnt.GetAPIKey(), []byte(tnt.GetID())); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to save tenant information in tenant API Key bucket",
		})
	}

	return ctx.JSON(http.StatusOK, tnt)
}

func DeleteTenantAllHandler(ctx echo.Context) error {
	tenantKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to find tenant bucket",
		})
	}

	tenantAPIKv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantAPIKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to find tenant api key bucket",
		})
	}

	allKeys, err := tenantKv.Keys()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive keys from the Key Value store",
		})
	}

	for _, key := range allKeys {
		value, err := tenantKv.Get(key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to retreive value corresponding to a key from the Key Value store",
			})
		}

		var tnt Tenant
		if err := json.Unmarshal(value.Value(), &tnt); err != nil {
			shared.Logger.Error(err.Error())
		}

		if err := tenantAPIKv.Delete(tnt.APIKey); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to delete old tenant API token in tenant bucket",
			})
		}

		if err := tenantKv.Delete(key); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to delete old tenant API token in tenant bucket",
			})
		}
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "All tenants deleted",
	})
}