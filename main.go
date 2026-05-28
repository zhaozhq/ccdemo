package main

import (
	"hello/pkg/logger"
)

func main() {
	logger.Init(logger.Config{
		Dir:          "./logs",
		FileMinLevel: "info",
	})
	defer logger.Sync()

	logger.Info("Hello, World!")
	logger.Debug("this is debug, screen only")
}
