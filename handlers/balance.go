package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mr-shifu/paycent-telegram-bot/client"
	"github.com/mr-shifu/paycent-telegram-bot/store"
	"github.com/rs/zerolog"
)

type BalanceHandler struct {
	bot    *tgbotapi.BotAPI
	client *client.PaycentClient
	store  *store.SessionStore
	logger zerolog.Logger
}

func NewBalanceHandler(bot *tgbotapi.BotAPI, c *client.PaycentClient, s *store.SessionStore, logger zerolog.Logger) *BalanceHandler {
	return &BalanceHandler{bot: bot, client: c, store: s, logger: logger}
}

// HandleBalance handles /balance — shows who owes what in the linked group
func (h *BalanceHandler) HandleBalance(msg *tgbotapi.Message) {
	sess, ok := h.requireAuth(msg)
	if !ok {
		return
	}

	groupID, ok := h.groupIDForChat(msg.Chat.ID)
	if !ok {
		h.send(msg.Chat.ID, "This chat isn't linked to a Paycent group yet.\nUse /newgroup <name> or /linkgroup <id>.")
		return
	}

	balances, err := h.client.GetBalances(sess.PaycentJWT, groupID)
	if err != nil {
		h.logger.Error().Err(err).Msg("get balances failed")
		h.send(msg.Chat.ID, "Failed to fetch balances. Please try again.")
		return
	}

	if len(balances) == 0 {
		h.send(msg.Chat.ID, "Everyone is settled up ✓")
		return
	}

	var sb strings.Builder
	sb.WriteString("*Balances:*\n\n")
	for _, b := range balances {
		name := b.Name
		if name == "" {
			name = b.Username
		}
		if b.Amount > 0 {
			sb.WriteString(fmt.Sprintf("• %s is owed $%.2f\n", name, b.Amount))
		} else if b.Amount < 0 {
			sb.WriteString(fmt.Sprintf("• %s owes $%.2f\n", name, -b.Amount))
		}
	}
	sb.WriteString("\nUse /settle @user <amount> to record a payment.")

	h.send(msg.Chat.ID, sb.String())
}

// HandleWhoOwes handles /whoowes — alias for balance with a different framing
func (h *BalanceHandler) HandleWhoOwes(msg *tgbotapi.Message) {
	h.HandleBalance(msg)
}

func (h *BalanceHandler) groupIDForChat(chatID int64) (string, bool) {
	// TODO: replace with lookup via GET /api/v1/group/groups?telegram_chat_id=<chatID>
	_ = chatID
	return "", false
}

func (h *BalanceHandler) requireAuth(msg *tgbotapi.Message) (*store.Session, bool) {
	sess, ok := h.store.Get(msg.From.ID)
	if !ok {
		h.send(msg.Chat.ID, "You need to connect your Paycent account first.\nSend me a private message and type /start.")
		return nil, false
	}
	return sess, true
}

func (h *BalanceHandler) send(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	if _, err := h.bot.Send(m); err != nil {
		h.logger.Error().Err(err).Msg("send message failed")
	}
}
