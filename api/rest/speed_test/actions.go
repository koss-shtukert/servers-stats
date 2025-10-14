package speed_test

import (
	"net/http"

	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
)

func SpeedTest(l *zerolog.Logger, e *echo.Echo, cfg *config.Config, b *bot.Bot) {
	e.GET("/speed-test", handleSpeedTest(l, cfg, b))
}

func handleSpeedTest(l *zerolog.Logger, cfg *config.Config, b *bot.Bot) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Use bot's rate limiting mechanism to prevent multiple speedtests
		if !b.CanExecuteCommand("speedtest") {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Speedtest is already running or rate limited",
			})
		}

		// Execute speedtest with proper control
		go b.ExecuteJob("speedtest", func() {
			job.SpeedTestJob(l, cfg, b)()
		})

		return c.JSON(http.StatusAccepted, map[string]string{
			"message": "Speedtest started",
		})
	}
}
