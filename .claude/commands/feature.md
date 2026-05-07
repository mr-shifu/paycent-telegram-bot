Plan a new feature for paycent-telegram-bot: $ARGUMENTS

Read @AGENTS.md first. Do not write any code yet.

Steps:
1. Determine if this requires a new paycent-core endpoint — if yes, note it as a dependency and reference the relevant paycent-core issue
2. List every file to create or modify:
   - `handlers/` — handler method (which file, which handler struct)
   - `client/paycent.go` — new API method (or stub if endpoint not yet in core)
   - `bot/router.go` — command registration
   - `store/session.go` — new pending state if multi-step flow needed
3. Write the plan to `specs/<feature-name>.md` with:
   - One-paragraph description
   - New command(s): syntax and example usage
   - Files to create/modify with function signatures
   - Whether a paycent-core endpoint is needed and its status (available / pending issue #N)
   - Stub plan if endpoint not yet available
   - Validation:
     ```
     go build ./...
     go vet ./...
     ```

Do not implement — plan only. Output the path to the created spec file.
