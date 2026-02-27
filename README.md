# faz

`faz` is a lightweight local task tracker for agent workflows.
It uses Go + Cobra for the CLI and SQLite for issue storage and graph relations.
It is inspired by [Beads](https://github.com/steveyegge/beads).

## Why faz

- Personal local-first task tracking for daily coding workflows
- Minimal friction for humans and AI agents
- No GitHub integration or remote coupling
- Explicit lifecycle and dependency commands with practical output
- Fast project context recovery from the terminal

## Quick start

```bash
go install .
faz init
faz onboard
```

## Core commands

```bash
faz recap
faz create "Checkout revamp" --type epic --priority 1 --description "Improve checkout"
faz create "Address validation" --type task --priority 1 --parent faz-ab12 --description "Client and server checks"
faz dep add faz-ab12.0 faz-ab12
faz list --status open
faz children faz-ab12
faz ready
faz show faz-ab12.0
faz claim faz-ab12.0
faz close faz-ab12.0
faz reopen faz-ab12.0
faz info
```

## Storage model

`faz init` creates:

- `.faz/faz.db`
- `.gitignore` (if missing) and ensures `.faz/` is listed

Main schema:

- `issues`: lifecycle and hierarchy (`parent_id`)
- `dependencies`: issue graph (`issue_id` depends on `depends_on_id`)

## Notes

- `ready` lists unblocked open non-epic issues that are not actively claimed.
- Root IDs use `<project>-xxxx` and child IDs use `<parent>.<n>`.
- Valid types: `epic`, `task`, `bug`, `feature`, `chore`, `decision`.
- Valid statuses: `open`, `in_progress`, `closed`.
- Status symbols in list outputs: `○` open, `◐` in_progress, `✓` closed.
