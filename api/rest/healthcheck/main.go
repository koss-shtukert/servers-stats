package healthcheck

import (
	"github.com/labstack/echo/v4"
)

func GET(e *echo.Echo) {
	Healthcheck(e)
}
