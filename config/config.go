package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment              string `mapstructure:"app_env"`
	LogLevel                 string `mapstructure:"log_level"`
	CronRunDiskUsageJob      bool   `mapstructure:"cron_run_disk_usage_job"`
	CronDiskUsageJobPath     string `mapstructure:"cron_disk_usage_job_path"`
	CronDiskUsageJobInterval string `mapstructure:"cron_disk_usage_job_interval"`
	CronRunSpeedTestJob      bool   `mapstructure:"cron_run_speed_test_job"`
	CronSpeedTestJobInterval string `mapstructure:"cron_speed_test_job_interval"`
	TgBotApiKey              string `mapstructure:"tgbot_api_key"`
	TgBotChatId              string `mapstructure:"tgbot_chat_id"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	switch {
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		v.SetConfigFile(path)
	default:
		if path != "" {
			v.AddConfigPath(path)
		}
		v.AddConfigPath(".")
		v.SetConfigName("config")
	}

	v.SetDefault("cron_run_disk_usage_job", false)
	v.SetDefault("cron_disk_usage_job_path", "")
	v.SetDefault("cron_disk_usage_job_interval", "")

	v.SetDefault("cron_run_speedtest_job", false)
	v.SetDefault("cron_speedtest_job_interval", "")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if strings.TrimSpace(cfg.CronDiskUsageJobPath) == "" {
		cfg.CronRunDiskUsageJob = false
	}

	required := map[string]string{
		"app_env":       cfg.Environment,
		"log_level":     cfg.LogLevel,
		"tgbot_api_key": cfg.TgBotApiKey,
		"tgbot_chat_id": cfg.TgBotChatId,
	}
	for k, v := range required {
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("required config %s is empty", k)
		}
	}

	return &cfg, nil
}
