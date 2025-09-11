package motioneye_disk_usage

import (
	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func REST(l *zerolog.Logger, e *echo.Echo, c *config.Config, b *bot.Bot) {
	Stats(l, e, c, b)
}
