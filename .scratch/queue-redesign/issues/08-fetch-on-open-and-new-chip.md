# 08: Fetch on open, the "N new" chip, and check now

**What to build:** Completes release 3 (with 07). Loading the Queue starts a background htmx request that runs the scrape, so the Queue is fresh whenever the user looks, with no button. The existing hourly per-Feed limit still decides which Feeds are fetched, so opening the app often costs nothing. Queue Posts created after the page's cutoff (from 03) are new: they are counted in a lime "N new" chip in the header and are not shown until the user taps it, so the card never changes under a tap. Tapping the chip moves the cutoff forward and shows them. A manual "check now" replaces today's Refresh button. Trash all keeps sparing Posts after the current cutoff.

**Blocked by:** 03

Status: done

- [ ] Opening the Queue triggers a background scrape without delaying the first card
- [ ] Posts fetched after the page loaded don't change the current card or the count; the chip shows how many there are
- [ ] Tapping the chip moves the cutoff forward, and the new Posts join the Queue in newest-first order and in the count
- [ ] With no new Posts, there is no chip
- [ ] "Check now" runs the scrape on demand and the results arrive behind the chip; the old Refresh button is gone
- [ ] Trash all after tapping the chip includes the Posts the chip revealed and spares any fetched since
- [ ] Handler tests cover the chip's count and moving the cutoff
- [ ] Matches the visual system from 02 at phone and desktop width (manual check)
- [ ] `make check` passes
