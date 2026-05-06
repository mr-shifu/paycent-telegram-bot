package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mr-shifu/paycent-telegram-bot/client"
	"github.com/mr-shifu/paycent-telegram-bot/store"
	"github.com/rs/zerolog"
)

type StartHandler struct {
	bot    *tgbotapi.BotAPI
	client *client.PaycentClient
	store  *store.SessionStore
	logger zerolog.Logger
}

func NewStartHandler(bot *tgbotapi.BotAPI, c *client.PaycentClient, s *store.SessionStore, logger zerolog.Logger) *StartHandler {
	return &StartHandler{bot: bot, client: c, store: s, logger: logger}
}

func (h *StartHandler) Handle(msg *tgbotapi.Message) {
	userID := msg.From.ID

	if _, ok := h.store.Get(userID); ok {
		h.send(msg.Chat.ID, "You're already connected to Paycent.\n\nQuick commands:\n/newgroup <name> — create a group\n/balance — check balances\n/add <amount> <desc> — log an expense")
		return
	}

	h.store.SetPending(userID, store.PendingPhone)
	h.send(msg.Chat.ID, "Welcome to Paycent Bot! 👋\n\nI help you split expenses in group chats without needing the app.\n\nFirst, send me your phone number (with country code, e.g. +1234567890):")
}

func (h *StartHandler) HandlePhone(msg *tgbotapi.Message) {
	userID := msg.From.ID
	phone := msg.Text

	if err := h.client.RequestOTP(phone); err != nil {
		h.logger.Error().Err(err).Str("phone", phone).Msg("request OTP failed")
		h.send(msg.Chat.ID, "Couldn't send a verification code to that number. Please check it and try again.")
		return
	}

	h.store.SetPhone(userID, phone)
	h.store.SetPending(userID, store.PendingOTP)
	h.send(msg.Chat.ID, fmt.Sprintf("Code sent to %s. Send me the 6-digit code:", phone))
}

func (h *StartHandler) HandleOTP(msg *tgbotapi.Message) {
	userID := msg.From.ID
	otp := msg.Text

	phone, ok := h.store.GetPhone(userID)
	if !ok {
		h.send(msg.Chat.ID, "Session expired. Please send /start to try again.")
		return
	}

	result, err := h.client.VerifyOTP(phone, otp, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("verify OTP failed")
		h.send(msg.Chat.ID, "Invalid or expired code. Please send /start to try again.")
		return
	}

	h.store.Set(userID, result.Token, result.UserID)
	h.store.ClearPhone(userID)
	h.send(msg.Chat.ID, "You're connected to Paycent ✓\n\nAdd me to a group chat and use:\n/newgroup <name> — create a new group\n/balance — check your balances\n/add <amount> <desc> — log an expense")
}

func (h *StartHandler) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error().Err(err).Msg("send message failed")
	}
}
