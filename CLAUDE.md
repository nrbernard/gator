# gator

RSS reader: Go + Echo, SQLite via sqlc, server-rendered `html/template` views styled with Tailwind, deployed to Fly.io. Commands live in the `Makefile`.

## Deploys

A push to `main` deploys to production (`.github/workflows/fly-deploy.yml`): tests, then `flyctl deploy`, then `goose up` on the machine. Work on a branch; merging to `main` is a release.

Migrations run *after* the new binary is live, so for a short window new code serves against the old schema. Make each schema change backward compatible: add a column in one release, start reading it in the next.

## Database

- Schema is goose migrations in `sql/schema/NNN_name.sql`, numbered sequentially. Queries are in `sql/queries/`.
- `internal/database/` is sqlc output. Change the SQL in `sql/queries/` or `sql/schema/`, then run `make generate` and commit the regenerated files with it.
- `make migrate-*` needs `DATABASE_PATH` set; locally that's `./data/gator.db`.
- `internal/service/feed_service_test.go` builds its in-memory schema by hand in `setupTestDB`. When a migration touches `users`, `feeds`, or `posts`, update that schema too.

## Running locally

- `make run` starts Tailwind in watch mode and `air` for live reload. It expects the standalone `./tailwindcss` binary at the repo root (gitignored), plus `air` and `goose` on `PATH`.
- The server parses `internal/views/*.html` relative to the working directory: run it from the repo root.
- `static/css/output.css` is a build artifact (gitignored); the Docker build generates it.
- `go-sqlite3` needs cgo.

## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

The five default triage roles plus `done`, recorded as a `Status:` line in each issue file. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
