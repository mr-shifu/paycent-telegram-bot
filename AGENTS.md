# Paycent Telegram Bot — Agent Guide

Go Telegram bot for Paycent. Zero-friction acquisition channel — users split expenses without installing the app. Thin layer over the `paycent-core` REST API; owns no database.

- **Module**: `github.com/mr-shifu/paycent-telegram-bot`
- **Go**: 1.25+
- **Telegram SDK**: go-telegram-bot-api/v5
- **HTTP Client**: go-resty/v2 → paycent-core API
- **Config**: spf13/viper (`.env`)
- **Logging**: zerolog

---

## Directory Structure

```
paycent-telegram-bot/
├── cmd/main.go              # Entry point
├── bot/
│   ├── bot.go               # Wires handlers; polling vs webhook mode
│   └── router.go            # Dispatches commands to handlers
├── handlers/
│   ├── start.go             # /start + phone/OTP account linking
│   ├── group.go             # /newgroup, /linkgroup
│   ├── expense.go           # /add, /history
│   ├── balance.go           # /balance, /whoowes
│   └── settle.go            # /settle @user <amount>
├── client/paycent.go        # Typed HTTP client for paycent-core
├── store/session.go         # In-memory JWT + pending state per Telegram user
└── internal/config/config.go
```

---

## Key Commands

```bash
go build ./...          # verify compilation
go vet ./...            # static analysis
go run cmd/main.go      # run locally (polling mode)
go mod tidy             # tidy dependencies
```

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Yes | From @BotFather |
| `PAYCENT_API_URL` | Yes | e.g. `http://localhost:8080` |
| `PAYCENT_BOT_API_KEY` | Yes | Internal bot↔API key |
| `USE_WEBHOOK` | No | `false` for polling (default) |
| `WEBHOOK_URL` | Prod | Public HTTPS URL |
| `LOG_LEVEL` | No | `debug`/`info`/`warn`/`error` |

---

## Architecture

- Bot owns **no data** — all state lives in `paycent-core`
- Only local state: in-memory `SessionStore` (JWT + pending flow per `telegram_user_id`)
- Message flow: `Telegram update → router.Dispatch → handler → client.PaycentClient → API`
- Two modes: **polling** (local dev) and **webhook** (production)

---

## Supported Commands

| Command | What it does |
|---|---|
| `/start` | Account linking via phone + OTP |
| `/newgroup <name>` | Create a Paycent group linked to this chat |
| `/linkgroup <id>` | Link chat to an existing Paycent group |
| `/add <amount> <desc>` | Log an expense (you paid, split equally) |
| `/balance` | Show balances in this group |
| `/whoowes` | Alias for /balance |
| `/settle @user <amount>` | Record a payment |
| `/history` | Last 10 activity items |

---

## Known Stubs

`groupIDForChat()` in `expense.go`, `balance.go`, `settle.go` returns `("", false)` until `paycent-core` implements `GET /api/v1/group/groups?telegram_chat_id=<id>` (tracked in paycent-core issue #362).

---

## Before Marking Any Task Done

- `go build ./...` must succeed
- `go vet ./...` must produce no output
- Every handler must call `requireAuth()` before any API or session access
- Every `client/paycent.go` method must handle both errors and non-2xx responses

---

## Handler Rules

- Every handler struct has: `bot`, `client`, `store`, `logger` fields + `send()` + `requireAuth()` helpers
- Never put logic in `bot/router.go` — dispatches only
- Never call paycent-core API from `bot/bot.go` or `bot/router.go`
- Reply text must be short and scannable — chat interface, not a form

---

## Adding a New Command

1. Add handler method in `handlers/` (one file per domain)
2. Register in `bot/router.go` switch block
3. Add `client/paycent.go` method if a new API endpoint is needed
4. If endpoint doesn't exist in core yet, add a stub with a `// TODO: paycent-core#<issue>` comment

---

## Sensitive Areas — Require Human Confirmation

- `store/session.go` — JWT handling
- `bot/bot.go` — webhook registration (wrong changes break all delivery)
- `internal/config/config.go` — affects all deployments
- Any change to `TELEGRAM_BOT_TOKEN` or `PAYCENT_BOT_API_KEY` usage

---

## What Not to Do

- Do not add local DB or file persistence — all data lives in paycent-core
- Do not use Telegram usernames as Paycent user IDs — always resolve via API
- Do not send unsolicited messages — only reply to commands or pending flow steps
- Do not commit `.env` or real tokens
