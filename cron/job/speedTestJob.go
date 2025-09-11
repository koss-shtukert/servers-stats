package job

import (
	"context"
	"fmt"
	"github.com/koss-shtukert/servers-stats/common"
	"github.com/koss-shtukert/servers-stats/config"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/showwin/speedtest-go/speedtest"
)

func SpeedTestJob(l *zerolog.Logger, c *config.Config, n common.Notifier) func() {
	return func() {
		logger := l.With().Str("type", "SpeedTestJob").Logger()
		logger.Debug().Msg("Starting")

		ctxUser, cancelUser := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelUser()

		userCh := make(chan *speedtest.User, 1)
		errCh := make(chan error, 1)

		go func() {
			u, err := speedtest.FetchUserInfo()
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

		servers, err := speedtest.FetchServers()
		if err != nil {
			logger.Err(err).Msg("FetchServers failed")
			n.SendMessage("⚠️ Speedtest: не вдалося отримати список серверів")
			return
		}

		var targets speedtest.Servers
		targets, err = servers.FindServer([]int{})
		if err != nil || len(targets) == 0 {
			if err == nil {
				err = fmt.Errorf("no server found")
			}
			logger.Err(err).Msg("FindServer failed")
			n.SendMessage("⚠️ Speedtest: не знайдено підходящого сервера")
			return
		}
		s := targets[0]

		if err := s.PingTest(nil); err != nil {
			logger.Err(err).Msg("PingTest failed")
			n.SendMessage("⚠️ Speedtest: помилка під час PingTest")
			return
		}
		if err := s.DownloadTest(); err != nil {
			logger.Err(err).Msg("DownloadTest failed")
			n.SendMessage("⚠️ Speedtest: помилка під час DownloadTest")
			return
		}
		if err := s.UploadTest(); err != nil {
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
