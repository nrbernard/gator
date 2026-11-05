# 03: Open, trash all, and the empty Queue

**What to build:** Completes release 1 (with 01 and 02). Opening a Queue Post Keeps it and marks it Read, because clicking through counts as wanting it. The Queue page records the moment it loaded as its cutoff. "Trash all" tells the user how many Posts it will trash, asks first, and trashes only Queue Posts created at or before the cutoff, so Posts fetched after the page loaded are spared. An empty Queue says the user is caught up, shows when gator last fetched, and offers the Reading list. Posts stay in the Queue until the user decides; nothing leaves it on its own.

**Blocked by:** 01

Status: ready-for-agent

- [ ] State service: Open on a Queue Post moves it to the Reading list marked Read
- [ ] State service: trash all moves every Queue Post created at or before the cutoff to the Trash, with the Queue as its origin, and leaves Posts created after the cutoff in the Queue
- [ ] Opening a Post from the Queue card goes to the Post's site and the Post is then on the Reading list, Read
- [ ] Trash all shows the number of Posts it will trash and does nothing until the user confirms
- [ ] Handler tests cover the trash-all confirmation and the trash itself, rendering through the real templates
- [ ] The empty Queue says the user is caught up, shows the last fetch time, and links to the Reading list
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
