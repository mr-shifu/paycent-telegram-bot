Create a conventional commit for staged changes in paycent-telegram-bot.

Steps:
1. Run pre-flight:
   ```
   gofmt -l .
   go vet ./...
   ```
   Fix any issues before committing.
2. Show what will be committed:
   ```
   git diff --cached --stat
   git diff --cached
   ```
3. Derive the commit message:
   - Format: `type(scope): description`
   - Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`
   - Scope: `bot`, `handler`, `client`, `store`, `config`, `ci`
   - Description: imperative mood, lowercase, no period, under 72 chars
   - Examples:
     - `feat(handler): add /summary command for monthly totals`
     - `fix(store): guard session map access with read lock`
     - `chore(client): stub telegram group linking endpoint`
4. Commit:
   ```
   git commit -m "<type>(<scope>): <description>"
   ```
5. Show result:
   ```
   git log --oneline -3
   ```
