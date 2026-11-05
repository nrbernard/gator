# 09: Search across every place

**What to build:** Release 4. Search keeps today's title-and-description match and covers Posts in every place: the Queue, the Reading list, the Archive, and the Trash. Each result is labelled with its place and offers that place's actions (for example Keep and Trash for a Queue Post, Restore for a trashed one), so the user can act on a Post without leaving search. Actions go through the state service like everywhere else.

**Blocked by:** 06

Status: ready-for-agent

- [ ] A search finds matching Posts in each of the four places
- [ ] Each result shows its place and offers that place's actions, and an action updates the result in place
- [ ] Search no longer reads `post_saves` or `post_reads`
- [ ] Test: Posts seeded in each place are all found, each labelled with the right place
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
