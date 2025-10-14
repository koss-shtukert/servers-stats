package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"

	"github.com/rs/zerolog"
	"github.com/showwin/speedtest-go/speedtest"
)

func SpeedTestJob(l *zerolog.Logger, c *config.Config, n common.Notifier) func() {
	return func() {
		logger := l.With().Str("type", "SpeedTestJob").Logger()
		start := time.Now()
		logger.Info().Time("start_time", start).Msg("SpeedTest job started")

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		defer func() {
			duration := time.Since(start)
			logger.Info().Dur("duration", duration).Msg("SpeedTest job completed")
		}()

		speedtest.WithUserConfig(&speedtest.UserConfig{
			Debug:      true,
			SavingMode: true,
		})

		ctxUser, cancelUser := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelUser()

		userCh := make(chan *speedtest.User, 1)
		errCh := make(chan error, 1)

		go func() {
			logger.Debug().Msg("Starting FetchUserInfo")
			start := time.Now()
			u, err := speedtest.FetchUserInfo()
			logger.Debug().Dur("fetch_user_duration", time.Since(start)).Msg("FetchUserInfo completed")
			if err != nil {
				errCh <- err
				return
			}
			userCh <- u
		}()

		var user *speedtest.User
		select {
		case <-ctxUser.Done():
			logger.Error().Msg("FetchUserInfo timeout")
			n.SendMessage("⚠️ Speedtest: timeout під час отримання інформації про мережу")
			return
		case err := <-errCh:
			logger.Err(err).Msg("FetchUserInfo failed")
			n.SendMessage("⚠️ Speedtest: не вдалося отримати інформацію про мережу")
			return
		case u := <-userCh:
			user = u
		}

		serversCh := make(chan speedtest.Servers, 1)
		go func() {
			logger.Debug().Msg("Starting FetchServers")
			start := time.Now()
			servers, err := speedtest.FetchServers()
			logger.Debug().Dur("fetch_servers_duration", time.Since(start)).Msg("FetchServers completed")
			if err != nil {
				errCh <- err
				return
			}
			serversCh <- servers
		}()

		var servers speedtest.Servers
		select {
		case <-ctx.Done():
			logger.Error().Msg("FetchServers timeout")
			n.SendMessage("⚠️ Speedtest: timeout під час отримання серверів")
			return
		case err := <-errCh:
			logger.Err(err).Msg("FetchServers failed")
			n.SendMessage("⚠️ Speedtest: не вдалося отримати список серверів")
			return
		case servers = <-serversCh:
		}

		var targets speedtest.Servers
		targets, err := servers.FindServer([]int{})
		if err != nil || len(targets) == 0 {
			if err == nil {
				err = fmt.Errorf("no server found")
			}
			logger.Err(err).Msg("FindServer failed")
			n.SendMessage("⚠️ Speedtest: не знайдено підходящого сервера")
			return
		}
		s := targets[0]
		logger.Info().Str("server_name", s.Name).Str("server_country", s.Country).Str("server_id", s.ID).Msg("Selected speedtest server")

		logger.Debug().Msg("Starting PingTest")
		if err := runWithTimeout(ctx, logger, "PingTest", func() error {
			return s.PingTest(nil)
		}); err != nil {
			logger.Err(err).Msg("PingTest failed")
			n.SendMessage("⚠️ Speedtest: помилка під час PingTest")
			return
		}

		logger.Debug().Msg("Starting DownloadTest")
		if err := runWithTimeout(ctx, logger, "DownloadTest", func() error {
			return s.DownloadTest()
		}); err != nil {
			logger.Err(err).Msg("DownloadTest failed")
			n.SendMessage("⚠️ Speedtest: помилка під час DownloadTest")
			return
		}

		logger.Debug().Msg("Starting UploadTest")
		if err := runWithTimeout(ctx, logger, "UploadTest", func() error {
			return s.UploadTest()
		}); err != nil {
			logger.Err(err).Msg("UploadTest failed")
			n.SendMessage("⚠️ Speedtest: помилка під час UploadTest")
			return
		}

		msg := formatSpeedtest(user, s, c)
		n.SendMessage(msg)

		logger.Debug().Msg("Finished")
	}
}

func formatSpeedtest(user *speedtest.User, s *speedtest.Server, c *config.Config) string {
	dlMbps := s.DLSpeed.Mbps()
	ulMbps := s.ULSpeed.Mbps()
	pingMs := float64(s.Latency) / float64(time.Millisecond)

	status := speedStatus(dlMbps, ulMbps, pingMs, c)

	isp := strings.TrimSpace(user.Isp)
	if isp == "" {
		isp = "n/a"
	}

	return fmt.Sprintf(
		"🚀 Ookla Speedtest\n\n"+
			"⬇️ Download: %.2f MB/s\n"+
			"⬆️ Upload:   %.2f MB/s\n"+
			"🕒 Ping:     %.1f ms\n"+
			"🏷 ISP:      %s\n"+
			"🗺 Server:   %s — %s (%s) • ID %s\n"+
			"✅ Status:   %s",
		dlMbps,
		ulMbps,
		pingMs,
		isp,
		s.Name, s.Country, s.Sponsor, s.ID,
		status,
	)
}

func speedStatus(dlMbps, ulMbps, pingMs float64, c *config.Config) string {
	downRate := dlMbps / c.CronSpeedTestJobExpDown
	upRate := ulMbps / c.CronSpeedTestJobExpUp

	if downRate < c.CronSpeedTestJobCritPct || upRate < c.CronSpeedTestJobCritPct || pingMs > c.CronSpeedTestJobCritLat {
		return "🔴 Poor"
	}
	if downRate < c.CronSpeedTestJobWarnPct || upRate < c.CronSpeedTestJobWarnPct || pingMs > c.CronSpeedTestJobWarnLat {
		return "🟡 Degraded"
	}
	return "🟢 OK"
}

func runWithTimeout(ctx context.Context, logger zerolog.Logger, name string, fn func() error) error {
	start := time.Now()
	done := make(chan error, 1)
	go func() {
		logger.Debug().Str("operation", name).Msg("Operation started")
		err := fn()
		logger.Debug().Str("operation", name).Dur("duration", time.Since(start)).Err(err).Msg("Operation completed")
		done <- err
	}()

	select {
	case <-ctx.Done():
		logger.Error().Str("operation", name).Dur("duration", time.Since(start)).Msg("Operation timeout")
		return fmt.Errorf("%s timeout after %v", name, time.Since(start))
	case err := <-done:
		return err
	}
}
