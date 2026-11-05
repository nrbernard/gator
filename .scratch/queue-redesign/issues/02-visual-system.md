# 02: Queue and shell in the prototype's visual system

**What to build:** Part of release 1 (with 01 and 03). The Queue looks and feels like prototype variant D ("Deck, plain") on branch `prototype/queue-card` (commit 4e70ef6, `make prototype-queue`): a neutral dark page, one light card sized to its content, and Trash and Keep fixed in the thumb zone so they stay in the same place for every Post. Geist type, a single lime accent (kept for the new-posts chip, focus rings, and selection), Lucide icons, and the card entrance animation, which is off under reduced motion. A long description is clamped to a few lines, and "Show more" appears only when the description is actually cut off. A header and its menu reach the Reading list and Feeds. The Reading list and Feeds pages adopt the same system. The prototype's code is not promoted: screens are rebuilt in the app's templates and its Tailwind build, not a CDN. No swipe, no card stack, no keyboard shortcuts.

**Blocked by:** 01

Status: done

- [ ] The Queue matches variant D at phone size and at desktop width (manual check against the prototype)
- [ ] Keep and Trash sit in the same place on screen whatever the length of the card's content
- [ ] A clamped description shows "Show more" and expands; a short one shows no "Show more"
- [ ] The card entrance animation runs on each new card and is off when the user prefers reduced motion
- [ ] Focus rings are visible on every control, using the accent
- [ ] Fonts and icons are served by the app or its build, and the styles come from the Tailwind build, not from a CDN
- [ ] The header and its menu link to the Queue, the Reading list, and Feeds
- [ ] The Reading list and Feeds pages use the same type, colours, and icons
- [ ] `make check` passes
