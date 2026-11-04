# gator

A single-user RSS reader: follow feeds, read, save, and mark posts read in the browser. Go + Echo, SQLite, htmx, Tailwind.

## Prerequisites

- Go 1.24 with cgo (a C compiler; `go-sqlite3` needs it)
- [goose](https://github.com/pressly/goose) for migrations, [air](https://github.com/air-verse/air) for live reload
- The [standalone Tailwind CLI](https://github.com/tailwindlabs/tailwindcss/releases) (v4), saved as `./tailwindcss` in the repo root
- [sqlc](https://sqlc.dev) if you change SQL queries; [golangci-lint](https://golangci-lint.run) v2 for the pre-commit hook

## First run

```sh
mkdir -p data
export DATABASE_PATH=./data/gator.db
make migrate-up
sqlite3 data/gator.db "INSERT INTO users (id, name) VALUES ('$(uuidgen | tr A-Z a-z)', 'nick');"
make hooks
make run
```

Open http://localhost:8080.

- **The user is hardcoded.** Every request runs as the user named `nick` (`internal/middleware/auth.go`), and nothing creates it: without that row, every page returns a 500.
- **Feeds are fetched on demand.** Posts update when you click **Refresh Posts**; there is no background fetcher.

## Development

- `make test` runs the tests; `make lint` runs golangci-lint.
- After changing `sql/queries/` or `sql/schema/`, run `make generate` to regenerate `internal/database/`.
- New migrations go in `sql/schema/` as the next numbered file.

## Deploy

Pushing to `main` deploys to Fly.io via GitHub Actions: `make check` (format, build, lint, test), then deploy. The container applies pending migrations on start, before the server. The database lives on the `gator_data` volume at `/data/gator.db`.
