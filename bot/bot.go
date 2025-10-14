package bot

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"
)

type Bot struct {
	tgBot    *tgbotapi.BotAPI
	chatId   int64
	logger   *zerolog.Logger
	lastCmd  map[string]time.Time
	cmdMutex sync.RWMutex
	running  map[string]bool
	runMutex sync.RWMutex
}

func CreateBot(c *config.Config, l *zerolog.Logger) (*Bot, error) {
	logger := l.With().Str("type", "bot").Logger()

	tgBot, err := tgbotapi.NewBotAPI(c.TgBotApiKey)
	if err != nil {
		return nil, fmt.Errorf("error creating bot: %w", err)
	}

	chatId, err := strconv.ParseInt(c.TgBotChatId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat id: %w", err)
	}

	bot := &Bot{
		tgBot:   tgBot,
		chatId:  chatId,
		logger:  &logger,
		lastCmd: make(map[string]time.Time),
		running: make(map[string]bool),
	}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Hi! Type /help to see available commands."},
		{Command: "help", Description: "Show help information"},
		{Command: "server_disk_usage", Description: "Show server disk usage"},
		{Command: "motioneye_disk_usage", Description: "Show motioneye disk usage"},
		{Command: "speedtest", Description: "Run speed test"},
	}
	if _, err := tgBot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		logger.Err(err).Msg("Bot SetMyCommands error")
	}

	return bot, nil
}

func (b *Bot) StartPolling(ctx context.Context, l *zerolog.Logger, c *config.Config) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				b.logger.Info().Msg("Telegram polling stopped")
				return
			default:
				// Use manual polling with proper error handling
				u := tgbotapi.NewUpdate(0)
				u.Timeout = 60

				updates, err := b.tgBot.GetUpdates(u)
				if err != nil {
					b.logger.Err(err).Msg("Failed to get updates")
					time.Sleep(5 * time.Second)
					continue
				}

				for _, update := range updates {
					b.handleUpdate(update, l, c)
				}
			}
		}
	}()
}

func (b *Bot) handleUpdate(update tgbotapi.Update, l *zerolog.Logger, c *config.Config) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	switch update.Message.Command() {
	case "start":
		b.SendMessage("Hi! Type /help to see available commands.")

	case "help":
		msg := "Available commands:\n" +
			"/server_disk_usage — Server disk usage\n" +
			"/motioneye_disk_usage — Motioneye disk usage\n" +
			"/speedtest — Run speedtest\n"
		b.SendMessage(msg)

	case "server_disk_usage":
		if b.CanExecuteCommand("server_disk_usage") {
			go b.ExecuteJob("server_disk_usage", func() {
				job.ServerDiskUsageJob(l, c, b)()
			})
		} else {
			b.SendMessage("⚠️ Please wait before running this command again")
		}

	case "motioneye_disk_usage":
		if b.CanExecuteCommand("motioneye_disk_usage") {
			go b.ExecuteJob("motioneye_disk_usage", func() {
				job.MotioneyeDiskUsageJob(l, c, b)()
			})
		} else {
			b.SendMessage("⚠️ Please wait before running this command again")
		}

	case "speedtest":
		if b.CanExecuteCommand("speedtest") {
			b.SendMessage("Running speedtest…")
			go b.ExecuteJob("speedtest", func() {
				job.SpeedTestJob(l, c, b)()
			})
		} else {
			b.SendMessage("⚠️ Speedtest is already running or please wait")
		}

	default:
		b.SendMessage("Unknown command. Try /help")
	}
}

func (b *Bot) SendMessage(m string) {
	msg := tgbotapi.NewMessage(b.chatId, m)
	if _, err := b.tgBot.Send(msg); err != nil {
		b.logger.Err(err).Msg("Failed to send message")
	}
}

func (b *Bot) CanExecuteCommand(cmd string) bool {
	b.cmdMutex.RLock()
	lastTime, exists := b.lastCmd[cmd]
	b.cmdMutex.RUnlock()

	b.runMutex.RLock()
	isRunning := b.running[cmd]
	b.runMutex.RUnlock()

	// Check if command is already running
	if isRunning {
		return false
	}

	// Rate limiting: 30 seconds for speedtest, 10 seconds for others
	cooldown := 10 * time.Second
	if cmd == "speedtest" {
		cooldown = 30 * time.Second
	}

	if exists && time.Since(lastTime) < cooldown {
		return false
	}

	return true
}

func (b *Bot) ExecuteJob(cmd string, jobFunc func()) {
	// Mark as running
	b.runMutex.Lock()
	b.running[cmd] = true
	b.runMutex.Unlock()

	// Update last execution time
	b.cmdMutex.Lock()
	b.lastCmd[cmd] = time.Now()
	b.cmdMutex.Unlock()

	// Execute job
	jobFunc()

	// Mark as finished
	b.runMutex.Lock()
	b.running[cmd] = false
	b.runMutex.Unlock()
}
