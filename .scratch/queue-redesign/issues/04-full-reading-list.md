# 04: Full Reading list

**What to build:** Part of release 2 (with 05 and 06). The Reading list shows the Posts the user kept, unread first and then most recently kept. Each Post offers Open, Read/Unread, and Trash (05 adds Archive). Opening a Reading list Post marks it Read. The user can mark a read Post unread again, to undo an accidental open. Read Posts stay on the Reading list until the user archives or trashes them.

**Blocked by:** 01

Status: ready-for-agent

- [ ] State service: Open on a Reading list Post marks it Read and leaves it on the Reading list
- [ ] State service: toggling Read flips the mark on a Reading list Post and is rejected for Posts in other places
- [ ] State service: Trash moves a Reading list Post to the Trash with the Reading list as its origin
- [ ] The Reading list orders unread Posts first, then by when they were kept, most recent first
- [ ] Each Post offers Open, Read/Unread, and Trash, and each action updates the list in place over htmx
- [ ] A read Post stays on the Reading list
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
