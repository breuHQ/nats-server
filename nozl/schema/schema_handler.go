package schema

import (
	"bytes"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func UploadOpenApiSpecHandler(ctx echo.Context) error {
	serviceID := ctx.FormValue("service_id")
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

	if err = ParseOpenApiV3Schema(serviceID, buf.Bytes()); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Error in parsing file",
		})
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Open API file parsed successfully",
	})
}

// func DeleteOpenApiV3Schema(serviceID string) error {


// 	return nil
// }

// func DeleteOpenApiSpecHandler(ctx echo.Context) error {
// 	serviceID := ctx.QueryParam("service_id")
// 	if serviceID == "" {
// 		return ctx.JSON(http.StatusBadRequest, echo.Map{
// 			"message": "service_id is required",
// 		})
// 	}

// 	if err := DeleteOpenApiV3Schema(serviceID); err != nil {
// 		return ctx.JSON(http.StatusInternalServerError, echo.Map{
// 			"message": "Error in deleting file",
// 		})
// 	}

// 	return ctx.JSON(http.StatusOK, echo.Map{
// 		"message": "Open API file deleted successfully",
// 	})
// }
