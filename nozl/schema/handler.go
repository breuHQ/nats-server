package schema

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SpecUploadHandler(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "File uploaded successfully",
	})
}
