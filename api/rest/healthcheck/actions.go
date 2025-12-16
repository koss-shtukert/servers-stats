package healthcheck

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Healthcheck(e *echo.Echo) {
	e.GET("/healthcheck", handleHealthcheck())
}

func handleHealthcheck() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Servers stats server",
		})
	}
}
