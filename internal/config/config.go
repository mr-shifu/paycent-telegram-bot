package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken  string `mapstructure:"TELEGRAM_BOT_TOKEN"`
	PaycentAPIURL  string `mapstructure:"PAYCENT_API_URL"`
	PaycentBotKey  string `mapstructure:"PAYCENT_BOT_API_KEY"`
	WebhookURL     string `mapstructure:"WEBHOOK_URL"`
	WebhookPort    string `mapstructure:"WEBHOOK_PORT"`
	UseWebhook     bool   `mapstructure:"USE_WEBHOOK"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("WEBHOOK_PORT", "8443")
	viper.SetDefault("USE_WEBHOOK", false)

	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
