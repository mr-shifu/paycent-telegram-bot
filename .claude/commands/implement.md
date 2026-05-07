---
description: Execute a plan from specs/
---

Implement the plan from: $ARGUMENTS

If no argument given, list all files in specs/ and ask which to implement.

Steps:
1. Read the spec file in full
2. Read @AGENTS.md to confirm handler rules before touching any file
3. Implement in order: store → client → handler → router
4. After each file, run:
   ```
   go build ./...
   go vet ./...
   ```
   Fix errors before continuing.
5. Confirm every new handler:
   - Has `requireAuth()` as the first call
   - Has a `send()` helper using the standard signature
   - Handles both `err != nil` and `resp.IsError()` in client calls
   - Uses short, scannable reply text
6. If the spec includes a stub (paycent-core endpoint not yet available):
   - Add `// TODO: paycent-core#<issue>` comment on the stub
   - Return a user-friendly "not yet available" message from the handler
7. Report what was implemented and any deviations from the spec.
