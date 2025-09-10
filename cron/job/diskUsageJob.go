package job

import (
	"bytes"
	"fmt"
	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/rs/zerolog"
	"os/exec"
	"strings"
)

func DiskUsageJob(l *zerolog.Logger, c *config.Config, b *bot.Bot) func() {
	return func() {
		logger := l.With().Str("type", "DiskUsageJob").Logger()

		logger.Debug().Msg("Starting")

		path := "/host" + c.CronDiskUsageJobPath

		cmd := exec.Command("sh", "-c", "df -h "+path)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			logger.Err(err).Str("stderr", stderr.String()).Msg("Failed to execute df")
			return
		}

		for _, line := range strings.Split(out.String(), "\n") {
			if strings.HasSuffix(line, path) {
				fields := strings.Fields(line)
				if len(fields) >= 5 {
					used := fields[2]
					avail := fields[3]
					usageStr := fields[4]

					percent := 0
					if _, err := fmt.Sscanf(usageStr, "%d%%", &percent); err != nil {
						l.Err(err).Str("raw", usageStr).Msg("Failed to parse usage percentage")
					}

					b.SendMessage(formatDiskUsage(used, avail, usageStr, percent))

					logger.Debug().Msg("Finished")
					return
				}
			}
		}

		logger.Warn().Msgf("Could not parse disk usage from df output on path %s", path)
	}
}

func formatDiskUsage(used, avail, usageStr string, percent int) string {
	status := "🟢 OK"
	if percent >= 90 {
		status = "🔴 CRITICAL"
	} else if percent >= 70 {
		status = "🟡 Warning"
	}

	return fmt.Sprintf(
		"💾 Motioneye Disk Usage\n\n"+
			"📊 Used:    %s\n"+
			"📦 Avail:   %s\n"+
			"📈 Usage:   %s\n"+
			"✅ Status:  %s",
		used, avail, usageStr, status,
	)
}
