# gator

RSS reader: Go + Echo, SQLite via sqlc, server-rendered `html/template` views styled with Tailwind, deployed to Fly.io. Commands live in the `Makefile`.

## Deploys

A push to `main` deploys to production (`.github/workflows/fly-deploy.yml`): `make check` (the same checks as the pre-commit hook), then `flyctl deploy`. Work on a branch; merging to `main` is a release.

The container runs `goose up` on every start, before the server (Dockerfile `CMD`), so a failed migration fails the deploy's health check. A migration that succeeds stays applied even if the deploy then rolls back to the old code, so keep schema changes backward compatible: add a column in one release, start reading it in the next.

## Database

- Schema is goose migrations in `sql/schema/NNN_name.sql`, numbered sequentially. Queries are in `sql/queries/`.
- `internal/database/` is sqlc output. Change the SQL in `sql/queries/` or `sql/schema/`, then run `make generate` and commit the regenerated files with it.
- `make` targets default `DATABASE_PATH` to `./data/gator.db`; an exported value overrides it.
- Service tests get an in-memory SQLite database with the real migrations applied: `setupTestDB` in `internal/service/feed_service_test.go`.

## Running locally

Prerequisites, first-run setup (including the hardcoded `nick` user every request depends on), and `make hooks` are in `README.md`.

- The server parses `internal/views/*.html` relative to the working directory: run it from the repo root.
- `static/css/output.css` is a build artifact (gitignored); the Docker build generates it.
- The pre-commit hook in `.githooks/` runs gofmt, build, lint, and test.

## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

The five default triage roles plus `done`, recorded as a `Status:` line in each issue file. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
