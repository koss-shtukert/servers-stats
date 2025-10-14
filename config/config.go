package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment                       string  `mapstructure:"app_env"`
	LogLevel                          string  `mapstructure:"log_level"`
	CronRunMotioneyeDiskUsageJob      bool    `mapstructure:"cron_run_motioneye_disk_usage_job"`
	CronMotioneyeDiskUsageJobPath     string  `mapstructure:"cron_motioneye_disk_usage_job_path"`
	CronMotioneyeDiskUsageJobInterval string  `mapstructure:"cron_motioneye_disk_usage_job_interval"`
	CronRunServerDiskUsageJob         bool    `mapstructure:"cron_run_server_disk_usage_job"`
	CronServerDiskUsageJobPath        string  `mapstructure:"cron_server_disk_usage_job_path"`
	CronServerDiskUsageJobInterval    string  `mapstructure:"cron_server_disk_usage_job_interval"`
	CronRunSpeedTestJob               bool    `mapstructure:"cron_run_speed_test_job"`
	CronSpeedTestJobInterval          string  `mapstructure:"cron_speed_test_job_interval"`
	CronSpeedTestJobExpDown           float64 `mapstructure:"cron_speed_test_job_exp_down"`
	CronSpeedTestJobExpUp             float64 `mapstructure:"cron_speed_test_job_exp_up"`
	CronSpeedTestJobWarnPct           float64 `mapstructure:"cron_speed_test_job_warn_pct"`
	CronSpeedTestJobCritPct           float64 `mapstructure:"cron_speed_test_job_crit_pct"`
	CronSpeedTestJobWarnLat           float64 `mapstructure:"cron_speed_test_job_warn_lat"`
	CronSpeedTestJobCritLat           float64 `mapstructure:"cron_speed_test_job_crit_lat"`
	TgBotApiKey                       string  `mapstructure:"tgbot_api_key"`
	TgBotChatId                       string  `mapstructure:"tgbot_chat_id"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// Enable environment variables support
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("SERVERS_STATS")

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

	// Set defaults
	v.SetDefault("cron_run_disk_usage_job", false)
	v.SetDefault("cron_disk_usage_job_path", "")
	v.SetDefault("cron_disk_usage_job_interval", "")
	v.SetDefault("cron_run_speedtest_job", false)
	v.SetDefault("cron_speedtest_job_interval", "")

	// Bind environment variables for sensitive data
	v.BindEnv("tgbot_api_key", "TGBOT_API_KEY")
	v.BindEnv("tgbot_chat_id", "TGBOT_CHAT_ID")

	// Try to read config file, but don't fail if it doesn't exist
	// when using environment variables
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with env vars and defaults
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if strings.TrimSpace(cfg.CronMotioneyeDiskUsageJobPath) == "" {
		cfg.CronRunMotioneyeDiskUsageJob = false
	}

	if strings.TrimSpace(cfg.CronServerDiskUsageJobPath) == "" {
		cfg.CronRunServerDiskUsageJob = false
	}

	// Validate required fields
	required := map[string]string{
		"app_env":       cfg.Environment,
		"log_level":     cfg.LogLevel,
		"tgbot_api_key": cfg.TgBotApiKey,
		"tgbot_chat_id": cfg.TgBotChatId,
	}
	for k, v := range required {
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("required config %s is empty (set via config file or environment variable)", k)
		}
	}

	// Validate API key format (basic check)
	if !strings.Contains(cfg.TgBotApiKey, ":") {
		return nil, fmt.Errorf("invalid telegram bot API key format")
	}

	return &cfg, nil
}
