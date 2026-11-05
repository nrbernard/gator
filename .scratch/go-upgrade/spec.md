# Upgrade Go to 1.27

Status: done

## Problem Statement

gator builds and deploys on Go 1.24.2. Go only supports its two most recent releases (currently 1.27.3 and 1.26.10), so the production binary is compiled with a toolchain that no longer receives security fixes.

Staying on 1.24 also leaves the project stuck behind its own tools. goose v3.28+ requires Go 1.26, so the migration CLI baked into the production image had to be pinned to v3.26.0 after `goose@latest` broke a deploy, and the goose library used by the tests is held at the same old version. The same squeeze will recur with every dependency that raises its minimum Go version.

## Solution

Move the whole project to Go 1.27.3: the module, the production image, local Docker Compose, CI, and the documented prerequisite. Take goose to the latest release (v3.29.0) for both the migration CLI in the image and the library used by the tests, still pinned to an exact version. Ship it as its own deploy, separate from any other change, so a regression is attributable to the upgrade alone.

From the maintainer's perspective nothing about running, testing, or deploying gator changes: the same commands work, the Go toolchain is fetched automatically, and deploys keep running migrations after the new release goes live.

## User Stories

1. As the maintainer, I want production built with a supported Go release, so that the binary receives security fixes.
2. As the maintainer, I want the upgrade to land as a single, separate deploy, so that any regression traces back to the upgrade and nothing else.
3. As the maintainer, I want the target to be the newest Go release, so that the project stays supported for as long as possible before the next upgrade.
4. As the maintainer, I want every place that names a Go version to agree, so that local builds, Docker Compose, CI, and Fly all compile with the same toolchain.
5. As the maintainer, I want the Go version stated once per tool that needs it, with no stale 1.24 references left behind, so that the next upgrade is a mechanical search.
6. As the maintainer, I want the migration CLI in the production image pinned to an exact goose version, so that a new upstream release can never break a deploy again.
7. As the maintainer, I want the goose CLI in the image and the goose library in the module on the same version, so that tests apply migrations the same way production does.
8. As the maintainer, I want the latest goose release, so that migrations get upstream fixes and the pin stops being a workaround.
9. As the maintainer, I want the full production image built and inspected locally before pushing, so that a broken base image, goose pin, or cgo build is caught before it reaches the deploy pipeline.
10. As the maintainer, I want the built image checked for the server binary, the compiled CSS, and the expected goose version, so that I know every build stage produced its output.
11. As the maintainer, I want the existing test suite to pass unchanged on the new toolchain, so that I know the compiler upgrade didn't change behavior.
12. As the maintainer, I want golangci-lint rebuilt against a Go at least as new as the module's target, so that the pre-commit hook keeps working instead of refusing to lint.
13. As the maintainer, I want the pre-commit hook to pass on the upgrade commit itself, so that the upgrade meets the same bar as any other change.
14. As the maintainer, I want my local Go to switch to 1.27.3 automatically in this repo, so that I don't have to install a new Go by hand before building.
15. As the maintainer, I want CI to compile and test with 1.27.3 without workflow changes, so that the deploy pipeline needs no edits for the upgrade.
16. As the maintainer, I want the deploy after the upgrade to complete every job, including the migrations job, so that I know the upgraded goose CLI runs on the Fly machine.
17. As the maintainer, I want a post-deploy check that the live site serves pages, so that I know the upgraded binary actually runs in production.
18. As the maintainer, I want a clear rollback path, so that if production misbehaves I can revert one commit and redeploy the previous toolchain.
19. As a reader of the site, I want pages to keep loading normally through and after the upgrade, so that the upgrade is invisible to me.
20. As a reader, I want feed refreshes to keep working after the upgrade, so that new posts still arrive when I click Refresh.
21. As a future contributor, I want the README's prerequisite to name the Go version the project actually uses, so that I set up the right toolchain on the first try.
22. As a future contributor, I want the reason for pinning goose recorded next to the pin, so that I don't "fix" it back to `@latest`.
23. As an agent working in this repo, I want the version requirements consistent across config and docs, so that I don't pick up a stale version from one of them.
24. As the maintainer, I want the upgrade to leave application code untouched, so that the diff is easy to review and the risk stays in the toolchain alone.

## Implementation Decisions

- **Target version: Go 1.27.3**, the latest patch release. 1.26.10 is also supported, but it falls out of support when Go 1.28 ships (about February 2027); 1.27 lasts until Go 1.29 (about August 2027).
- **One version, everywhere it is named.** The module's `go` directive, the Fly build argument, the Docker Compose build argument, and the README prerequisite all move to 1.27.3 in the same commit. The Dockerfile's own `GO_VERSION` default stays as is because every build path overrides it.
- **Base images.** The builder and goose-builder stages keep using the `golang:<version>-bookworm` images; `golang:1.27.3-bookworm` exists. The runtime stage stays on `debian:bookworm-slim`. Moving to Debian trixie is not part of this upgrade.
- **goose: v3.29.0 for both the library and the CLI.** The module's goose dependency is upgraded from v3.26.0, and the CLI installed in the image's goose-builder stage moves to the same exact version. The CLI is never installed with `@latest`. The comment next to the pin stays and says to keep it in step with the module.
- **Indirect dependencies** move only as far as `go mod tidy` and the goose upgrade require. No other direct dependency is upgraded as part of this work.
- **Local toolchain** comes from Go's automatic toolchain switching (`GOTOOLCHAIN=auto`, the default): any newer-than-1.21 Go downloads 1.27.3 on first use in the repo. No separate install step is documented.
- **CI** needs no workflow change: the runner's preinstalled Go switches to 1.27.3 through the same mechanism when it runs the test step.
- **golangci-lint** must be reinstalled with Go ≥ 1.27, because a golangci-lint built with an older Go refuses to lint a module that targets a newer one. The existing install was built with Go 1.26.
- **No application code changes.** New language features and library idioms available in 1.25–1.27 are not adopted here.
- **Delivery:** one commit containing only the upgrade, pushed on its own once local verification passes. Rollback is reverting that commit and pushing; there are no schema changes, so the database is unaffected either way.

## Testing Decisions

- A good check here exercises the build and runtime exactly as production does and asserts on observable outputs (an image that builds, a binary that serves, a goose that reports the expected version), not on how the toolchain got there.
- **No new test code.** The upgrade changes the toolchain, not behavior, so the existing suite is the regression check.
- **Seam 1, primary: the production image.** Build the full Dockerfile locally with the new Go version, then run the image and confirm the server binary and compiled CSS are present and that goose reports v3.29.0. This is the seam that caught the `goose@latest` failure, and it covers the base image, the goose pin, and the cgo sqlite build together. Prior art: the local verification done for the goose pin fix.
- **Seam 2: the existing test and lint suite.** `go test ./...` and golangci-lint on the new toolchain, through the pre-commit hook locally and the test step in CI. Prior art: the service tests (real migrations applied to in-memory SQLite) and the handler regression test (real templates and handlers through Echo).
- **Seam 3: post-deploy smoke.** The deploy run finishes with every job successful, including migrations, and the live posts page returns 200.

## Out of Scope

- Adopting new Go features or modernizing code (generic iterators, `t.Context()`, `os.Root`, and so on).
- Upgrading Echo, go-sqlite3, or any other direct dependency besides goose.
- Adding `actions/setup-go`, a CI lint job, or a Go version matrix to the workflow.
- Moving the goose CLI to a `tool` directive in the module.
- Pinning the Tailwind install in the image's CSS stage.
- Moving images from Debian bookworm to trixie.
- Fixing the `sqlc` version drift in generated code (generated with v1.28.0, local sqlc is v1.30.0).

## Further Notes

- The Tailwind stage installs whatever npm considers latest, which is the same class of risk that broke the goose stage. It's worth its own ticket.
- A `tool` directive for goose would make the module the single source of truth for the CLI version, and the image could run the module's goose instead of pinning it separately. It's a reasonable follow-up once the project is on a current Go.
- The next upgrade is due when Go 1.29 ships (about August 2027). Following this spec, it's a version-string change plus the same three checks.
