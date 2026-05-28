package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"ccdemo/src/pkg/bot"
	"ccdemo/src/pkg/fetcher"
	"ccdemo/src/pkg/logger"
	"ccdemo/src/pkg/mcp"
	"ccdemo/src/pkg/scheduler"
	"ccdemo/src/pkg/storage"
	"go.uber.org/zap"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		logger.Fatal("invalid int64 env", zap.String("key", key), zap.Error(err))
	}
	return n
}

func main() {
	logger.Init(logger.Config{
		Dir:          "./logs",
		FileMinLevel: "info",
	})
	defer logger.Sync()

	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		logger.Fatal("failed to create data directory", zap.Error(err))
	}

	dbPath := getEnv("DB_PATH", "./data/hotsearch.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		logger.Fatal("failed to open storage", zap.Error(err))
	}
	defer store.Close()

	fetchers := []fetcher.Fetcher{
		fetcher.NewWeiboFetcher(),
		fetcher.NewBaiduFetcher(),
		fetcher.NewZhihuFetcher(),
		fetcher.NewDouyinFetcher(),
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	var telegramBot *bot.Bot
	var sched *scheduler.Scheduler

	if botToken != "" {
		var err error
		telegramBot, err = bot.New(botToken, store)
		if err != nil {
			logger.Fatal("failed to create bot", zap.Error(err))
		}

		chatID := getEnvInt64("TELEGRAM_CHAT_ID", 0)
		if chatID != 0 {
			sched = scheduler.New(store, fetchers, telegramBot, chatID)
			cronSpec := getEnv("PUSH_CRON", "0 9 * * *")
			if err := sched.Start(cronSpec); err != nil {
				logger.Fatal("failed to start scheduler", zap.Error(err))
			}
			logger.Info("scheduler started", zap.String("cron", cronSpec), zap.Int64("chat_id", chatID))

			// Initial fetch and store
			if err := sched.FetchAndStore(context.Background()); err != nil {
				logger.Error("initial fetch and store failed", zap.Error(err))
			}
		}

		// Start bot polling in background goroutine
		go telegramBot.Start()
		logger.Info("telegram bot polling started")
	}

	mcpServer := mcp.NewServer(store)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down")
		if sched != nil {
			sched.Stop()
		}
		if err := store.Close(); err != nil {
			logger.Error("failed to close storage", zap.Error(err))
		}
		os.Exit(0)
	}()

	logger.Info("starting MCP server")
	if err := mcpServer.Serve(); err != nil {
		logger.Fatal("MCP server error", zap.Error(err))
	}
}
