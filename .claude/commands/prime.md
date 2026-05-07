Read the following in order:
1. @AGENTS.md — full project context and architecture
2. @cmd/main.go — entry point
3. @bot/router.go — command dispatch table
4. @client/paycent.go — all paycent-core API methods and their status

Then run:
- `git status`
- `git log --oneline -10`
- `git diff --stat HEAD~1`

Then summarize:
- Current branch and any uncommitted changes
- The last 3 significant changes in git log
- Which commands are currently stubbed (groupIDForChat returns false) and which are fully wired
- Which paycent-core endpoints from issue #362 are still pending
- What you are ready to help with
