# Redesign gator around the Queue

Status: ready-for-agent

Vocabulary is defined in `GLOSSARY.md`; the core model decision is `docs/adr/0001-exclusive-post-states.md`.

## Problem Statement

gator shows every fetched post in one long list, with two independent flags on each: Saved (the heart) and Read (the check mark, also set by opening a post). Nothing ever asks the user to decide about a post, so the list grows without limit, "unread" mostly means "not looked at yet", and the Saved tab is the only place that resembles a curated list. Feeds update only when the user clicks Refresh. The interface is crude and utilitarian: a single page of tabs, emoji buttons, and a search box.

## Solution

Make triage the centre of the app. Every newly fetched Post lands in the **Queue**, which is the home screen. The Queue shows one Post at a time, with its title, Feed, published time, and description. The user taps **Keep** to send it to the **Reading list** or **Trash** to throw it away. The Reading list replaces the old list of all posts: it is what the user chose to read. Finished Posts go to the **Archive**, where search can still find them. Nothing is ever purged; anything trashed can be restored to where it came from.

Every Post is in exactly one of the four places. Fetching starts automatically when the Queue opens, and new Posts wait behind a chip until the user asks for them, so the screen never shifts under a tap.

The visual direction is prototype variant D ("Deck, plain") on branch `prototype/queue-card` (commit 4e70ef6; run `make prototype-queue`): a dark page, one light card sized to its content, and two plain buttons fixed in the thumb zone. There is no swipe, no card stack, and nothing that feels like a dating app: simple and fast.

## User Stories

### Triage

1. As the reader, I want the Queue to be the first screen I see, so that deciding about new Posts is the main thing the app asks of me.
2. As the reader, I want the Queue to show one Post at a time, with its title, Feed, published time, and description, so that I can decide from what's in front of me.
3. As the reader, I want to tap Keep to put a Post on my Reading list, so that I can save it for real reading.
4. As the reader, I want to tap Trash to throw a Post away, so that it never comes back to the Queue.
5. As the reader, I want Keep and Trash to stay in the same place on screen for every Post, so that I can triage quickly without looking for the buttons.
6. As the reader, I want the next Post to appear as soon as I decide, so that triage feels fast.
7. As the reader, I want the Queue newest first, so that fresh Posts aren't stuck behind an old backlog.
8. As the reader, I want to see how many Posts are left in the Queue, so that I know how much triage remains.
9. As the reader, I want a long description clamped to a few lines with a way to expand it, so that most decisions take a glance but a close call doesn't force me to open the site.
10. As the reader, I want "Show more" to appear only when the description is actually cut off, so that the card has no controls that do nothing.
11. As the reader, I want opening a Queue Post to Keep it and mark it Read, so that clicking through counts as wanting it.
12. As the reader, I want to trash the whole Queue in one action after time away, so that a backlog never becomes a chore.
13. As the reader, I want that action to tell me how many Posts it will trash and ask first, so that I don't throw away more than I meant to.
14. As the reader, I want that action to leave alone any Posts I haven't been shown yet, so that it only trashes what I have actually seen.
15. As the reader, I want Posts to stay in the Queue until I decide, so that nothing disappears on its own.
16. As the reader, I want an empty Queue to say I'm caught up, show when it last fetched, and offer the Reading list, so that I know I'm done and that fetching works.

### Fetching

17. As the reader, I want opening the Queue to fetch new Posts in the background, so that it's fresh whenever I look, without a button.
18. As the reader, I want newly fetched Posts to wait behind a "N new" chip until I tap it, so that the card doesn't change under my thumb mid-decision.
19. As the reader, I want a manual "check now", so that I can fetch on demand.
20. As the reader, I want one broken Feed not to stop the others from fetching, so that a dead site doesn't empty my Queue.

### Reading list

21. As the reader, I want the Reading list to show the Posts I kept, unread first and then most recently kept, so that what's still to read is on top.
22. As the reader, I want opening a Reading list Post to mark it Read, so that I don't have to mark it by hand.
23. As the reader, I want to mark a read Post unread again, so that an accidental open is easy to undo.
24. As the reader, I want each Reading list Post to offer Open, Read/Unread, Archive, and Trash, so that I can manage the list from the list.
25. As the reader, I want read Posts to stay on the Reading list until I archive or trash them, so that I decide when a Post is finished.

### Archive and Trash

26. As the reader, I want archived Posts kept, most recently archived first, so that I can find what I finished recently.
27. As the reader, I want to move an archived Post back to the Reading list, so that I can reread it.
28. As the reader, I want to trash an archived Post, so that the Archive holds only what is worth keeping.
29. As the reader, I want to see the Trash and restore any Post from it, so that a mis-tap is never permanent.
30. As the reader, I want a restored Post to go back where it came from (the Queue, the Reading list, or the Archive), so that restoring takes one tap and no decision.
31. As the reader, I want nothing in the Trash to be purged, so that I can always get a Post back.

### Search

32. As the reader, I want to search Posts in every place, the Queue included, so that I can find anything I've seen.
33. As the reader, I want each result to show where the Post is and offer that place's actions, so that I can act on it without leaving search.

### Feeds

34. As the reader, I want to add and remove Feeds, as today.
35. As the reader, I want each Feed to show when it was last fetched and its last fetch error, so that a broken Feed doesn't fail silently.
36. As the reader, I want a new Feed's existing items to enter the Queue, so that I see what the Feed is like.
37. As the reader, I want removing a Feed to clear its Posts from the Queue but keep the ones I kept, archived, or trashed, so that removing a Feed doesn't erase what I saved from it.

### Existing data and the phone

38. As the reader, I want my saved Posts to be on the Reading list after the redesign, with their Read marks, so that nothing I saved is lost.
39. As the reader, I want my old unsaved, unread Posts in the Queue and my old unsaved, read Posts in the Trash, so that I start from a sensible place and can still restore anything.
40. As the reader, I want to add gator to my phone's home screen and have it open full-screen, so that triage feels like an app.

### Maintainer

41. As the maintainer, I want the redesign to ship in slices, each its own deploy, so that a regression traces to one slice.
42. As the maintainer, I want each schema change to work with the previous release's code, so that a rollback after a migration doesn't break the app.

## Implementation Decisions

- **Stack unchanged:** Go, Echo, SQLite via sqlc, server-rendered `html/template`, htmx, and Tailwind. Triage actions are htmx requests that swap in the next card. Keep and Trash use ordinary buttons; there is no swipe and there are no keyboard shortcuts.
- **One state per Post, per user.** A new table holds a row per (user, Post) with the state (Reading list, Archive, or Trash), the place a trashed Post came from, the Read mark, and when the state last changed. A Post with no row for the user is in their Queue, so a newly fetched Post needs no extra write to be queued. The data stays keyed by user even though there is one user.
- **Move rules live in one place**, a service that owns every state change: Keep, Trash, Open (Keep and mark Read when the Post is in the Queue, mark Read on the Reading list), Archive, Un-archive, toggle Read, and Restore. An invalid move, such as archiving a Queue Post, is rejected there rather than in handlers or templates.
- **Backfill:** the migration that creates the state table fills it from `post_saves` and `post_reads`. Saved Posts go to the Reading list with their Read mark. Unsaved, read Posts go to the Trash, with the Queue recorded as where they came from. Unsaved, unread Posts get no row, so they are in the Queue. `post_saves` and `post_reads` stay until a later release has stopped reading them; then a migration drops them.
- **Backward compatibility:** the state table is additive, so the previous release still runs against the new schema. If a deploy rolls back after the backfill, triage done in the new release is invisible to the old code, and saves made under the old code during the rollback are not re-migrated. That is acceptable for one user.
- **Queue order** is newest published first. The count is the number of Queue Posts already shown.
- **Fetching on open:** loading the Queue triggers a background htmx request that runs the existing scrape. The scrape already fetches only Feeds not fetched in the last hour, and that per-Feed limit stays to be polite to publishers, so opening the app often costs nothing. The ~15-minute threshold discussed during design is unnecessary: the hourly per-Feed limit already covers it.
- **New-posts cutoff:** the Queue page records the moment it loaded. Queue Posts created after that moment are "new": they are counted in the header chip and shown only when the chip is tapped, which moves the cutoff forward. Trashing the whole Queue applies only to Queue Posts created at or before the cutoff.
- **Fetch errors:** one Feed's failure is recorded on that Feed (last error and when it happened) and the scrape continues with the rest. Today the first error aborts the whole scrape. A successful fetch clears the error.
- **Removing a Feed stops following it and leaves the row in place,** instead of deleting it, which today cascades to every Post. A removed Feed is not fetched or listed, and its Queue Posts no longer appear. Its Reading list, Archive, and Trash Posts stay, still showing the Feed's name. Re-adding the same URL un-removes it, since feed URLs are unique.
- **New Feeds** send all their current items to the Queue on the next fetch, as today.
- **Search** extends today's title-and-description match to every place, and each result is labelled with its place.
- **Navigation:** the Queue is at `/` (replacing the redirect to `/posts`). The Reading list, Archive, Trash, and Feeds are reached from the header and its menu.
- **Visual system:** take the prototype's variant D as the reference: Geist type, the neutral dark page with a light card, one lime accent for the new-posts chip, focus rings, and selection, Lucide icons, and the card entrance animation (off under reduced motion). The prototype's code is not promoted; screens are rebuilt in the app's templates and Tailwind build. The other screens follow the same system; there is no separate design pass for each.
- **Installable:** a web app manifest, an icon, and the meta tags for a full-screen launch from the home screen. No service worker and no offline mode.
- **Delivery slices,** each one deploy:
  1. Post states, the backfill, and the Queue as the home screen (Keep, Trash, Open, count, trash all, empty state), plus a basic Reading list so kept Posts stay visible. This replaces today's Unread, Saved, and All tabs, which stop making sense once the flags no longer drive them.
  2. The full Reading list (read toggle, ordering, Archive and Trash actions), the Archive, and the Trash with restore.
  3. Fetching on open, the new-posts chip and cutoff, and per-Feed fetch errors that don't stop the scrape.
  4. Search across every place.
  5. The Feeds screen: last fetch and last error, and removal that keeps kept Posts.
  6. Installability, and the migration that drops `post_saves` and `post_reads`.

## Testing Decisions

- Good tests here drive the app through its public seams, service calls or HTTP requests against a real migrated database, and assert on where Posts end up and what the screens show, not on SQL or template internals.
- **The state service is the main seam.** Test every move and every rejected move against an in-memory SQLite database with the real migrations applied (`setupTestDB` in `internal/service/feed_service_test.go` is the prior art). That includes Open from the Queue keeping and marking Read, Restore returning a Post to where it came from, and trash-all sparing Posts created after the cutoff.
- **The backfill gets its own test:** migrate up to the version before the state table, seed Posts with each combination of saved and read, apply the rest, and assert each Post's place.
- **Fetching:** extend the existing scrape tests (`TestFeedService_ScrapeFeeds_*`) so that a failing Feed records its error and the others still fetch, and a removed Feed is skipped.
- **Handlers:** request-level tests for the Queue's triage actions and the trash-all confirmation, rendering through the real templates and Echo as `internal/handler/post_handler_test.go` does.
- **Visual checks** are manual, at phone size and desktop width, against the prototype, on each slice that changes a screen.

## Out of Scope

- Swipe gestures and keyboard shortcuts.
- An Undo for the last triage action (the Trash is the way back).
- Passwords, accounts, or any change to the hardcoded `nick` user.
- Reader view (full post text inside gator).
- Feed rules (auto-Keep or auto-Trash by Feed or keyword).
- Folders, tags, or filtering the Queue by Feed.
- OPML import and export.
- Bulk actions other than trashing the whole Queue.
- Purging the Trash, or any automatic expiry.
- Offline mode or a service worker.
- Multi-user support.

## Further Notes

- The prototype on `prototype/queue-card` contains three rejected directions (A Broadsheet, B Deck, C Conveyor) for reference. B's swipe, card stack, and KEEP/TRASH stamps were rejected as too much like a dating app.
- There is no PRODUCT.md or DESIGN.md yet. The maintainer plans a more in-depth design pass later; `/impeccable init` would capture product context first.
- Fly scales the machine to zero, so a scheduled background fetch would need outside infrastructure. Fetching on open avoids that. If the Queue ever needs to fill while the app is closed, a scheduled job that wakes the machine to fetch is the follow-up.
- `scripts/mark_read.go` and `scripts/fix_post_timestamps.sql` write to the old tables and become obsolete with the final slice.
