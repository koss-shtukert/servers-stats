package main

import (
	"github.com/koss-shtukert/servers-stats/cron"
	"log"

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

	tgBot, err := bot.CreateBot(cfg.TgBotApiKey, cfg.TgBotChatId, &logr)
	if err != nil {
		log.Fatal("Telegram bot error: ", err)
	}

	cronJob := cron.NewCron(&logr, cfg, tgBot)

	if cfg.CronRunDiskUsageJob {
		cronJob.AddDiskUsageJob()
	}

	if cfg.CronRunSpeedTestJob {
		cronJob.AddSpeedTestJob()
	}

	cronJob.Start()

	logr.Info().Str("type", "core").Msg("Cron started")

	select {}
}
