Create a GitHub issue for paycent-telegram-bot: $ARGUMENTS

Steps:
1. Parse the argument as the issue title/description
2. Determine:
   - Type: bug, feature, or chore
   - Affected area: handler, client, store, bot, config, ci
   - Whether blocked on a paycent-core endpoint (issue #362 or new)
3. Create the issue:
   ```
   gh issue create \
     --repo mr-shifu/paycent-telegram-bot \
     --title "<title>" \
     --label "<type>" \
     --body "<body>"
   ```

Issue body structure:
```
## Overview
<One paragraph: what, why, scope>

## Acceptance Criteria
- [ ] <concrete, testable criterion>
- [ ] <concrete, testable criterion>

## Implementation Notes
<Affected files, paycent-core dependency if any>

## Blocked by
<paycent-core issue link if endpoint not yet implemented, otherwise N/A>
```

Output the created issue URL.
