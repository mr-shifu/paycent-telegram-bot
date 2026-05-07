---
description: Create a structured PR with bot review checklist
---

Create a pull request for the current branch in paycent-telegram-bot.

Steps:
1. Gather context:
   ```
   git log main..HEAD --oneline
   git diff main...HEAD --stat
   ```
2. Run the review checklist:
   ```
   go build ./...
   go vet ./...
   ```
3. Create the PR targeting `main` (or $ARGUMENTS if specified):
   ```
   gh pr create --base main --title "<title>" --body "<body>"
   ```

PR body structure:
```
## Summary
- <bullet: what changed>
- <bullet: why>

## Risk
<High / Medium / Low> — <one line reason>

## Paycent-core dependencies
<List any endpoints this PR depends on and their status — available or pending issue #N>

## Test plan
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] New handlers call `requireAuth()` as first step
- [ ] New client methods handle both transport errors and non-2xx responses
- [ ] Stubs added with `// TODO: paycent-core#<issue>` for unimplemented endpoints
- [ ] Tested locally in polling mode with a real bot token

🤖 Generated with [Claude Code](https://claude.ai/claude-code)
```
