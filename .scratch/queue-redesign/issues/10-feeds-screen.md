# 10: Feeds screen: fetch status, and removal that keeps kept Posts

**What to build:** Release 5. Each Feed on the Feeds screen shows when it was last fetched and its last fetch error (recorded by 07), so a broken Feed doesn't fail silently. Removing a Feed stops following it and leaves its row in place, instead of deleting it and cascading to every Post. A removed Feed is not fetched or listed, and its Queue Posts no longer appear. Its Reading list, Archive, and Trash Posts stay, still showing the Feed's name. Re-adding the same URL un-removes it. Adding a Feed still sends all its current items to the Queue on the next fetch. Any schema change keeps working with the previous release's code.

**Blocked by:** 01, 02, 07

Status: done

- [ ] Each Feed shows its last fetch time and, when present, its last error
- [ ] Removing a Feed takes its Queue Posts out of the Queue and its count, and leaves its Reading list, Archive, and Trash Posts in place with the Feed's name
- [ ] Scrape test: a removed Feed is skipped
- [ ] A removed Feed is not listed on the Feeds screen
- [ ] Re-adding a removed Feed's URL un-removes it rather than failing as a duplicate
- [ ] A newly added Feed's existing items appear in the Queue after the next fetch
- [ ] Any migration is additive: the previous release's code still runs against it
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
