package bot

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koss-shtukert/servers-stats/config"
	"github.com/koss-shtukert/servers-stats/cron/job"
	"github.com/rs/zerolog"
)

type Bot struct {
	tgBot  *tgbotapi.BotAPI
	chatId int64
	logger *zerolog.Logger
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
		tgBot:  tgBot,
		chatId: chatId,
		logger: &logger,
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
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tgBot.GetUpdatesChan(u)

	go func() {
		defer b.tgBot.StopReceivingUpdates()
		for {
			select {
			case <-ctx.Done():
				b.logger.Info().Msg("Telegram polling stopped")
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				b.handleUpdate(update, l, c)
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
		go job.ServerDiskUsageJob(l, c, b)()

	case "motioneye_disk_usage":
		go job.MotioneyeDiskUsageJob(l, c, b)()

	case "speedtest":
		b.SendMessage("Running speedtest…")
		go job.SpeedTestJob(l, c, b)()

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
