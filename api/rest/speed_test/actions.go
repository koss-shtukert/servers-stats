package speed_test

import (
	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"
	"net/http"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
)

func SpeedTest(l *zerolog.Logger, e *echo.Echo, cfg *config.Config, b *bot.Bot) {
	e.GET("/speed-test", handleSpeedTest(l, cfg, b))
}

func handleSpeedTest(l *zerolog.Logger, cfg *config.Config, b *bot.Bot) echo.HandlerFunc {
	return func(c echo.Context) error {
		go job.SpeedTestJob(l, cfg, b)()

		return c.JSON(http.StatusOK, "Ok")
	}
}
