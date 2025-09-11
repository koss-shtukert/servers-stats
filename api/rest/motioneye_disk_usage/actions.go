package motioneye_disk_usage

import (
	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"
	"net/http"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
)

func Stats(l *zerolog.Logger, e *echo.Echo, cfg *config.Config, b *bot.Bot) {
	e.GET("/motioneye-disk-usage", handleMotioneyeDiskUsage(l, cfg, b))
}

func handleMotioneyeDiskUsage(l *zerolog.Logger, cfg *config.Config, b *bot.Bot) echo.HandlerFunc {
	return func(c echo.Context) error {
		job.MotioneyeDiskUsageJob(l, cfg, b)()

		return c.JSON(http.StatusOK, "Ok")
	}
}
