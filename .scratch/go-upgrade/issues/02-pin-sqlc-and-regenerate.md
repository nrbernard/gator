# 02: Pin sqlc as a module tool and regenerate

**What to build:** Everyone who regenerates the database code uses the same sqlc. The latest sqlc (v1.31.1 when written) is pinned in the module as a tool dependency, `make generate` runs that pinned tool instead of whatever sqlc is on `PATH`, and the generated code is regenerated once with it so the version headers match. It ships as its own commit and deploy; no application behavior changes.

**Blocked by:** 01 (sqlc v1.31.1 requires Go 1.26)

Status: done

- [x] sqlc is declared as a tool dependency in the module at an exact version
- [x] `make generate` uses the pinned tool and works on a machine with no sqlc installed
- [x] The regenerated code is committed; the diff shows only what the new sqlc version changes, reviewed as its own commit
- [x] Running `sqlc diff` with the pinned tool reports no differences
- [x] The module's `go` directive is unchanged by adding the tool
- [x] `go test ./...` passes and the pre-commit hook passes on the commit
- [x] README and CLAUDE.md no longer tell people to install sqlc separately

## Comments

Implemented on branch `go-upgrade`: a07343d pins sqlc v1.31.1 as a `tool` and points `make generate` at `go tool sqlc`; da410a1 is the regeneration, which only changes the `sqlc v1.28.0` → `v1.31.1` header in the seven generated files. Ships as its own deploy after ticket 01's.
