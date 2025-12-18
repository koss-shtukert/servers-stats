package job

import (
	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/rs/zerolog"
)

func PlexDiskUsageJob(l *zerolog.Logger, c *config.Config, n common.Notifier) func() {
	return func() {
		logger := l.With().Str("type", "PlexDiskUsageJob").Logger()
		logger.Debug().Msg("Starting")

		result, err := common.GetDiskUsage(&logger, c.CronPlexDiskUsageJobPath)
		if err != nil {
			logger.Err(err).Msg("Failed to get disk usage")
			n.SendMessage("⚠️ Plex: failed to check disk usage")
			return
		}

		logger.Info().Str("used", result.Used).Str("available", result.Available).Str("usage", result.UsageStr).Int("percentage", result.Percentage).Msg("Disk usage retrieved successfully")
		n.SendMessage(common.FormatDiskUsageMessage("Plex", result.Used, result.Available, result.UsageStr, result.Percentage))
		logger.Debug().Msg("Finished")
	}
}
