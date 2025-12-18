package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/koss-shtukert/servers-stats/api"
	"github.com/koss-shtukert/servers-stats/cron"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/koss-shtukert/servers-stats/logger"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal("Config error: ", err)
	}

	logr, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatal("Logger error: ", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tgBot, err := bot.CreateBot(cfg, &logr)
	if err != nil {
		log.Fatal("Telegram bot error: ", err)
	}

	cronJob := cron.NewCron(&logr, cfg, tgBot)

	if cfg.CronRunMotioneyeDiskUsageJob {
		cronJob.AddMotioneyeDiskUsageJob()
	}

	if cfg.CronRunPlexDiskUsageJob {
		cronJob.AddPlexDiskUsageJob()
	}

	if cfg.CronRunServerDiskUsageJob {
		cronJob.AddServerDiskUsageJob()
	}

	if cfg.CronRunSpeedTestJob {
		cronJob.AddSpeedTestJob()
	}

	if cfg.CronRunMotioneyeMetricsJob {
		cronJob.AddMotioneyeMetricsJob()
	}

	if cfg.CronRunServerMetricsJob {
		cronJob.AddServerMetricsJob()
	}

	if cfg.CronRunPlexMetricsJob {
		cronJob.AddPlexMetricsJob()
	}

	s := api.CreateServer(&logr, cfg, tgBot)

	cronJob.Start()
	logr.Info().Str("type", "core").Msg("Cron started")

	tgBot.StartPolling(ctx, &logr, cfg)
	logr.Info().Str("type", "core").Msg("Telegram polling started")

	go func() {
		if err := s.Start(); err != nil {
			logr.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
	logr.Info().Str("type", "core").Msg("Server started")

	<-ctx.Done()
	logr.Info().Str("type", "core").Msg("Shutdown signal received")

	// Graceful shutdown
	if err := s.Shutdown(); err != nil {
		logr.Error().Err(err).Msg("Error during server shutdown")
	}

	logr.Info().Str("type", "core").Msg("Application shutdown complete")
}
