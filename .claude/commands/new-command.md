Scaffold a new Telegram bot command: $ARGUMENTS

Expected format: `/<command> <args> — <description>`
Example: `/summary — show monthly expense summary for the linked group`

Read @AGENTS.md before starting.

Steps:
1. **Client method** (`client/paycent.go`):
   - Add a typed method for the required paycent-core endpoint
   - If the endpoint doesn't exist yet in core, add a stub that returns `("", false)` with a `// TODO: paycent-core#<issue>` comment
2. **Handler method** (`handlers/<domain>.go`):
   - Add to the appropriate handler file (group, expense, balance, settle, start)
   - Follow the exact struct shape: `bot`, `client`, `store`, `logger` fields
   - First line: `sess, ok := h.requireAuth(msg)` — no exceptions
   - Reply text: short, scannable, markdown-formatted
3. **Router** (`bot/router.go`):
   - Add `case "<command>":` to the `switch msg.Command()` block
   - Call the handler method
4. Validate:
   ```
   go build ./...
   go vet ./...
   ```
5. Confirm the command is listed in the `/help` reply text if a help handler exists.
