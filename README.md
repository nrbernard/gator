# gator

A single-user RSS reader built around triage: every new post lands in the Queue, and you Keep it for the Reading list or Trash it. Go + Echo, SQLite, htmx, Tailwind.

## Prerequisites

- Go 1.27 with cgo (a C compiler; `go-sqlite3` needs it)
- [goose](https://github.com/pressly/goose) for migrations, the `sqlite3` CLI for `make setup`, [air](https://github.com/air-verse/air) for live reload
- The [standalone Tailwind CLI](https://github.com/tailwindlabs/tailwindcss/releases) (v4), saved as `./tailwindcss` in the repo root
- [golangci-lint](https://golangci-lint.run) v2 for the pre-commit hook: `make install-lint` installs the pinned version

## First run

```sh
make setup   # creates ./data/gator.db, migrates, adds the `nick` user, enables hooks
make run
```

Open http://localhost:8080.

- **The user is hardcoded.** Every request runs as the user named `nick` (`internal/middleware/auth.go`), and only `make setup` creates it: without that row, every page returns a 500.
- **Feeds are fetched on demand.** Posts update when you click **Refresh Posts**; there is no background fetcher.

## Development

- `make test` runs the tests; `make lint` runs golangci-lint.
- After changing `sql/queries/` or `sql/schema/`, run `make generate` to regenerate `internal/database/`. It runs the [sqlc](https://sqlc.dev) version pinned in `go.mod` (`go tool sqlc`), so there is nothing to install.
- New migrations go in `sql/schema/` as the next numbered file.

## Deploy

Pushing to `main` deploys to Fly.io via GitHub Actions: `make check` (format, build, lint, test), then deploy. The container applies pending migrations on start, before the server. The database lives on the `gator_data` volume at `/data/gator.db`.
