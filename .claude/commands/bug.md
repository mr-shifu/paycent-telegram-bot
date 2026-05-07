---
description: Plan a surgical bug fix — minimal changes, root cause focused
---

Plan a surgical bug fix for paycent-telegram-bot: $ARGUMENTS

Read @AGENTS.md first. Do not write any code yet.

Steps:
1. Read the relevant source files to identify the root cause
2. Check for common bot pitfalls:
   - Missing `requireAuth()` call before API access
   - Race condition on `SessionStore` (missing mutex usage)
   - Wrong `pendingState` transition leaving user stuck
   - Client method not checking `resp.IsError()`
   - Reply text using context from wrong chat ID
3. Write the plan to `specs/bug-<name>.md` with:
   - Root cause (exact file and line)
   - Minimal fix — exact lines to change
   - Verification:
     ```
     go build ./...
     go vet ./...
     ```

Rules:
- Touch as few files as possible
- Do not refactor handler structure unless it directly caused the bug

Do not implement — plan only. Output the path to the created spec file.
