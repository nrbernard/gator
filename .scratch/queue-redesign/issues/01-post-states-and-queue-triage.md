# 01: Post states, the backfill, and Keep/Trash on the Queue

**What to build:** The tracer bullet for the redesign in `../spec.md`, and the first ticket of release 1 (with 02 and 03, merged to `main` together). Every Post is in exactly one place (`docs/adr/0001-exclusive-post-states.md`). A new per-(user, Post) state table records the place (Reading list, Archive, or Trash), where a trashed Post came from, the Read mark, and when the state last changed; a Post with no row is in the Queue. The migration that creates it backfills from `post_saves` and `post_reads`, and those tables stay in place. A state service owns every move; this ticket gives it Keep and Trash and the rejection of invalid moves. `/` becomes the Queue: one Post at a time, newest published first, showing its title, Feed, published time, and description, with the number of Posts left, and Keep and Trash buttons that swap in the next card over htmx. A basic Reading list page lists the kept Posts so nothing kept disappears. The Unread, Saved, and All tabs, and the heart and check-mark buttons, are removed. Plain styling is fine; 02 applies the visual system.

Interim behaviour until later tickets: today's Refresh button stays on the Queue as the only way to fetch (08 replaces it), and search keeps matching title and description across all Posts but shows no per-result actions (09 adds them).

**Blocked by:** None (can start immediately)

Status: done

- [ ] The state migration is additive: the previous release's code still runs against the migrated schema
- [ ] Backfill test: migrate to the version before the state table, seed Posts with each combination of saved and read, apply the rest, and assert that saved Posts are on the Reading list with their Read mark, unsaved read Posts are in the Trash with the Queue as their origin, and unsaved unread Posts are in the Queue
- [ ] State service tests against a real migrated database: Keep moves a Queue Post to the Reading list, Trash moves it to the Trash with the Queue as its origin, and moves that don't apply (for example Keep on a Post already on the Reading list) are rejected by the service
- [ ] `/` shows the Queue instead of redirecting to `/posts`, one Post at a time, newest published first, with the count of Queue Posts left
- [ ] Keep and Trash are ordinary buttons; each request moves the Post and responds with the next card (or an empty Queue)
- [ ] Handler tests drive Keep and Trash through Echo and the real templates, and check where the Post ends up and what the response shows
- [ ] A Reading list page lists the kept Posts and is linked from the Queue
- [ ] The Unread, Saved, and All tabs and the Saved/Read toggle buttons are gone, along with their routes
- [ ] Refresh still fetches, and search still finds Posts by title and description
- [ ] `make check` passes
