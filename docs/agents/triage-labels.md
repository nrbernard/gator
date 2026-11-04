# Triage Labels

The skills speak in terms of five canonical triage roles. This file maps those roles to the status strings used in this repo's issue tracker. Issues are local markdown files, so a "label" is the value of the `Status:` line near the top of each issue file.

| Label in mattpocock/skills | Label in our tracker | Meaning                                  |
| -------------------------- | -------------------- | ---------------------------------------- |
| `needs-triage`             | `needs-triage`       | Maintainer needs to evaluate this issue  |
| `needs-info`               | `needs-info`         | Waiting on reporter for more information |
| `ready-for-agent`          | `ready-for-agent`    | Fully specified, ready for an AFK agent  |
| `ready-for-human`          | `ready-for-human`    | Requires human implementation            |
| `wontfix`                  | `wontfix`            | Will not be actioned                     |
| (repo-specific)            | `done`               | Implemented and merged; closed           |

When a skill mentions a role (e.g. "apply the AFK-ready triage label"), use the corresponding status string from this table.

`done` is not a canonical role. A ticket moves to `done` from `ready-for-agent` or `ready-for-human` once its work is merged. Like `wontfix`, it is a closed state: skills looking for open work should skip `done` and `wontfix` tickets.

Edit the right-hand column to match whatever vocabulary you actually use.
