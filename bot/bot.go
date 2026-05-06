package bot

import (
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mr-shifu/paycent-telegram-bot/client"
	"github.com/mr-shifu/paycent-telegram-bot/handlers"
	"github.com/mr-shifu/paycent-telegram-bot/internal/config"
	"github.com/mr-shifu/paycent-telegram-bot/store"
	"github.com/rs/zerolog"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	router *Router
	cfg    *config.Config
	logger zerolog.Logger
}

func New(cfg *config.Config, logger zerolog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	logger.Info().Str("username", api.Self.UserName).Msg("bot authorized")

	c := client.NewPaycentClient(cfg.PaycentAPIURL, cfg.PaycentBotKey)
	s := store.NewSessionStore()

	startH := handlers.NewStartHandler(api, c, s, logger)
	groupH := handlers.NewGroupHandler(api, c, s, logger)
	expenseH := handlers.NewExpenseHandler(api, c, s, logger)
	balanceH := handlers.NewBalanceHandler(api, c, s, logger)
	settleH := handlers.NewSettleHandler(api, c, s, logger)

	router := NewRouter(startH, groupH, expenseH, balanceH, settleH, s, logger)

	return &Bot{api: api, router: router, cfg: cfg, logger: logger}, nil
}

func (b *Bot) Run() error {
	if b.cfg.UseWebhook {
		return b.runWebhook()
	}
	return b.runPolling()
}

func (b *Bot) runPolling() error {
	b.logger.Info().Msg("starting in polling mode")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		b.router.Dispatch(update)
	}
	return nil
}

func (b *Bot) runWebhook() error {
	webhookURL := fmt.Sprintf("%s/%s", b.cfg.WebhookURL, b.cfg.TelegramToken)
	b.logger.Info().Str("url", webhookURL).Msg("setting webhook")

	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("create webhook: %w", err)
	}
	if _, err := b.api.Request(wh); err != nil {
		return fmt.Errorf("set webhook: %w", err)
	}

	updates := b.api.ListenForWebhook("/" + b.cfg.TelegramToken)

	addr := ":" + b.cfg.WebhookPort
	b.logger.Info().Str("addr", addr).Msg("webhook server listening")

	go func() {
		for update := range updates {
			b.router.Dispatch(update)
		}
	}()

	return http.ListenAndServe(addr, nil)
}
