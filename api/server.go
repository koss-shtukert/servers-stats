package api

import (
	"github.com/koss-shtukert/servers-stats/api/rest/motioneye_disk_usage"
	"github.com/koss-shtukert/servers-stats/api/rest/speed_test"
	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type Server struct {
	server *echo.Echo
	tgBot  *bot.Bot
	logger *zerolog.Logger
	config *config.Config
}

func CreateServer(l *zerolog.Logger, c *config.Config, b *bot.Bot) *Server {
	logger := l.With().Str("type", "server").Logger()

	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log := logger.Debug()

			if v.Error != nil {
				log = logger.Error()
			}

			log.
				Str("host", v.Host).
				Str("uri", v.URI).
				Str("method", v.Method).
				Int("status", v.Status).
				Any("headers", v.Headers).
				Str("remote_ip", v.RemoteIP).
				Str("request_id", v.RequestID)

			if v.Error == nil {
				log.Msg("request")
			} else {
				log.Msg(v.Error.Error())
			}

			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"GET"},
	}))

	e.HideBanner = true

	motioneye_disk_usage.REST(l, e, c, b)
	speed_test.REST(l, e, c, b)

	s := &Server{
		server: e,
		tgBot:  b,
		logger: &logger,
		config: c,
	}

	return s
}

func (s *Server) Start() {
	s.logger.Err(s.server.Start(":1324"))
}
