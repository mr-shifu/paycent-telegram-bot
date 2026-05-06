package main

import (
	"os"

	"github.com/mr-shifu/paycent-telegram-bot/bot"
	"github.com/mr-shifu/paycent-telegram-bot/internal/config"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	b, err := bot.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize bot")
	}

	if err := b.Run(); err != nil {
		logger.Fatal().Err(err).Msg("bot stopped")
	}
}
