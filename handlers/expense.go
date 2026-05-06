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

type ExpenseHandler struct {
	bot    *tgbotapi.BotAPI
	client *client.PaycentClient
	store  *store.SessionStore
	logger zerolog.Logger
}

func NewExpenseHandler(bot *tgbotapi.BotAPI, c *client.PaycentClient, s *store.SessionStore, logger zerolog.Logger) *ExpenseHandler {
	return &ExpenseHandler{bot: bot, client: c, store: s, logger: logger}
}

// HandleAdd handles /add <amount> <description>
// Example: /add 60 dinner
func (h *ExpenseHandler) HandleAdd(msg *tgbotapi.Message) {
	sess, ok := h.requireAuth(msg)
	if !ok {
		return
	}

	args := strings.Fields(msg.CommandArguments())
	if len(args) < 2 {
		h.send(msg.Chat.ID, "Usage: /add <amount> <description>\nExample: /add 60 dinner")
		return
	}

	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil || amount <= 0 {
		h.send(msg.Chat.ID, "Invalid amount. Example: /add 60 dinner")
		return
	}

	description := strings.Join(args[1:], " ")

	groupID, ok := h.groupIDForChat(msg.Chat.ID)
	if !ok {
		h.send(msg.Chat.ID, "This chat isn't linked to a Paycent group yet.\nUse /newgroup <name> to create one or /linkgroup <id> to connect an existing group.")
		return
	}

	expense, err := h.client.AddExpense(sess.PaycentJWT, groupID, description, amount)
	if err != nil {
		h.logger.Error().Err(err).Msg("add expense failed")
		h.send(msg.Chat.ID, "Failed to log expense. Please try again.")
		return
	}

	h.send(msg.Chat.ID, fmt.Sprintf(
		"✓ *%s* — $%.2f\nYou paid · split equally\n\nUse /balance to see updated balances.",
		expense.Description, expense.Amount,
	))
}

// HandleHistory handles /history
func (h *ExpenseHandler) HandleHistory(msg *tgbotapi.Message) {
	sess, ok := h.requireAuth(msg)
	if !ok {
		return
	}

	groupID, ok := h.groupIDForChat(msg.Chat.ID)
	if !ok {
		h.send(msg.Chat.ID, "This chat isn't linked to a Paycent group yet.")
		return
	}

	items, err := h.client.GetActivity(sess.PaycentJWT, groupID)
	if err != nil {
		h.logger.Error().Err(err).Msg("get activity failed")
		h.send(msg.Chat.ID, "Failed to fetch history. Please try again.")
		return
	}

	if len(items) == 0 {
		h.send(msg.Chat.ID, "No activity yet. Use /add <amount> <desc> to log your first expense.")
		return
	}

	var sb strings.Builder
	sb.WriteString("*Recent activity:*\n\n")
	limit := 10
	if len(items) < limit {
		limit = len(items)
	}
	for _, item := range items[:limit] {
		sb.WriteString(fmt.Sprintf("• %s\n", item.Description))
	}

	h.send(msg.Chat.ID, sb.String())
}

// groupIDForChat looks up the Paycent group linked to a Telegram chat.
// Currently a stub — will be backed by the DB/API once the link endpoint is implemented.
func (h *ExpenseHandler) groupIDForChat(chatID int64) (string, bool) {
	// TODO: replace with lookup via GET /api/v1/group/groups?telegram_chat_id=<chatID>
	_ = chatID
	return "", false
}

func (h *ExpenseHandler) requireAuth(msg *tgbotapi.Message) (*store.Session, bool) {
	sess, ok := h.store.Get(msg.From.ID)
	if !ok {
		h.send(msg.Chat.ID, "You need to connect your Paycent account first.\nSend me a private message and type /start.")
		return nil, false
	}
	return sess, true
}

func (h *ExpenseHandler) send(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	if _, err := h.bot.Send(m); err != nil {
		h.logger.Error().Err(err).Msg("send message failed")
	}
}
