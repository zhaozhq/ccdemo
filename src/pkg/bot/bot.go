package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"ccdemo/src/pkg/logger"
	"ccdemo/src/pkg/storage"
	"go.uber.org/zap"
)

// Bot wraps Telegram Bot API.
type Bot struct {
	api   *tgbotapi.BotAPI
	store *storage.Storage
}

// New creates a Bot. Returns error if token is invalid.
func New(token string, store *storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}
	return &Bot{
		api:   api,
		store: store,
	}, nil
}

// formatMessage formats hot search items into a Markdown message.
func formatMessage(items []storage.HotSearch) string {
	if len(items) == 0 {
		return "暂无热搜数据。"
	}

	var b strings.Builder
	b.WriteString("*今日热搜 Top20*\n\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. [%s](%s) 🔥%d (%s)\n", i+1, item.Title, item.URL, item.Heat, item.Platform)
	}
	return b.String()
}

// SendHotSearch sends formatted hot search list to a chat.
func (b *Bot) SendHotSearch(chatID int64, items []storage.HotSearch) error {
	text := formatMessage(items)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

// Start begins long-polling update loop and handles commands.
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		chatID := update.Message.Chat.ID
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		switch cmd {
		case "start":
			msg := tgbotapi.NewMessage(chatID, "欢迎使用热搜 Bot！发送 /hot 查看今日热搜。")
			msg.ParseMode = tgbotapi.ModeMarkdown
			if _, err := b.api.Send(msg); err != nil {
				logger.Error("send start message", zap.Error(err))
			}
		case "hot":
			var items []storage.HotSearch
			var err error
			if args != "" {
				items, err = b.store.ListByPlatform(args)
			} else {
				items, err = b.store.ListAll()
			}
			if err != nil {
				logger.Error("query hot search", zap.Error(err))
				msg := tgbotapi.NewMessage(chatID, "查询热搜数据失败，请稍后重试。")
				if _, sendErr := b.api.Send(msg); sendErr != nil {
					logger.Error("send error message", zap.Error(sendErr))
				}
				continue
			}
			if err := b.SendHotSearch(chatID, items); err != nil {
				logger.Error("send hot search", zap.Error(err))
			}
		default:
			msg := tgbotapi.NewMessage(chatID, "未知命令，可用命令：/start, /hot, /hot <platform>")
			if _, err := b.api.Send(msg); err != nil {
				logger.Error("send unknown command message", zap.Error(err))
			}
		}
	}
}
