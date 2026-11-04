# 03: Reject commits with stale generated code

**What to build:** A commit that changes the SQL queries or schema without regenerating the database code is stopped before it lands. When a commit touches SQL, the pre-commit hook checks the generated code with the pinned sqlc and rejects the commit with a message saying to run `make generate`.

**Blocked by:** 02

**Status:** ready-for-agent

- [ ] The hook runs the generated-code check only when the commit touches SQL queries or schema, so other commits are no slower
- [ ] Changing a query without regenerating makes the hook fail with a message naming `make generate`
- [ ] Changing a query and regenerating passes the hook
- [ ] The check uses the pinned sqlc from ticket 02, not a separately installed one
- [ ] CLAUDE.md's description of the hook's checks includes the new one
