# 06: Trash screen and Restore

**What to build:** Completes release 2 (with 04 and 05). The Trash screen, reached from the header menu, lists every trashed Post. Restore takes one tap and returns the Post to where it came from: the Queue, the Reading list, or the Archive. Nothing in the Trash is ever purged, and nothing expires.

**Blocked by:** 04, 05

Status: ready-for-agent

- [ ] State service: Restore returns a Post trashed from the Queue to the Queue, one trashed from the Reading list to the Reading list (keeping its Read mark), and one trashed from the Archive to the Archive
- [ ] State service: Restore is rejected for a Post that isn't in the Trash
- [ ] The Trash screen lists trashed Posts, each with Restore, and the list updates in place over htmx
- [ ] The Trash is reachable from the header menu
- [ ] Nothing purges or expires trashed Posts
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
