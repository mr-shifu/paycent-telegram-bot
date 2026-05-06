package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mr-shifu/paycent-telegram-bot/handlers"
	"github.com/mr-shifu/paycent-telegram-bot/store"
	"github.com/rs/zerolog"
)

type Router struct {
	start   *handlers.StartHandler
	group   *handlers.GroupHandler
	expense *handlers.ExpenseHandler
	balance *handlers.BalanceHandler
	settle  *handlers.SettleHandler
	store   *store.SessionStore
	logger  zerolog.Logger
}

func NewRouter(
	start *handlers.StartHandler,
	group *handlers.GroupHandler,
	expense *handlers.ExpenseHandler,
	balance *handlers.BalanceHandler,
	settle *handlers.SettleHandler,
	store *store.SessionStore,
	logger zerolog.Logger,
) *Router {
	return &Router{
		start:   start,
		group:   group,
		expense: expense,
		balance: balance,
		settle:  settle,
		store:   store,
		logger:  logger,
	}
}

func (r *Router) Dispatch(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message

	// Commands
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			r.start.Handle(msg)
		case "newgroup":
			r.group.HandleNewGroup(msg)
		case "linkgroup":
			r.group.HandleLinkGroup(msg)
		case "add":
			r.expense.HandleAdd(msg)
		case "history":
			r.expense.HandleHistory(msg)
		case "balance":
			r.balance.HandleBalance(msg)
		case "whoowes":
			r.balance.HandleWhoOwes(msg)
		case "settle":
			r.settle.HandleSettle(msg)
		case "help":
			r.sendHelp(msg)
		default:
			// unknown command — silently ignore in group chats, respond in private
			if msg.Chat.IsPrivate() {
				r.sendHelp(msg)
			}
		}
		return
	}

	// Plain text in private chat — handle multi-step flows (phone, OTP)
	if !msg.Chat.IsPrivate() {
		return
	}

	userID := msg.From.ID
	switch r.store.GetPending(userID) {
	case store.PendingPhone:
		r.start.HandlePhone(msg)
	case store.PendingOTP:
		r.start.HandleOTP(msg)
	}
}

func (r *Router) sendHelp(msg *tgbotapi.Message) {
	// no-op placeholder — bot.go sends the help message directly
	_ = msg
}
