# 07: Per-Feed fetch errors don't stop the scrape

**What to build:** Part of release 3 (with 08). Today the first fetch error aborts the whole scrape, so one dead site keeps every other Feed from fetching. Instead, a Feed's failure is recorded on that Feed (its last error and when it happened) and the scrape carries on with the rest. A successful fetch clears the Feed's error. The schema change is additive. Showing the error on the Feeds screen is ticket 10.

**Blocked by:** None (can start immediately)

Status: done

- [ ] Scrape test: with one failing Feed and one working Feed, the working Feed's Posts are created and the failing Feed records its error and the time
- [ ] Scrape test: a Feed with a recorded error that then fetches successfully has its error cleared
- [ ] The hourly per-Feed fetch limit and the 429 handling are unchanged
- [ ] The migration is additive: the previous release's code still runs against it
- [ ] `make check` passes
