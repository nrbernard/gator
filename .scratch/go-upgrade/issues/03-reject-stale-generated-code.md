# 03: Reject commits with stale generated code

**What to build:** A commit that changes the SQL queries or schema without regenerating the database code is stopped before it lands. When a commit touches SQL, the pre-commit hook checks the generated code with the pinned sqlc and rejects the commit with a message saying to run `make generate`.

**Blocked by:** 02

Status: done

- [x] The hook runs the generated-code check only when the commit touches SQL queries or schema, so other commits are no slower
- [x] Changing a query without regenerating makes the hook fail with a message naming `make generate`
- [x] Changing a query and regenerating passes the hook
- [x] The check uses the pinned sqlc from ticket 02, not a separately installed one
- [x] CLAUDE.md's description of the hook's checks includes the new one

## Comments

Implemented on branch `go-upgrade` in ac3c2a5, extended by review fix 1eefd09. The check runs when the commit stages `sql/queries/`, `sql/schema/` (including deletions), `sqlc.yaml`, `internal/database/` or `go.mod`. It regenerates into a temp dir with the pinned sqlc and diffs against `internal/database/`, because `sqlc diff` misses generated files left behind by a deleted query. CI doesn't run it, so `--no-verify` can still bypass it.
