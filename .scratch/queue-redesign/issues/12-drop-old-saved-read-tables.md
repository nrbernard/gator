# 12: Drop `post_saves` and `post_reads`

**What to build:** Completes release 6 (with 11). Once a release that no longer reads `post_saves` or `post_reads` is live, a migration drops both tables, along with their queries and generated code. `scripts/mark_read.go` and `scripts/fix_post_timestamps.sql` write to the old tables and are removed. The production release before this one must not read either table, so a rollback to it still works after the drop.

**Blocked by:** 01, 02, 03, 04, 05, 06, 07, 08, 09, 10, 11

Status: ready-for-agent

- [ ] No code outside migrations reads or writes `post_saves` or `post_reads`, and the release currently in production doesn't either
- [ ] A new migration drops both tables; the backfill test still passes, since it migrates up from before the state table
- [ ] Their queries and generated code are gone (`make generate`), and the generated-code check passes
- [ ] The obsolete scripts are removed
- [ ] `make check` passes
