package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mr-shifu/paycent-telegram-bot/client"
	"github.com/mr-shifu/paycent-telegram-bot/store"
	"github.com/rs/zerolog"
)

type GroupHandler struct {
	bot    *tgbotapi.BotAPI
	client *client.PaycentClient
	store  *store.SessionStore
	logger zerolog.Logger
}

func NewGroupHandler(bot *tgbotapi.BotAPI, c *client.PaycentClient, s *store.SessionStore, logger zerolog.Logger) *GroupHandler {
	return &GroupHandler{bot: bot, client: c, store: s, logger: logger}
}

// HandleNewGroup handles /newgroup <name>
func (h *GroupHandler) HandleNewGroup(msg *tgbotapi.Message) {
	sess, ok := h.requireAuth(msg)
	if !ok {
		return
	}

	args := strings.TrimSpace(msg.CommandArguments())
	if args == "" {
		h.send(msg.Chat.ID, "Usage: /newgroup <group name>\nExample: /newgroup Bali Trip")
		return
	}

	group, err := h.client.CreateGroup(sess.PaycentJWT, args, msg.Chat.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("create group failed")
		h.send(msg.Chat.ID, "Failed to create group. Please try again.")
		return
	}

	h.send(msg.Chat.ID, fmt.Sprintf("✓ Group *%s* created.\n\nNow log your first expense:\n/add <amount> <description>", group.Name))
}

// HandleLinkGroup handles /linkgroup — links this chat to an existing Paycent group
func (h *GroupHandler) HandleLinkGroup(msg *tgbotapi.Message) {
	sess, ok := h.requireAuth(msg)
	if !ok {
		return
	}

	groupID := strings.TrimSpace(msg.CommandArguments())
	if groupID == "" {
		h.send(msg.Chat.ID, "Usage: /linkgroup <group_id>\n\nFind your group ID in the Paycent app under group settings.")
		return
	}

	if err := h.client.LinkGroup(sess.PaycentJWT, groupID, msg.Chat.ID); err != nil {
		h.logger.Error().Err(err).Msg("link group failed")
		h.send(msg.Chat.ID, "Failed to link group. Make sure the group ID is correct and you're a member.")
		return
	}

	h.send(msg.Chat.ID, "✓ This chat is now linked to your Paycent group.\n\nTry /balance to see where everyone stands.")
}

func (h *GroupHandler) requireAuth(msg *tgbotapi.Message) (*store.Session, bool) {
	sess, ok := h.store.Get(msg.From.ID)
	if !ok {
		h.send(msg.Chat.ID, "You need to connect your Paycent account first.\nSend me a private message and type /start.")
		return nil, false
	}
	return sess, true
}

func (h *GroupHandler) send(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	if _, err := h.bot.Send(m); err != nil {
		h.logger.Error().Err(err).Msg("send message failed")
	}
}
