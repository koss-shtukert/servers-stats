package job

import (
	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/rs/zerolog"
)

func MotioneyeDiskUsageJob(l *zerolog.Logger, c *config.Config, n common.Notifier) func() {
	return func() {
		logger := l.With().Str("type", "MotioneyeDiskUsageJob").Logger()
		logger.Debug().Msg("Starting")

		result, err := common.GetDiskUsage(&logger, c.CronMotioneyeDiskUsageJobPath)
		if err != nil {
			logger.Err(err).Msg("Failed to get disk usage")
			n.SendMessage("⚠️ Motioneye: failed to check disk usage")
			return
		}

		logger.Info().Str("used", result.Used).Str("available", result.Available).Str("usage", result.UsageStr).Int("percentage", result.Percentage).Msg("Disk usage retrieved successfully")
		n.SendMessage(common.FormatDiskUsageMessage("Motioneye", result.Used, result.Available, result.UsageStr, result.Percentage))
		logger.Debug().Msg("Finished")
	}
}
