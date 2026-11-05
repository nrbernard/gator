# 11: Installable on the phone

**What to build:** Part of release 6 (with 12). The user can add gator to their phone's home screen and have it open full-screen, so triage feels like an app: a web app manifest, an icon, and the meta tags for a full-screen launch, styled to match the dark page from 02. No service worker and no offline mode.

**Blocked by:** 02

Status: done

- [ ] The manifest is served and linked from every page, with a name, the icon, a start URL of the Queue, and a standalone display
- [ ] The icon replaces the emoji favicon and is served at the sizes the home screen needs
- [ ] Added to an iPhone home screen, gator opens full-screen to the Queue (manual check)
- [ ] No service worker is registered
- [ ] `make check` passes
