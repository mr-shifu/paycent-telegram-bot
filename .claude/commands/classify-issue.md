---
description: Route a GitHub issue to /bug, /feature, or /chore
---

Classify GitHub issue #$ARGUMENTS and route it to the right workflow.

Steps:
1. Fetch the issue:
   ```
   gh issue view $ARGUMENTS --repo mr-shifu/paycent-telegram-bot
   ```
2. Read the title, body, and labels
3. Classify as one of:
   - **`/bug`** — broken or incorrect bot behaviour
   - **`/feature`** — new command or capability
   - **`/chore`** — dependency update, refactor, CI change
4. Report:
   - Issue title and number
   - Classification and reason
   - Whether it is blocked on a paycent-core endpoint (if yes, reference or create a core issue)
   - Suggested command to run next
   - Suggested labels if missing
