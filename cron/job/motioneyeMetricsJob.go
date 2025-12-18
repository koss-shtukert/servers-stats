package job

import (
	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/koss-shtukert/servers-stats/metrics"
	"github.com/rs/zerolog"
)

func MotioneyeMetricsJob(l *zerolog.Logger, c *config.Config) func() {
	return func() {
		logger := l.With().Str("type", "MotioneyeMetricsJob").Logger()
		logger.Debug().Msg("Starting")

		result, err := common.GetDiskUsage(&logger, c.CronMotioneyeDiskUsageJobPath)
		if err != nil {
			logger.Err(err).Msg("Failed to get disk usage")
			return
		}

		metrics.RecordDiskUsageDetailed(c.CronMotioneyeDiskUsageJobPath, "motioneye", result.Percentage, result.UsedBytes, result.AvailBytes)
		logger.Info().Str("used", result.Used).Str("avail", result.Available).Str("usage", result.UsageStr).Int("percentage", result.Percentage).Msg("Motioneye disk usage metrics recorded")
		logger.Debug().Msg("Finished")
	}
}
