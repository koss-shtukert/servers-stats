package common

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type DiskUsageResult struct {
	Used       string
	Available  string
	UsageStr   string
	Percentage int
	UsedBytes  float64
	AvailBytes float64
}

func GetDiskUsage(logger *zerolog.Logger, path string) (*DiskUsageResult, error) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == "" {
		return nil, fmt.Errorf("invalid path provided: %s", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "df", "-h", cleanPath)
	logger.Debug().Str("command", "df -h").Str("path", cleanPath).Msg("Executing disk usage command")

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Run(); err != nil {
		logger.Err(err).Str("stderr", stderr.String()).Str("path", cleanPath).Dur("duration", time.Since(start)).Msg("Failed to execute df")
		return nil, fmt.Errorf("failed to execute df: %w", err)
	}
	logger.Debug().Dur("duration", time.Since(start)).Int("output_size", len(out.String())).Msg("Command executed successfully")

	output := out.String()
	lines := strings.Split(output, "\n")
	logger.Debug().Int("total_lines", len(lines)).Str("target_path", path).Msg("Parsing df output")

	for i, line := range lines {
		if strings.HasSuffix(line, path) {
			logger.Debug().Int("line_number", i).Str("line_content", line).Msg("Found matching line")
			fields := strings.Fields(line)
			
			// Handle multi-line df output where filesystem name is on separate line
			if len(fields) < 6 && len(fields) >= 5 {
				// Line starts with whitespace, fields are: size, used, avail, use%, mount
				used := fields[1]
				avail := fields[2]
				usageStr := fields[3]
				
				percent := 0
				if _, err := fmt.Sscanf(usageStr, "%d%%", &percent); err != nil {
					logger.Err(err).Str("raw", usageStr).Msg("Failed to parse usage percentage")
					return nil, fmt.Errorf("failed to parse usage percentage: %w", err)
				}
				
				usedBytes, err := ParseDiskSize(used)
				if err != nil {
					logger.Err(err).Str("used", used).Msg("Failed to parse used size")
					return nil, fmt.Errorf("failed to parse used size: %w", err)
				}
				
				availBytes, err := ParseDiskSize(avail)
				if err != nil {
					logger.Err(err).Str("avail", avail).Msg("Failed to parse available size")
					return nil, fmt.Errorf("failed to parse available size: %w", err)
				}
				
				return &DiskUsageResult{
					Used:       used,
					Available:  avail,
					UsageStr:   usageStr,
					Percentage: percent,
					UsedBytes:  usedBytes,
					AvailBytes: availBytes,
				}, nil
			} else if len(fields) >= 6 {
				// Single line format: filesystem, size, used, avail, use%, mount
				used := fields[len(fields)-4]
				avail := fields[len(fields)-3]
				usageStr := fields[len(fields)-2]

				percent := 0
				if _, err := fmt.Sscanf(usageStr, "%d%%", &percent); err != nil {
					logger.Err(err).Str("raw", usageStr).Msg("Failed to parse usage percentage")
					return nil, fmt.Errorf("failed to parse usage percentage: %w", err)
				}

				usedBytes, err := ParseDiskSize(used)
				if err != nil {
					logger.Err(err).Str("used", used).Msg("Failed to parse used size")
					return nil, fmt.Errorf("failed to parse used size: %w", err)
				}

				availBytes, err := ParseDiskSize(avail)
				if err != nil {
					logger.Err(err).Str("avail", avail).Msg("Failed to parse available size")
					return nil, fmt.Errorf("failed to parse available size: %w", err)
				}

					return &DiskUsageResult{
						Used:       used,
						Available:  avail,
						UsageStr:   usageStr,
						Percentage: percent,
						UsedBytes:  usedBytes,
						AvailBytes: availBytes,
					}, nil
				}
			}
		}
	}

	logger.Warn().Str("path", cleanPath).Int("total_lines", len(lines)).Str("df_output", strings.TrimSpace(output)).Msg("Could not parse disk usage from df output")
	return nil, fmt.Errorf("could not parse disk usage from df output")
}

func FormatDiskUsageMessage(serviceName, used, avail, usageStr string, percent int) string {
	status := "🟢 OK"
	if percent >= 90 {
		status = "🔴 CRITICAL"
	} else if percent >= 70 {
		status = "🟡 Warning"
	}

	return fmt.Sprintf(
		"💾 %s disk usage\n\n"+
			"📊 Used:    %s\n"+
			"📦 Avail:   %s\n"+
			"📈 Usage:   %s\n"+
			"✅ Status:  %s",
		serviceName,
		used,
		avail,
		usageStr,
		status,
	)
}
