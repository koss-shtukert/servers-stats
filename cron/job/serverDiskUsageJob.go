package job

import (
	"bytes"
	"fmt"
	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/rs/zerolog"
	"os/exec"
	"strings"
)

func ServerDiskUsageJob(l *zerolog.Logger, c *config.Config, n common.Notifier) func() {
	return func() {
		logger := l.With().Str("type", "ServerDiskUsageJob").Logger()

		logger.Debug().Msg("Starting")

		path := c.CronServerDiskUsageJobPath

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
				if len(fields) < 5 {
					continue
				}

				used := fields[len(fields)-4]
				avail := fields[len(fields)-3]
				usageStr := fields[len(fields)-2]

				percent := 0
				if _, err := fmt.Sscanf(usageStr, "%d%%", &percent); err != nil {
					l.Err(err).Str("raw", usageStr).Msg("Failed to parse usage percentage")
				}

				n.SendMessage(formatServerDiskUsage(used, avail, usageStr, percent))
				logger.Debug().Msg("Finished")
				return
			}
		}

		logger.Warn().Msgf("Could not parse disk usage from df output on path %s", path)
	}
}

func formatServerDiskUsage(used, avail, usageStr string, percent int) string {
	status := "🟢 OK"
	if percent >= 90 {
		status = "🔴 CRITICAL"
	} else if percent >= 70 {
		status = "🟡 Warning"
	}

	return fmt.Sprintf(
		"💾 Server disk usage\n\n"+
			"📊 Used:    %s\n"+
			"📦 Avail:   %s\n"+
			"📈 Usage:   %s\n"+
			"✅ Status:  %s",
		used,
		avail,
		usageStr,
		status,
	)
}
