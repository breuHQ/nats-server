package schema

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

func serviceIDExists(serviceID string) bool {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)
	if err != nil {
		shared.Logger.Error("Failed to retreive KV store!")
		return false
	}

	if _, err := kv.Get(serviceID); err != nil {
		shared.Logger.Error("Failed to retreive value from KV store!")
		return false
	}

	return true
}

func UploadOpenApiSpecHandler(ctx echo.Context) error {
	serviceID := ctx.FormValue("service_id")
	file, err := ctx.FormFile("file")

	fileName := file.Filename

	if err != nil || serviceID == "" || fileName == "" {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "File Upload Error",
		})
	}

	if !serviceIDExists(serviceID) {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "Service ID does not exist",
		})
	}

	openApiFile, err := file.Open()

	if err != nil {
		return ctx.JSON(http.StatusExpectationFailed, echo.Map{
			"message": "File Opening Error",
		})
	}
	defer openApiFile.Close()

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, openApiFile); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Error in parsing file",
		})
	}

	if err = ParseOpenApiV3Schema(serviceID, buf.Bytes(), fileName); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Error in parsing file",
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Open API file parsed successfully",
	})
}

func DeleteSchemaFile(fileID string) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)
	if err != nil {
		shared.Logger.Error("Failed to retreive KV store!")
		return err
	}

	_, err = kv.Get(fileID)
	if err != nil {
		shared.Logger.Error("Failed to retreive KV pair or the key does not exist!")
		return err
	}

	err = kv.Delete(fileID)
	if err != nil {
		shared.Logger.Error("Failed to Delete KV pair")
		return err
	}

	return nil

}

func DeleteStoredOperations(serviceID string, fileID string) error {

	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaKV)
	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return err
	}

	allKeys, err := kv.Keys()
	if err != nil {
		shared.Logger.Error("Failed to retreive schemaDetails KV store!")
		return err
	}

	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			shared.Logger.Error("Failed to retreive value from KV store!")
			return err
		}

		schemaVal := new(Schema)
		if err := json.Unmarshal(value.Value(), &schemaVal); err != nil {
			shared.Logger.Error("Failed to unmarshal schemaDetails from KV store!")
			return err
		}

		if schemaVal.File.FileID == fileID && schemaVal.File.ServiceID == serviceID {
			kv.Delete(key)
		}

	}
	return nil
}

func DeleteOpenApiSpecHandler(ctx echo.Context) error {
	serviceID := ctx.QueryParam("service_id")
	fileID := ctx.QueryParam("file_id")
	if serviceID == "" {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "service_id is required",
		})
	}

	if fileID == "" {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "file_id is required",
		})
	}

	if err := DeleteStoredOperations(serviceID, fileID); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Error in deleting openapi spec",
		})
	}

	if err := DeleteSchemaFile(fileID); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Error in deleting schema file",
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Open API file deleted successfully",
	})
}

func GetAllOpenApiSpecHandler(ctx echo.Context) error {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.SchemaFileKV)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Error in retreiving schema files",
		})
	}

	openapiFileList := []SchemaFile{}
	allKeys, err := kv.Keys()

	// If there are no keys, return empty list
	// This might not be the best way to handle this
	if err != nil {
		return ctx.JSON(http.StatusOK, openapiFileList)
	}

	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Error in retreiving schema files",
			})
		}

		schemaFile := new(SchemaFile)
		if err := json.Unmarshal(value.Value(), &schemaFile); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Error in retreiving schema files",
			})
		}

		openapiFileList = append(openapiFileList, *schemaFile)
	}
	return ctx.JSON(http.StatusOK, openapiFileList)
}
