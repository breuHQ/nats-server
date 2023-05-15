package schema

import (
	"bytes"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func UploadOpenApiSpecHandler(ctx echo.Context) error {
	serviceID := ctx.FormValue("service_id")
	fileName := ctx.FormValue("file_name")
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "File Upload Error",
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

// TODO: Edit these functions below

// func GetOpenApiV3Schema(serviceID string, fileID string) (string, error) {
// 	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.ServiceKV)
// 	if err != nil {
// 		return "", err
// 	}

// 	// TODO: GET the Operations from the KV store
// 	// Use wildcards serviceID.FileID.*  to get the operations
// 	key := fmt.Sprintf("%s-%s", serviceID, fileID)
// 	value, err := kv.Get(key)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(value.Value()), nil
// }

// func GetOpenApiSpecHandler(ctx echo.Context) error {
// 	serviceID := ctx.QueryParam("service_id")
// 	fileID := ctx.QueryParam("file_id")
// 	if serviceID == "" {
// 		return ctx.JSON(http.StatusBadRequest, echo.Map{
// 			"message": "service_id is required",
// 		})
// 	}
// 	if fileID == "" {
// 		return ctx.JSON(http.StatusBadRequest, echo.Map{
// 			"message": "file_id is required",
// 		})
// 	}

// 	openApiSpec, err := GetOpenApiV3Schema(serviceID, fileID)
// 	if err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, echo.Map{
// 			"message": "Error in retreiving file",
// 		})
// 	}

// 	return ctx.JSON(http.StatusOK, echo.Map{
// 		"message": "Open API file retreived successfully",
// 		"file":    openApiSpec,
// 	})

// }

// func DeleteOpenApiV3Schema(serviceID string, fileID string) error {

// 	// TODO: Delete the file from the KV store
// 	// Use wildcards serviceID.FileID.*  to get the operations
// 	// Delete each individual kv pair (serviceID.FileID.OperationID)
// 	return nil
// }

// func DeleteOpenApiSpecHandler(ctx echo.Context) error {
// 	serviceID := ctx.QueryParam("service_id")
// 	fileID := ctx.QueryParam("file_id")
// 	if serviceID == "" {
// 		return ctx.JSON(http.StatusBadRequest, echo.Map{
// 			"message": "service_id is required",
// 		})
// 	}
// 	if fileID == "" {
// 		return ctx.JSON(http.StatusBadRequest, echo.Map{
// 			"message": "file_id is required",
// 		})
// 	}

// 	if err := DeleteOpenApiV3Schema(serviceID, fileID); err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, echo.Map{
// 			"message": "Error in deleting file",
// 		})
// 	}

// 	return ctx.JSON(http.StatusOK, echo.Map{
// 		"message": "Open API file deleted successfully",
// 	})
// }
