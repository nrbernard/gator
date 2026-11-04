# 02: Pin sqlc as a module tool and regenerate

**What to build:** Everyone who regenerates the database code uses the same sqlc. The latest sqlc (v1.31.1 when written) is pinned in the module as a tool dependency, `make generate` runs that pinned tool instead of whatever sqlc is on `PATH`, and the generated code is regenerated once with it so the version headers match. It ships as its own commit and deploy; no application behavior changes.

**Blocked by:** 01 (sqlc v1.31.1 requires Go 1.26)

**Status:** ready-for-agent

- [ ] sqlc is declared as a tool dependency in the module at an exact version
- [ ] `make generate` uses the pinned tool and works on a machine with no sqlc installed
- [ ] The regenerated code is committed; the diff shows only what the new sqlc version changes, reviewed as its own commit
- [ ] Running `sqlc diff` with the pinned tool reports no differences
- [ ] The module's `go` directive is unchanged by adding the tool
- [ ] `go test ./...` passes and the pre-commit hook passes on the commit
- [ ] README and CLAUDE.md no longer tell people to install sqlc separately
