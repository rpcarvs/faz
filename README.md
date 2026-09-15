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

- Installs the `faz-task-management` and `faz-spec-driven` skills.
- Adds or updates the managed FAZ task-management context block.
- Installs a SessionStart hook that runs `faz init && faz onboard` inside Git repositories.
- Keeps spec-driven development explicitly invoked. The SessionStart hook runs only normal Faz initialization and onboarding.
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

## Spec-driven development

`faz-spec-driven` is an optional, explicitly invoked skill for turning a human development brief into tracked specifications, architecture, and an outcome-based plan. It is not part of the normal Faz task loop.

Invoke it with the complete brief, including goals, constraints, architecture ideas, stack preferences, and desired behavior.

For Codex, the installed skill metadata disables implicit invocation. For Claude, the installed skill disables model invocation. In both providers, explicit invocation passes the complete trailing text as the development brief.

Codex invocation:

```text
$faz-spec-driven Add CSV import with validation, audit logging, and a PostgreSQL-backed importer. Keep the existing CLI compatible.
```

Claude invocation:

```text
/faz-spec-driven Add CSV import with validation, audit logging, and a PostgreSQL-backed importer. Keep the existing CLI compatible.
```

The explicit-invocation metadata follows the [OpenAI skill documentation](https://learn.chatgpt.com/docs/build-skills) and [Claude Code skill documentation](https://code.claude.com/docs/en/skills).

The skill writes tracked documents under `faz-specs/`:

```text
faz-specs/
  SPECS.md
  ARCHITECTURE.md
  PLAN01.md
  PLAN02.md
```

`SPECS.md` and `ARCHITECTURE.md` describe the project broadly. Each numbered plan describes the requested work at a high level. A plan's Work outcomes section contains a summary table and detailed `Wxx` subsections with expected results, dependencies, requirement references, and verification references. Outcomes are not fixed Faz tasks: agents choose the issue types, hierarchy, count, dependencies, and any newly discovered bug work.

The skill drafts or revises the documents, asks focused questions when consequential information is missing, and presents the revision for approval. After approval it creates or reconciles appropriate Faz work, reports readiness, and stops. A separate human instruction is required before implementation begins.

Invoke the skill with no arguments to reconcile approved documents against implementation, tests, and related Faz state:

```text
$faz-spec-driven
```

```text
/faz-spec-driven
```

Reconciliation reports behavior outside the approved scope as a discrepancy and proposes changes for approval. It does not silently redefine specifications. Completion summaries record dated acceptance evidence, not task counts or closed-task status. `.faz/` remains local and ignored; SDD documents in `faz-specs/` are intended to be tracked.

### Optional plan and work associations

Issues can optionally link to a plan and outcome. This does not change ordinary Faz workflows, and `plan_id` and `work_id` are not unique identifiers. Multiple issues of any type may share either or both values; Faz issue IDs remain the item identities.

Existing initialized databases are upgraded automatically when opened by the new binary. No extra `faz init` is required; uninitialized projects still require initialization.

```bash
faz create "Build CSV parser" --type task --plan PLAN01 --work W01 --description "Parse and validate imported rows"
faz create "Explore shared outcome" --type task --work W1 --description "Work association can be used without a plan"
faz update faz-ab12 --plan PLAN01 --work W01
faz update faz-ab12 --clear-plan
faz update faz-ab12 --clear-work
faz list --plan PLAN01
faz list --work W01
faz list --all --plan PLAN01 --work W01
```

Work filtering is exact: `W1` and `W01` are different values. A work-only query may return items from multiple plans; combine `--plan` and `--work` when selecting a particular outcome. `--clear-plan` and `--clear-work` deliberately remove only the selected association and do not change descendants.

When creating a child, omitted plan/work values inherit from its parent at creation time. Explicit child values override the corresponding parent value. Later parent edits or reparenting do not rewrite existing child associations.

For explicit `--plan` assignment, Faz resolves `PLAN01` to the regular file `faz-specs/PLAN01.md` under the Git root. Path escapes, plan symlinks that escape `faz-specs/`, a specs directory resolving outside the Git root, and non-regular targets are rejected. Once an association is stored, a missing plan file does not prevent normal reads, claims, closures, or other issue lifecycle operations. `faz show` and successful `faz claim` print `SDD Plan` and `SDD Work` only when the corresponding association is present.

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
faz create "Linked work" --type task --plan PLAN01 --work W01 --description "Optional SDD association"
faz list --all --plan PLAN01 --work W01
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
