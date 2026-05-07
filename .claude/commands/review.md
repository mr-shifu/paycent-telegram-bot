---
description: Run the AGENTS.md review checklist against current changes
---

Run the AGENTS.md review checklist against current changes in paycent-telegram-bot.

Steps:
1. Get the diff:
   ```
   git diff --stat
   git diff
   ```
2. Run checks:
   ```
   go build ./...
   go vet ./...
   ```
3. Check each item and report pass/fail:

**All Changes**
- [ ] `go build ./...` compiles
- [ ] `go vet ./...` clean
- [ ] No `fmt.Println` or `log.Println` — uses zerolog
- [ ] No credentials committed

**New Handler Methods**
- [ ] `requireAuth()` called before any API or session access
- [ ] `send()` helper used — no inline `h.bot.Send()` calls
- [ ] Reply text is short and scannable (no walls of text)
- [ ] Uses `tgbotapi.ModeMarkdown` for formatting

**New Client Methods**
- [ ] Handles `err != nil` (transport error)
- [ ] Handles `resp.IsError()` (non-2xx HTTP)
- [ ] Returns typed result struct, not raw JSON

**New Commands**
- [ ] Registered in `bot/router.go` switch block
- [ ] If endpoint not yet in paycent-core: stub with `// TODO: paycent-core#<issue>`

**Session Store Changes**
- [ ] `sync.RWMutex` still used for all map access (not removed or bypassed)
- [ ] Only JWT and pending state stored — no expense or group data

Report each item as ✓ PASS, ✗ FAIL, or — N/A.
