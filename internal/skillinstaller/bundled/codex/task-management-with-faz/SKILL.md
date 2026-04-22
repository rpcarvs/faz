---
name: task-management-with-faz
description: Enforce faz task tracking for all work. Use when starting sessions, planning or executing features, fixing bugs, refactoring, researching, analyzing, or any workflow where you need to work.
---

# Task Management with faz

## Core Rule

No work without a claimed faz task.

- Run `faz claim <id>` successfully before action.
- Add meaningful and rich descriptions for the tasks so other agents will know precisely what needs to be done by just reading the task.

## Required Lifecycle

Follow this lifecycle in order for every coding session.

1. Orient
- Run `faz onboard`.
- Run `faz info`.
- Run `faz ready`.

2. Plan
- Reuse relevant open tasks when possible.
- If none exist, create one epic and child task(s).
- Do not batch epic and children creating together. This will fail as you need the epic id to define the children. Always create the epic first > get the id > and then create children.
- Create tasks one command at a time.

3. Execute
- Claim exactly one non-epic item before coding: task, bug, feature, chore, or decision.
- If `faz claim` fails because already claimed, do not work on it. Pick another ready task.
- Keep task details current with `faz update` when scope changes.
- Add new requirements as new child tasks under the epic before coding that new scope.

4. Close
- Run `faz close <id>` for completed non-epic work.
- Confirm epic progress with `faz show <epic-id>`.
- Close epics after all children are done. DO NOT end the sessions leaving open epics with no children.

## Mandatory Task Rules

- Tasks MUST be atomic units! Do not aggregate a lot of work into a task. Remember, atomic tasks!
- Add rich descriptions to the atomic tasks so ANY other Coding Agent or Human will understand know what to do!
- You first must created ALL the necessary epics and children before start working.
- Set blockers ALWAYS! You must be explicitly when a task is being blocked by another task. That is non-negotiable. Use `faz dep` to manage blockers.
- After ALL this is done, communicate to the user and wait for approval.

Example Case: User asks you ti develop a tool.
 - Given the request, you see may create 1 epic with topic A. It requires 3 atomic tasks.
 - Do not start coding yet. The plan may need another epic with topic B with 5 atomic tasks.
 - After all tasks are created you now set the blockers / dependencies with `faz dep`.
 - Now you ask user to approve the setup.

### Work classification

It is critical to properly classify the type of work you are executing. `faz` support many types and you MUST use them:

- task: A concrete unit of implementation work.
- bug: A defect fix for incorrect behavior.
- feature: New user-facing capability or product behavior.
- chore: Maintenance work with no direct feature outcome.
- decision: A recorded architectural or product choice.

Examples:

```bash
faz create "Implement token cache" --type task --priority 2 --description "..."
faz create "Fix login race" --type bug --priority 1 --description "..."
faz create "Add export API" --type feature --priority 2 --description "..."
faz create "Refactor parser" --type chore --priority 2 --description "..."
faz create "Choose DB strategy" --type decision --priority 1 --description "..."
```

### Dynamic issue creations

While implementing code and other works you may find bugs, stale code, missing documentation or other inconsistencies. Do not immediately and silently fix them. You must create a `faz` issue first as a `bug` or `chore` and notify about it when you finish your job. **That is critical** so these fixes do not happen silently.

## Non-Negotiable Constraints

- Never claim epics.
- If another workflow skill is active, still run faz lifecycle in parallel.
- Close epics after all children are closed and the epic is done.

## Command Set

- Orientation: `faz onboard`, `faz info`, `faz ready`
- Create epic: `faz create "Title" --type epic --priority 1 --description "..."`
- Create task: `faz create "Title" --type task --priority 1 --parent <epic-id> --description "..."`
- Show issue: `faz show <id>`
- List children: `faz children <epic-id>`
- Set A is blocked by B: `faz dep add <A> <B>`
- Remove the blocker: `faz dep remove <A> <B>`
- List task blockers: `faz dep list <A>`
- List what the task blocks: `faz dep list <B>`
- Claim work: `faz claim <id>`
- Update scope: `faz update <id> --description "..."`
- Close work: `faz close <id>`
- Reopen if needed: `faz reopen <id>`
- Recap help: `faz recap`

## Execution Pattern for Codex

Use this pattern whenever the user asks for work.

1. State a short draft plan in chat.
2. Run orientation commands.
3. Reuse or create epic plus child tasks.
4. Share selected task IDs.
5. Claim a non-epic task.
6. Execute the task.
7. Close finished tasks and report final statuses.

## Subagent Requirement

If spawning subagents that will also work, include this skill text and assign each subagent a specific faz task ID. Subagents do not inherit local skills automatically.

## Failure Protocol

- If a faz command fails, correct syntax and retry.
- Use `faz <command> --help` only after a failure.
- Use `faz recap` for quick command reminders.

