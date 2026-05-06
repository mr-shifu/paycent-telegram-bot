package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mr-shifu/paycent-telegram-bot/client"
	"github.com/mr-shifu/paycent-telegram-bot/store"
	"github.com/rs/zerolog"
)

type SettleHandler struct {
	bot    *tgbotapi.BotAPI
	client *client.PaycentClient
	store  *store.SessionStore
	logger zerolog.Logger
}

func NewSettleHandler(bot *tgbotapi.BotAPI, c *client.PaycentClient, s *store.SessionStore, logger zerolog.Logger) *SettleHandler {
	return &SettleHandler{bot: bot, client: c, store: s, logger: logger}
}

// HandleSettle handles /settle @username <amount>
// Records that the sender paid <amount> to the mentioned user.
func (h *SettleHandler) HandleSettle(msg *tgbotapi.Message) {
	sess, ok := h.requireAuth(msg)
	if !ok {
		return
	}

	args := strings.Fields(msg.CommandArguments())
	if len(args) < 2 {
		h.send(msg.Chat.ID, "Usage: /settle @username <amount>\nExample: /settle @sara 34")
		return
	}

	targetUsername := strings.TrimPrefix(args[0], "@")
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil || amount <= 0 {
		h.send(msg.Chat.ID, "Invalid amount. Example: /settle @sara 34")
		return
	}

	groupID, ok := h.groupIDForChat(msg.Chat.ID)
	if !ok {
		h.send(msg.Chat.ID, "This chat isn't linked to a Paycent group yet.")
		return
	}

	// TODO: resolve targetUsername → Paycent user ID via GET /api/v1/group/groups/{id}/members
	targetUserID := targetUsername

	payment, err := h.client.CreatePayment(sess.PaycentJWT, groupID, targetUserID, amount)
	if err != nil {
		h.logger.Error().Err(err).Msg("create payment failed")
		h.send(msg.Chat.ID, "Failed to record payment. Please try again.")
		return
	}

	h.send(msg.Chat.ID, fmt.Sprintf(
		"✓ Payment recorded: $%.2f to @%s\n\nUse /balance to see updated balances.",
		payment.Amount, targetUsername,
	))
}

func (h *SettleHandler) groupIDForChat(chatID int64) (string, bool) {
	// TODO: replace with lookup via GET /api/v1/group/groups?telegram_chat_id=<chatID>
	_ = chatID
	return "", false
}

func (h *SettleHandler) requireAuth(msg *tgbotapi.Message) (*store.Session, bool) {
	sess, ok := h.store.Get(msg.From.ID)
	if !ok {
		h.send(msg.Chat.ID, "You need to connect your Paycent account first.\nSend me a private message and type /start.")
		return nil, false
	}
	return sess, true
}

func (h *SettleHandler) send(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	if _, err := h.bot.Send(m); err != nil {
		h.logger.Error().Err(err).Msg("send message failed")
	}
}
