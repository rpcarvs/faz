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
# Install directly from GitHub:
go install github.com/rpcarvs/faz@latest

# Or clone the repo and install locally:
go install .
# If `faz` is not found, add GOPATH/bin to PATH (Bash example):
grep -q '$(go env GOPATH)/bin' ~/.bashrc || echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
faz init
faz onboard
```

## Install skills

`faz` ships built-in installer commands for the `task-management-with-faz` skill:

```bash
faz install codex-skill
faz install claude-skill
```

Behavior:

- Installs the skill into the target agent skill directory.
- Prints the final installed path after success.
- Fails if the destination already exists, unless `--force` is passed.

## Install context

`faz` also ships context installers for Codex and Claude global context files:

```bash
faz install codex-context
faz install claude-context
```

Behavior:

- Ensures the target context file exists.
- Manages the block between `<!-- FAZ-TASK-MANAGEMENT:BEGIN -->` and `<!-- FAZ-TASK-MANAGEMENT:END -->`.
- Upserts managed content: appends the block when missing, replaces it when present.
- By default writes to global files (`~/.codex/AGENTS.md` and `~/.claude/CLAUDE.md`).
- `--local` writes context files into the current project directory.

## Storage model

`faz init` creates:

- `.faz/taskstore.db`
- `.gitignore` (if missing) and ensures `.faz/` is listed

Main schema:

- `issues`: lifecycle and hierarchy (`parent_id`)
- `dependencies`: issue graph (`issue_id` depends on `depends_on_id`)

## Core commands

```bash
faz recap
faz install codex-skill
faz install claude-skill
faz install codex-context
faz install claude-context
faz create "Checkout revamp" --type epic --priority 1 --description "Improve checkout"
faz create "Address validation" --type task --priority 1 --parent faz-ab12 --description "Client and server checks"
faz dep add faz-ab12.0 faz-ab12
faz list --status open
faz monitor -t 5
faz monitor --all
faz children faz-ab12
faz ready
faz show faz-ab12.0
faz claim faz-ab12.0
faz close faz-ab12.0
faz reopen faz-ab12.0
faz info
```

## Notes

- `ready` lists unblocked open non-epic issues that are not actively claimed.
- `in_progress` is lease-based and can only be set via `faz claim`.
- `faz claim` is for executable work items. Epics are not claimable.
- If a task is already claimed, `faz claim` returns a non-zero exit code.
- Root IDs use `<project>-xxxx` and child IDs use `<parent>.<n>`.
- Valid types: `epic`, `task`, `bug`, `feature`, `chore`, `decision`.
- Valid statuses: `open`, `in_progress`, `closed`.
- Status symbols in list outputs: `○` open, `◐` in_progress, `✓` closed.
