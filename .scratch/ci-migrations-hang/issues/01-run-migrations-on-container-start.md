# 01: Run migrations on container start instead of from CI

**What to build:** Deploys finish reliably, and migrations are applied before new code serves traffic. The container runs `goose up` itself on start, then the server, the same way Docker Compose already starts it, and the separate CI migrations job goes away. A failing migration keeps the new machine from starting, so the deploy's health check fails instead of the old code running against a half-migrated schema.

**Blocked by:** None (can start immediately)

**Status:** needs-triage

Needs the maintainer's go-ahead on the approach (see Decision below) before it's `ready-for-agent`.

- [x] The production container applies pending migrations before starting the server, and exits non-zero if they fail
- [x] The CI "Run migrations" job, and the "Get machine ID" step that only fed it, are removed
- [x] Every remaining CI job has a `timeout-minutes`, so a hang fails in minutes instead of running for GitHub's 6-hour default
- [x] `setup-flyctl` is pinned to an exact flyctl version instead of `latest`
- [x] The full production image builds locally, and starting it against an empty database applies every migration before the server logs that it's listening
- [x] CLAUDE.md's Deploys section and the README's Deploy paragraph describe the new ordering; the "migrations run after the new binary is live" caveat is gone
- [ ] A deploy succeeds on every job, and the live posts page returns 200

## Diagnosis

Deploy run 41888107789 (commit `df3037e`, 2026-11-04) built and deployed fine, then hung in "Run migrations" for ~50 minutes until it was cancelled. That step runs `flyctl console --machine <id> --command "make migrate-up"`. No migrations were pending: `sql/schema/`, the workflow, and `fly.toml` were all unchanged since the last successful deploy (`955dd66`, 2025-10-16).

Timeline, from the CI log and `flyctl logs -a gator`:

| Time (UTC) | Event |
|---|---|
| 18:35:22 | CI "Start machine": `148eee9b229928 has been started` |
| 18:35:24 | Machine boots; app listening on `[::]:8080`; SSH listening on port 22 |
| 18:35:24 | CI `flyctl console` starts and prints nothing, ever |
| 18:40:40 | Fly proxy auto-stops the machine ("excess capacity … 0 out of 1 machines left running") |
| ~19:25 | Run cancelled |

Findings:

- **Wrong machine: ruled out.** The app has exactly one machine, `148eee9b229928`, and that's the one the workflow picked.
- **Auto-stop: confirmed, but not the first cause.** `fly.toml` has `auto_stop_machines = 'stop'` and `min_machines_running = 0`. The machine stopped 5 minutes into a console session that had already been silent, so it guaranteed the hang never ended but probably didn't start it.
- **New flyctl: likely, unconfirmed.** `setup-flyctl@master` with `version: latest` installed flyctl **v0.5.3** (released 2026-11-03). The last success ran a 2025 flyctl. Release notes from v0.4.69 to v0.5.3 mention nothing about `console`, SSH, or stdin that explains a silent hang.
- **Not tested:** running the same console command by hand with local flyctl v0.4.1 vs v0.5.3. The agent's permission system blocks remote shell writes against production. To test it yourself: `flyctl console -a gator --machine 148eee9b229928 --command "make migrate-up" </dev/null`. If it prints `goose: no migrations to run`, v0.4.1 works, which points at v0.5.3.

## Decision

Running migrations on container start removes the hanging step whatever its cause, and also fixes the ordering gotcha in CLAUDE.md (migrations ran *after* new code went live).

- **Cost:** `goose up` runs on every boot, which happens often because of auto-stop. With nothing pending it takes milliseconds.
- **Ruled out: Fly's `release_command`.** Release machines don't mount volumes, and the SQLite database lives on the `gator_data` volume at `/data`.
- **If you reject the approach:** keep the CI job, but still pin flyctl and add `timeout-minutes`, and keep the machine from auto-stopping during the step.
