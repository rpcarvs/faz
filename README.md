# faz

`faz` is a lightweight local task tracker for agent workflows.
It uses Go + Cobra for the CLI and SQLite for issue storage with graph-like task relationships.

## Why faz

- Personal local-first task tracking for daily coding workflows
- Minimal friction for humans and AI agents
- No GitHub integration or remote coupling
- Explicit lifecycle and dependency commands with practical output
- Fast project context recovery from the terminal

## Quick start

### Homebrew

```bash
brew tap rpcarvs/faz
brew install faz
```

### Go
```bash
# Install directly from GitHub:
go install github.com/rpcarvs/faz@latest

# Check the installed version:
faz -v

# Or clone the repo and install locally:
go install .
# If `faz` is not found, add GOPATH/bin to PATH (Bash example):
grep -q '$(go env GOPATH)/bin' ~/.bashrc || echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
faz init
faz onboard
```

## Install agent integration

`faz` can install its agent integration for Codex or Claude:

```bash
faz install codex
faz install claude
```

Behavior:

- Installs the `task-management-with-faz` skill.
- Adds or updates the managed FAZ task-management context block.
- Installs a SessionStart hook that runs `faz init && faz onboard` inside Git repositories.
- Prints all installed or updated paths.

Use `--local` to install into the current Git repository instead of the global agent config:

```bash
faz install codex --local
faz install claude --local
```

Local behavior:

- Resolves the Git repository root even when run from a subdirectory.
- Writes the managed context block to repo-root `AGENTS.md`.
- For Claude, writes repo-root `CLAUDE.md` as a pointer to `AGENTS.md`.
- Installs skills and hooks under repo-root `.codex/` or `.claude/`.

## Shell completion

`faz` exposes shell completion through Cobra/Fang:

```bash
faz completion bash
faz completion zsh
faz completion fish
```

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
faz install codex
faz install claude
faz install codex --local
faz install claude --local
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
faz -v
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
