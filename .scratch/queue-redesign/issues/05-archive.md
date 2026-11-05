# 05: Archive

**What to build:** Part of release 2 (with 04 and 06). Each Reading list Post gains an Archive action. The Archive screen, reached from the header menu, lists archived Posts with the most recently archived first. From it the user can move a Post back to the Reading list to reread it, or trash it.

**Blocked by:** 04

Status: done

- [ ] State service: Archive moves a Reading list Post to the Archive, and is rejected for a Queue Post
- [ ] State service: Un-archive moves an archived Post back to the Reading list
- [ ] State service: Trash moves an archived Post to the Trash with the Archive as its origin
- [ ] The Archive screen lists archived Posts, most recently archived first, each with Move to Reading list and Trash
- [ ] The Archive is reachable from the header menu
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
