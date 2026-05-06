> **Version**: v1.0
> **Date**: 2026-05-06

# Paycent Telegram Bot — AGENTS.md

Paycent Telegram Bot is a Go service that lets users create groups, log expenses, check balances, and record settlements entirely within Telegram — no app install required. It is a thin HTTP client over the `paycent-core` REST API: it owns no database, no business logic, and no persistent state beyond in-memory JWT sessions. Its primary role is zero-friction user acquisition — one organizer adding the bot pulls multiple new users into the Paycent ecosystem.

---

## Navigation Table

| Path | Purpose | Entry Point |
|---|---|---|
| `cmd/main.go` | Entry point — loads config, inits logger, starts bot | `main()` |
| `internal/config/config.go` | All env-var parsing via viper | `config.Load()` |
| `bot/bot.go` | Bot struct — wires all handlers, chooses polling vs webhook | `bot.New()`, `bot.Run()` |
| `bot/router.go` | Dispatches Telegram updates to handlers | `Router.Dispatch()` |
| `handlers/start.go` | `/start` command + multi-step phone/OTP linking | `StartHandler.Handle()`, `HandlePhone()`, `HandleOTP()` |
| `handlers/group.go` | `/newgroup`, `/linkgroup` commands | `GroupHandler.HandleNewGroup()`, `HandleLinkGroup()` |
| `handlers/expense.go` | `/add`, `/history` commands | `ExpenseHandler.HandleAdd()`, `HandleHistory()` |
| `handlers/balance.go` | `/balance`, `/whoowes` commands | `BalanceHandler.HandleBalance()` |
| `handlers/settle.go` | `/settle @user <amount>` command | `SettleHandler.HandleSettle()` |
| `client/paycent.go` | Typed HTTP client for all paycent-core API calls | `client.NewPaycentClient()` |
| `store/session.go` | In-memory JWT + pending flow state per Telegram user | `SessionStore.Get()`, `Set()`, `SetPending()` |

---

## Architecture

```
Telegram API
     │
     ▼ updates (polling or webhook)
┌─────────────────────────────────────────┐
│              bot.Bot                    │
│  Run() → runPolling() / runWebhook()    │
└──────────────────┬──────────────────────┘
                   │
                   ▼
        ┌──────────────────┐
        │   bot.Router     │
        │   Dispatch()     │   ← routes commands + plain text (OTP flow)
        └──────┬───────────┘
               │
    ┌──────────┼──────────────────────────┐
    ▼          ▼          ▼              ▼
StartHandler GroupHandler ExpenseHandler BalanceHandler SettleHandler
    │              │           │              │              │
    └──────────────┴───────────┴──────────────┴──────────────┘
                              │
                              ▼
                  ┌─────────────────────┐
                  │  client.Paycent     │
                  │  go-resty/v2 HTTP   │
                  └──────────┬──────────┘
                             │  REST /api/v1/*
                             ▼
                  ┌─────────────────────┐
                  │    paycent-core     │
                  │    (Go REST API)    │
                  └─────────────────────┘

Local state only:
  store.SessionStore  ← in-memory map[telegram_user_id → JWT + pendingState]
  (lost on restart — not a database)
```

---

## Development Commands

```bash
# Setup
cp .env.example .env     # fill in TELEGRAM_BOT_TOKEN + PAYCENT_API_URL

# Run (polling mode — no public URL needed)
go run cmd/main.go

# Build binary
go build -o paycent-telegram-bot ./cmd

# Verify compilation
go build ./...

# Static analysis
go vet ./...

# Tidy dependencies
go mod tidy
```

---

## Coding Rules

### Style
- Follow standard Go conventions — `gofmt` formatted
- Handler structs always have exactly: `bot *tgbotapi.BotAPI`, `client *client.PaycentClient`, `store *store.SessionStore`, `logger zerolog.Logger`
- Every handler file exposes a `send(chatID int64, text string)` helper and a `requireAuth(msg) (*Session, bool)` helper — never inline these

### Error Handling
- `client/paycent.go` methods must handle both `err != nil` (network/transport) and `resp.IsError()` (non-2xx HTTP) cases
- Never propagate raw errors to Telegram users — translate to a short, friendly message via `h.send()`
- Log errors with context: `h.logger.Error().Err(err).Msg("...")`

### Logging
- Use `zerolog` only — passed as `logger zerolog.Logger` into every handler
- **Do not** use `fmt.Println`, `log.Println`
- Log at `.Error()` for API failures, `.Info()` for lifecycle events, `.Debug()` for message processing

### Session Store
- In-memory only — does not survive restarts; do not treat as a DB
- Only store: JWT, Paycent user ID, and `pendingState` (multi-step flow tracking)
- Never store expense data, group data, or balances in the store

### Telegram UX
- Reply text must be short and scannable — this is a mobile chat interface
- Only speak when addressed (command or pending flow) — never send unsolicited messages
- Use `tgbotapi.ModeMarkdown` for formatting; avoid HTML mode

### What NOT to Do
- Do not add a local database or file-based persistence — all data lives in `paycent-core`
- Do not parse Telegram `@username` directly as a Paycent user ID — always resolve via API
- Do not put any logic in `bot/router.go` — it dispatches only; logic belongs in handlers
- Do not call `paycent-core` API from `bot/bot.go` or `bot/router.go`
- Do not add a command that requires a `paycent-core` endpoint that doesn't exist yet without a stub and a `// TODO: paycent-core#<issue>` comment
- Do not commit `.env` or any file containing real tokens

---

## Risk Areas

| Path / Area | Risk Level | Notes |
|---|---|---|
| `store/session.go` | **High** | Stores JWTs in memory; race conditions possible without `sync.RWMutex` (already present — do not remove) |
| `bot/bot.go` — webhook setup | **High** | Wrong webhook URL or token path breaks all message delivery silently |
| `client/paycent.go` — auth methods | **High** | Handles JWT exchange; errors here lock users out |
| `internal/config/config.go` | **Medium** | Env vars affect all deployments; missing `TELEGRAM_BOT_TOKEN` panics at startup |
| `handlers/start.go` — OTP flow | **Medium** | Multi-step state machine; wrong state transitions leave users stuck |
| `handlers/settle.go` | **Medium** | Records financial settlements; wrong `paidToUserID` resolution credits wrong person |
| `bot/router.go` | **Low** | Command dispatch; adding a case here is low-risk but removing one silently drops commands |

---

## Review Checklist

### All Changes
- [ ] `go build ./...` compiles without errors
- [ ] `go vet ./...` produces no output
- [ ] New handler calls `requireAuth()` before any API or session access
- [ ] New `client/paycent.go` method handles both `err != nil` and `resp.IsError()`
- [ ] No `fmt.Println` introduced — use zerolog
- [ ] Reply text is short and mobile-friendly

### New Commands
- [ ] Handler method added in appropriate `handlers/` file (grouped by domain)
- [ ] Registered in `bot/router.go` switch block
- [ ] `client/paycent.go` method added if new API endpoint needed
- [ ] If endpoint not yet in `paycent-core`: stub with `// TODO: paycent-core#<issue>`

### Security-Sensitive Changes
- [ ] `store/session.go` still uses `sync.RWMutex` for all map access
- [ ] `TELEGRAM_BOT_TOKEN` and `PAYCENT_BOT_API_KEY` not logged or exposed in responses
- [ ] Webhook endpoint validates Telegram token in URL path before processing

---

## Environment Variables

| Variable | Required | Purpose | Example |
|---|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Yes | Bot token from @BotFather | `123456:ABC-DEF...` |
| `PAYCENT_API_URL` | Yes | Base URL of paycent-core | `http://localhost:8080` |
| `PAYCENT_BOT_API_KEY` | Yes | Internal key for bot↔API auth | `secret-bot-key` |
| `USE_WEBHOOK` | No | `true` for webhook mode, `false` for polling | `false` |
| `WEBHOOK_URL` | Prod only | Public HTTPS URL the bot is reachable at | `https://bot.paycent.net` |
| `WEBHOOK_PORT` | No | Port for webhook HTTP server | `8443` |
| `LOG_LEVEL` | No | Verbosity: `debug`, `info`, `warn`, `error` | `info` |

---

## Cross-Project Dependencies

### paycent-core (upstream)
All data and business logic. This bot is a pure consumer.

| Endpoint | Used by | Status |
|---|---|---|
| `POST /api/v1/auth/telegram/verify` | `StartHandler.HandlePhone()` | Pending — issue #362 |
| `POST /api/v1/auth/telegram/link` | `StartHandler.HandleOTP()` | Pending — issue #362 |
| `POST /api/v1/group/groups` | `GroupHandler.HandleNewGroup()` | Available |
| `POST /api/v1/group/groups/{id}/telegram/link` | `GroupHandler.HandleLinkGroup()` | Pending — issue #362 |
| `GET  /api/v1/group/groups?telegram_chat_id=<id>` | `groupIDForChat()` stub | Pending — issue #362 |
| `POST /api/v1/group/groups/{id}/expenses` | `ExpenseHandler.HandleAdd()` | Available |
| `GET  /api/v1/group/groups/{id}/balances` | `BalanceHandler.HandleBalance()` | Available |
| `POST /api/v1/group/groups/{id}/payments` | `SettleHandler.HandleSettle()` | Available |
| `GET  /api/v1/group/groups/{id}/activity` | `ExpenseHandler.HandleHistory()` | Available |

**When paycent-core adds a new endpoint:** add a corresponding method to `client/paycent.go` and replace the relevant stub in `handlers/`.

**When paycent-core changes a response shape:** update the matching struct in `client/paycent.go`.

---

_This file follows the Agentic Engineering Workflow Framework v2.0 standard._
