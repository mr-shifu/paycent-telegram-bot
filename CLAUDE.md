# Paycent Telegram Bot — Claude Code Guide

## Project Overview

Telegram bot for Paycent — lets users create groups, log expenses, check balances, and settle debts entirely within Telegram. No app install required. Acts as the primary zero-friction acquisition channel for Paycent.

- **Module**: `github.com/mr-shifu/paycent-telegram-bot`
- **Go**: 1.25+
- **Telegram SDK**: go-telegram-bot-api/v5
- **HTTP Client**: go-resty/v2 (calls paycent-core REST API)
- **Config**: spf13/viper (env vars from `.env`)
- **Logging**: zerolog

## Directory Structure

```
paycent-telegram-bot/
├── cmd/main.go              # Entry point — loads config, inits logger, starts bot
├── bot/
│   ├── bot.go               # Bot struct: wires handlers, chooses polling vs webhook
│   └── router.go            # Dispatches incoming updates to handlers
├── handlers/
│   ├── start.go             # /start + multi-step phone/OTP account linking
│   ├── group.go             # /newgroup, /linkgroup
│   ├── expense.go           # /add, /history
│   ├── balance.go           # /balance, /whoowes
│   └── settle.go            # /settle @user <amount>
├── client/paycent.go        # Typed HTTP client for all paycent-core API calls
├── store/session.go         # In-memory JWT store + pending state per Telegram user
└── internal/config/config.go # Viper config — all env vars parsed here
```

## Architecture

The bot is a thin layer over the `paycent-core` REST API. It owns no database — all state lives in paycent-core. The only local state is the in-memory session store (`store/SessionStore`) which maps `telegram_user_id → JWT + pending flow state`.

**Message flow:**
```
Telegram update → bot.Router.Dispatch → handler → client.PaycentClient → paycent-core API
```

**Two runtime modes:**
- **Polling** (default, `USE_WEBHOOK=false`) — for local development, no public URL needed
- **Webhook** (`USE_WEBHOOK=true`) — for production, Telegram pushes updates to `WEBHOOK_URL/{token}`

## Development Commands

```bash
# Run locally (polling mode)
cp .env.example .env   # fill in TELEGRAM_BOT_TOKEN and PAYCENT_API_URL
go run cmd/main.go

# Build binary
go build -o paycent-telegram-bot ./cmd

# Verify compilation
go build ./...

# Tidy dependencies
go mod tidy

# Vet
go vet ./...
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Yes | From @BotFather |
| `PAYCENT_API_URL` | Yes | Base URL of paycent-core (e.g. `http://localhost:8080`) |
| `PAYCENT_BOT_API_KEY` | Yes | Internal key for bot→API authentication |
| `USE_WEBHOOK` | No | `false` for polling (default), `true` for production |
| `WEBHOOK_URL` | Production | Public HTTPS URL the bot is reachable at |
| `WEBHOOK_PORT` | No | Default `8443` |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` (default `info`) |

## Adding a New Command

1. Add the handler method to the relevant file in `handlers/` (or create a new file).
2. Register it in `bot/router.go` under the `switch msg.Command()` block.
3. Add the corresponding `client/paycent.go` method if a new API endpoint is involved.

Handler signature convention — every handler receives `*tgbotapi.Message` and calls `h.requireAuth()` first:

```go
func (h *FooHandler) HandleFoo(msg *tgbotapi.Message) {
    sess, ok := h.requireAuth(msg)
    if !ok {
        return
    }
    // ... call h.client, send reply
}
```

## Known Stubs (TODOs)

`groupIDForChat(chatID int64)` is stubbed in `expense.go`, `balance.go`, and `settle.go`. It returns `("", false)` until `paycent-core` exposes:
```
GET /api/v1/group/groups?telegram_chat_id=<id>
```
Once that endpoint exists, replace the stub with a `client.PaycentClient` call.

## Relation to paycent-core

This bot depends on `paycent-core` for all data. The backend changes required for full integration are tracked in **paycent-core issue #362**:
- `POST /api/v1/auth/telegram/verify` — request OTP
- `POST /api/v1/auth/telegram/link` — verify OTP, link Telegram identity, return JWT
- `POST /api/v1/group/groups/{id}/telegram/link` — link group to Telegram chat
- `GET  /api/v1/group/groups?telegram_chat_id=<id>` — resolve chat → group
- `POST /api/v1/auth/deeplink-token` — generate bot→app handoff token

## Deployment

Production runs in webhook mode behind a reverse proxy (Nginx). The bot binary listens on `WEBHOOK_PORT` for Telegram's HTTPS callbacks.

```bash
# Production build
go build -o paycent-telegram-bot ./cmd

# Run (webhook mode)
USE_WEBHOOK=true WEBHOOK_URL=https://bot.paycent.app ./paycent-telegram-bot
```

Telegram requires HTTPS for webhooks — ensure the server has a valid TLS certificate.
