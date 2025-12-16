package plex_disk_usage

import (
	"net/http"

	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
)

func Stats(l *zerolog.Logger, e *echo.Echo, cfg *config.Config, b *bot.Bot) {
	e.GET("/plex-disk-usage", handlePlexDiskUsage(l, cfg, b))
}

func handlePlexDiskUsage(l *zerolog.Logger, cfg *config.Config, b *bot.Bot) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !b.CanExecuteCommand("plex_disk_usage") {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Plex disk usage check is rate limited",
			})
		}

		go b.ExecuteJob("plex_disk_usage", func() {
			job.PlexDiskUsageJob(l, cfg, b)()
		})

		return c.JSON(http.StatusAccepted, map[string]string{
			"message": "Plex disk usage check started",
		})
	}
}
