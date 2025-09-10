package job

import (
	"context"
	"fmt"
	"github.com/koss-shtukert/servers-stats/config"
	"strings"
	"time"

	"github.com/koss-shtukert/servers-stats/bot"
	"github.com/rs/zerolog"
	"github.com/showwin/speedtest-go/speedtest"
)

func SpeedTestJob(l *zerolog.Logger, c *config.Config, b *bot.Bot) func() {
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
			b.SendMessage("⚠️ Speedtest: timeout під час отримання інформації про мережу")
			return
		case err := <-errCh:
			logger.Err(err).Msg("FetchUserInfo failed")
			b.SendMessage("⚠️ Speedtest: не вдалося отримати інформацію про мережу")
			return
		case u := <-userCh:
			user = u
		}

		servers, err := speedtest.FetchServers()
		if err != nil {
			logger.Err(err).Msg("FetchServers failed")
			b.SendMessage("⚠️ Speedtest: не вдалося отримати список серверів")
			return
		}

		var targets speedtest.Servers
		targets, err = servers.FindServer([]int{})
		if err != nil || len(targets) == 0 {
			if err == nil {
				err = fmt.Errorf("no server found")
			}
			logger.Err(err).Msg("FindServer failed")
			b.SendMessage("⚠️ Speedtest: не знайдено підходящого сервера")
			return
		}
		s := targets[0]

		if err := s.PingTest(nil); err != nil {
			logger.Err(err).Msg("PingTest failed")
			b.SendMessage("⚠️ Speedtest: помилка під час PingTest")
			return
		}
		if err := s.DownloadTest(); err != nil {
			logger.Err(err).Msg("DownloadTest failed")
			b.SendMessage("⚠️ Speedtest: помилка під час DownloadTest")
			return
		}
		if err := s.UploadTest(); err != nil {
			logger.Err(err).Msg("UploadTest failed")
			b.SendMessage("⚠️ Speedtest: помилка під час UploadTest")
			return
		}

		msg := formatSpeedtest(user, s)
		b.SendMessage(msg)

		logger.Debug().Msg("Finished")
	}
}

func fmtMBps(mbps float64) string {
	return fmt.Sprintf("%.2f MB/s", mbps)
}

func speedStatus(dlMbps, ulMbps, pingMs, lossPct float64) string {
	expDown := 950.0
	expUp := 950.0
	warnPct := 0.80
	critPct := 0.60
	latWarn := 10
	latCrit := 20
	lossWarn := 0.5
	lossCrit := 2.0

	downRate := dlMbps / expDown
	upRate := ulMbps / expUp

	if downRate < critPct || upRate < critPct || pingMs > float64(latCrit) || (lossPct > lossCrit) {
		return "🔴 Poor"
	}
	if downRate < warnPct || upRate < warnPct || pingMs > float64(latWarn) || (lossPct > lossWarn) {
		return "🟡 Degraded"
	}
	return "🟢 OK"
}

func formatSpeedtest(user *speedtest.User, s *speedtest.Server) string {
	dlMbps := float64(s.DLSpeed / 1024.0)
	ulMbps := float64(s.ULSpeed / 1024.0)
	pingMs := float64(s.Latency) / float64(time.Millisecond)

	status := speedStatus(dlMbps, ulMbps, pingMs, 0.0)

	isp := strings.TrimSpace(user.Isp)
	if isp == "" {
		isp = "n/a"
	}
	extIP := strings.TrimSpace(user.IP)
	if extIP == "" {
		extIP = "n/a"
	}

	return fmt.Sprintf(
		"🚀 Ookla Speedtest\n\n"+
			"⬇️ Download: %s\n"+
			"⬆️ Upload:   %s\n"+
			"🕒 Ping:     %.1f ms\n"+
			"🏷 ISP:      %s\n"+
			"🌐 External: %s\n"+
			"🗺 Server:   %s — %s (%s) • ID %s\n"+
			"✅ Status:   %s",
		fmtMBps(dlMbps),
		fmtMBps(ulMbps),
		pingMs,
		isp,
		extIP,
		s.Name, s.Country, s.Sponsor, s.ID,
		status,
	)
}
