package job

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"os/exec"

	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/rs/zerolog"
)

func MotioneyeDiskUsageJob(l *zerolog.Logger, c *config.Config, n common.Notifier) func() {
	return func() {
		logger := l.With().Str("type", "MotioneyeDiskUsageJob").Logger()

		logger.Debug().Msg("Starting")

		path := c.CronMotioneyeDiskUsageJobPath

		// Validate and clean path
		cleanPath := filepath.Clean(path)
		if cleanPath == "." || cleanPath == "" {
			logger.Error().Str("path", path).Msg("Invalid path provided")
			n.SendMessage("⚠️ Motioneye: invalid disk path provided")
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Execute command safely
		cmd := exec.CommandContext(ctx, "df", "-h", cleanPath)
		logger.Debug().Str("command", "df -h").Str("path", cleanPath).Msg("Executing disk usage command")

		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		start := time.Now()
		if err := cmd.Run(); err != nil {
			logger.Err(err).Str("stderr", stderr.String()).Str("path", cleanPath).Dur("duration", time.Since(start)).Msg("Failed to execute df")
			n.SendMessage("⚠️ Motioneye: failed to check disk usage")
			return
		}
		logger.Debug().Dur("duration", time.Since(start)).Int("output_size", len(out.String())).Msg("Command executed successfully")

		output := out.String()
		lines := strings.Split(output, "\n")
		logger.Debug().Int("total_lines", len(lines)).Str("target_path", path).Msg("Parsing df output")

		for i, line := range lines {
			if strings.HasSuffix(line, path) {
				logger.Debug().Int("line_number", i).Str("line_content", line).Msg("Found matching line")
				fields := strings.Fields(line)
				if len(fields) >= 5 {
					used := fields[2]
					avail := fields[3]
					usageStr := fields[4]

					percent := 0
					if _, err := fmt.Sscanf(usageStr, "%d%%", &percent); err != nil {
						logger.Err(err).Str("raw", usageStr).Msg("Failed to parse usage percentage")
						// Continue with percent = 0 if parsing fails
					}

					logger.Info().Str("used", used).Str("available", avail).Str("usage", usageStr).Int("percentage", percent).Msg("Disk usage retrieved successfully")
					n.SendMessage(formatDiskUsage(used, avail, usageStr, percent))

					logger.Debug().Msg("Finished")
					return
				}
			}
		}

		// No matching line found in df output
		logger.Warn().Str("path", cleanPath).Int("total_lines", len(lines)).Str("df_output", strings.TrimSpace(output)).Msg("Could not parse disk usage from df output")
		n.SendMessage("⚠️ Motioneye: could not parse disk usage data")
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
		"💾 Motioneye disk usage\n\n"+
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
