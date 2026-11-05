# 01: Upgrade Go to 1.27.3 and goose to v3.29.0

**What to build:** Production builds and runs on a supported Go release, and migrations use the latest goose, everything described in `../spec.md`. Every place that names a Go version moves to 1.27.3 together. The goose library and the goose CLI in the production image both move to v3.29.0, with the CLI still pinned to that exact version. Application code is unchanged. It ships as its own commit and deploy.

**Blocked by:** None (can start immediately)

**Status:** done

- [x] The module, the Fly build argument, the Docker Compose build argument, the Dockerfile default, and the README prerequisite all name Go 1.27.3 (`make check` fails until they agree); no 1.24 references remain outside vendored skills
- [x] The goose library and the goose CLI pinned in the image are both v3.29.0, with the pin's comment intact
- [x] golangci-lint is reinstalled with Go ≥ 1.27 and the pre-commit hook passes on the upgrade commit
- [x] CI's pinned golangci-lint is also built with Go ≥ 1.27 (a newer pinned release, or the install forced onto the 1.27 toolchain), so `make check` in CI doesn't refuse to lint
- [x] `go test ./...` and `make lint` pass on Go 1.27.3
- [x] The full production image builds locally, and running it shows the server binary, the compiled CSS, and `goose version: v3.29.0`
- [x] The upgrade is the only change in its deploy
- [x] The deploy run succeeds on every job, including migrations, and the live posts page returns 200

## Comments

Implemented in a2b7344 on branch `go-upgrade`. Local checks passed: `make check` on go1.27.3, and the full image built with `/app/main`, `/app/static/css/output.css` present and `goose -version` printing `goose version: v3.29.0`. golangci-lint stays at v2.15.0 (no release requires Go 1.27 yet), built with go.mod's Go; review fix 1eefd09 moved that install into `make install-lint`.

Left for the maintainer: push a2b7344 alone to `main`, confirm every deploy job passes (including migrations) and the live posts page returns 200, then set `Status: done`. Rollback: revert a2b7344 and push.

Deployed 2026-11-05: a2b7344 pushed alone to `main` (d4a59e3..a2b7344). Deploy run 41947986116 succeeded on every step and Fly release v25 is complete. On first boot the machine ran goose (`no migrations to run. current version: 10`) and started the server, and `https://gator.fly.dev/posts` returned 200. One health check failed in the second before the server was listening, then passed.
