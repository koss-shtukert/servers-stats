package server_disk_usage

import (
	"net/http"

	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
)

func Stats(l *zerolog.Logger, e *echo.Echo, cfg *config.Config, b *bot.Bot) {
	e.GET("/server-disk-usage", handleServerDiskUsage(l, cfg, b))
}

func handleServerDiskUsage(l *zerolog.Logger, cfg *config.Config, b *bot.Bot) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !b.CanExecuteCommand("server_disk_usage") {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Server disk usage check is rate limited",
			})
		}

		go b.ExecuteJob("server_disk_usage", func() {
			job.ServerDiskUsageJob(l, cfg, b)()
		})

		return c.JSON(http.StatusAccepted, map[string]string{
			"message": "Server disk usage check started",
		})
	}
}
