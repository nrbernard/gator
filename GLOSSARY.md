# Gator

A single-person RSS reader built around triage: every new post is either kept for reading or thrown away, so the reading list stays curated.

## Sources

**Feed**:
An RSS or Atom source the user subscribes to.
_Avoid_: Source, subscription, channel

**Remove a Feed**:
To stop following a Feed. gator no longer fetches or lists it, and its Queue Posts leave the Queue, but its Posts on the Reading list, in the Archive, or in the Trash stay. Adding its URL again brings it back.
_Avoid_: Delete, unsubscribe, unfollow

**Post**:
One item published by a Feed. Every Post is in exactly one Place.
_Avoid_: Article, item, entry, story

**Place**:
Where a Post is: the Queue, the Reading list, the Archive, or the Trash. A Post is in exactly one Place at a time.
_Avoid_: State, status, folder, location

## Triage

**Queue**:
The Posts the user has not yet decided on. Every newly fetched Post starts here; a Queue Post is undecided, not unread.
_Avoid_: Inbox, feed, unread, stream

**Triage**:
Deciding a Queue Post's fate from its title, Feed, and description: Keep it or Trash it.
_Avoid_: Sorting, review

**Keep**:
To move a Post from the Queue to the Reading list.
_Avoid_: Save, like, favourite, bookmark

**Open**:
To click through to a Post's site from the Queue or the Reading list. Opening a Queue Post Keeps it marked Read; opening a Reading list Post marks it Read.
_Avoid_: View, visit, read (as the action)

**Cutoff**:
The moment up to which a Queue page shows Posts: the end of the last complete second when the page loaded, or when the user last tapped the New-posts chip. Posts gator fetched after it are new; trash all trashes only Posts fetched by it.
_Avoid_: Timestamp, since, watermark

**New-posts chip**:
The "N new" chip on the Queue counting the Posts fetched after the cutoff. Tapping it moves the cutoff forward and shows them, so the card never changes under the user.
_Avoid_: Badge, notification, banner

**Trash**:
The Posts the user has thrown away, which can still be restored to wherever they came from, except a Post trashed from the Queue whose Feed has since been removed; as a verb, to move a Post there. Nothing in the Trash is ever purged.
_Avoid_: Delete, dismiss, discard, skip

## Reading

**Reading list**:
The Posts the user has kept and not yet archived; the curated replacement for the old list of all posts.
_Avoid_: Saved posts, favourites, read later

**Read**:
A mark on a Reading list Post that the user has read it, which the user can clear. A read Post stays on the Reading list until the user archives or trashes it.
_Avoid_: Seen, done, opened

**Archive**:
The Posts the user has finished with and taken off the Reading list, kept so they can be found again; as a verb, to move a Post there.
_Avoid_: Done, history, completed
