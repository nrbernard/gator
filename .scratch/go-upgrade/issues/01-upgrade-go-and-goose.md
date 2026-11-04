# 01: Upgrade Go to 1.27.3 and goose to v3.29.0

**What to build:** Production builds and runs on a supported Go release, and migrations use the latest goose, everything described in `../spec.md`. Every place that names a Go version moves to 1.27.3 together. The goose library and the goose CLI in the production image both move to v3.29.0, with the CLI still pinned to that exact version. Application code is unchanged. It ships as its own commit and deploy.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] The module, the Fly build argument, the Docker Compose build argument, and the README prerequisite all name Go 1.27.3; no 1.24 references remain outside vendored skills
- [ ] The goose library and the goose CLI pinned in the image are both v3.29.0, with the pin's comment intact
- [ ] golangci-lint is reinstalled with Go ≥ 1.27 and the pre-commit hook passes on the upgrade commit
- [ ] `go test ./...` and `make lint` pass on Go 1.27.3
- [ ] The full production image builds locally, and running it shows the server binary, the compiled CSS, and `goose version: v3.29.0`
- [ ] The upgrade is the only change in its deploy
- [ ] The deploy run succeeds on every job, including migrations, and the live posts page returns 200
