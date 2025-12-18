package job

import (
	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/koss-shtukert/servers-stats/metrics"
	"github.com/rs/zerolog"
)

func ServerMetricsJob(l *zerolog.Logger, c *config.Config) func() {
	return func() {
		logger := l.With().Str("type", "ServerMetricsJob").Logger()
		logger.Debug().Msg("Starting")

		path := c.CronServerDiskUsageJobPath
		if path == "" {
			path = "/"
		}
		result, err := common.GetDiskUsage(&logger, path)
		if err != nil {
			logger.Err(err).Msg("Failed to get disk usage")
			return
		}

		metrics.RecordDiskUsageDetailed(path, "server", result.Percentage, result.UsedBytes, result.AvailBytes)
		logger.Info().Str("used", result.Used).Str("avail", result.Available).Str("usage", result.UsageStr).Int("percentage", result.Percentage).Msg("Server disk usage metrics recorded")
		logger.Debug().Msg("Finished")
	}
}
