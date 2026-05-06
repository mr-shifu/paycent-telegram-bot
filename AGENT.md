# Agent Rules — Paycent Telegram Bot

Read CLAUDE.md first for architecture and commands.

## Before Marking Any Task Done

- `go build ./...` must succeed — no compilation errors.
- `go vet ./...` must produce no output.
- New handler code must call `requireAuth()` before touching any API or session state.
- Any new `client/paycent.go` method must handle both the error and non-2xx response cases.

## Handler Rules

- Every handler file follows the same shape: struct with `bot`, `client`, `store`, `logger` fields; a `send()` helper; a `requireAuth()` helper. Do not deviate from this pattern.
- Never put business logic in `bot/router.go` — it dispatches only. Logic lives in handlers.
- Never call `paycent-core` API directly from `bot/router.go` or `bot/bot.go`.
- Reply text must be short and scannable — this is a chat interface, not a web form. No walls of text.

## Session Store

- The session store is in-memory only — it does not survive restarts. Do not treat it as a database.
- Only store JWTs and pending flow state here. Never store expense data, group data, or balances.
- If adding a new multi-step flow, add the pending state to `store/session.go` as a new `pendingState` constant.

## Adding Commands

When adding a new command:
1. Handler method in `handlers/` — one file per domain (start, group, expense, balance, settle).
2. Register in `bot/router.go` switch block.
3. Add `client/paycent.go` method if a new API endpoint is needed.
4. Document the command in CLAUDE.md.

Never add a command that requires an endpoint not yet implemented in `paycent-core` without adding a clearly labeled stub and a `// TODO:` comment pointing to the relevant paycent-core issue.

## Sensitive Areas — Flag for Human Review

- `internal/config/config.go` — env var changes affect all deployments
- `bot/bot.go` — webhook registration logic; wrong changes break all message delivery
- `store/session.go` — JWT handling; security-sensitive
- Any change to how `TELEGRAM_BOT_TOKEN` or `PAYCENT_BOT_API_KEY` are used

## What Not to Do

- Do not add a local database or file-based persistence — all data lives in paycent-core.
- Do not parse Telegram usernames as Paycent user IDs — always resolve via the API.
- Do not send unsolicited messages to users — only reply to commands or pending flow steps.
- Do not add commands that only work in private chats without documenting that restriction clearly in the reply text.
- Do not commit `.env` or any file containing real tokens.
