package main

import (
	"context"
	"github.com/koss-shtukert/servers-stats/api"
	"github.com/koss-shtukert/servers-stats/cron"
	"log"
	"os/signal"
	"syscall"

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

	if cfg.CronRunServerDiskUsageJob {
		cronJob.AddServerDiskUsageJob()
	}

	if cfg.CronRunSpeedTestJob {
		cronJob.AddSpeedTestJob()
	}

	s := api.CreateServer(&logr, cfg, tgBot)

	cronJob.Start()
	logr.Info().Str("type", "core").Msg("Cron started")

	tgBot.StartPolling(ctx, &logr, cfg)
	logr.Info().Str("type", "core").Msg("Telegram polling started")

	go s.Start()
	logr.Info().Str("type", "core").Msg("Server started")

	select {}
}
