---
name: faz-task-management
description: Track implementation and its execution plan with Faz, and automatically file identified bugs. Read-only questions, research, exploration, reviews, and diagnosis do not require investigation tasks.
---

# Task Management with faz

## Core Rule

Use Faz for implementation: creating, modifying, or deleting project deliverables such as code, tests, configuration, or documentation.

Read-only questions, research, file inspection, codebase exploration, reviews, and diagnosis require no investigation task, claim, or Faz approval. Preparatory reading needs no separate task. Incidental caches, indexes, and temporary diagnostic output do not turn read-only assistance into implementation. Identified bugs must still be filed as described below.

When preparing implementation, inspect first, then create or reuse tasks for the planned changes. Obtain approval before implementation.

- Run `faz claim <id>` successfully before changing project deliverables.
- Add meaningful and rich descriptions for the tasks so other agents will know precisely what needs to be done by just reading the task.

## Required Lifecycle

Follow this lifecycle for implementation requests, not read-only assistance or bug filing alone.

1. Orient
- (Occasionally) Run `faz onboard`. This does not need to be execute all the time but occasionally.
- Run `faz info`.
- Run `faz ready`.

2. Plan
- Reuse relevant open tasks when possible.
- Create missing tasks. Group related tasks into epics by distinct domain or deliverable, not under a project-wide umbrella. Use one epic for a cohesive deliverable, or none for small standalone work.
- When using an epic, create it before its children so its ID is available.
- Create tasks one command at a time.

3. Execute
- Claim exactly one non-epic item before implementation: task, bug, feature, chore, or decision.
- If `faz claim` fails because already claimed, do not work on it. Pick another ready task.
- Use `faz update` only when task requirements change or need to be fixed.
- Add new requirements as tasks under the relevant epic, if any, before coding that new scope.

4. Close
- Run `faz close <id>` for completed non-epic work.
- When an epic is used, confirm its progress with `faz show <epic-id>`.
- Close epics after all children are done. DO NOT end the sessions leaving open epics with no children.

## Mandatory Task Rules

- Tasks MUST be atomic units! Do not aggregate a lot of work into a task. Remember, atomic tasks!
- Add rich descriptions to the atomic tasks so ANY other Coding Agent or Human will understand know what to do!
- Create the tasks and any necessary epics for the planned scope before implementation.
- Use `faz dep` for every real prerequisite, never merely to encode a preferred order. State what each blocker provides. Decompose work into independently implementable tasks where possible; genuine dependency chains remain valid.
- Identify intended file ownership in task descriptions. Tasks scheduled concurrently must not edit the same files or conflict through shared mutable resources. Split shared work into a prerequisite where appropriate, or serialize conflicting edits.
- After preparing the tasks, share the planned changes and wait for approval before implementation.

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

When a bug is identified, including during a read-only review or investigation of a user-reported bug, automatically create or reuse a Faz bug issue with the findings and reproduction details when available. Report the issue to the user. Filing requires no claim or approval and does not authorize fixing it; obtain approval and claim the issue before implementation.

During tracked implementation, also record discovered maintenance work or other inconsistencies as appropriate issues before fixing them. Do not silently expand the approved scope.

## Non-Negotiable Constraints

- Never claim epics.
- Another active skill does not by itself require the Faz lifecycle; apply the scope above.
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

## Execution Pattern

Use this pattern for implementation requests.

1. State a short draft plan in chat.
2. Run orientation commands.
3. Reuse or create tasks, grouping into epics where appropriate.
4. Share selected task IDs and wait for approval before implementation.
5. Claim a non-epic task.
6. Execute the task.
7. Close finished tasks and report final statuses.

## Subagent Requirement

For implementation subagents, include this skill text and assign each a specific Faz task ID. Read-only subagents need no task or claim; ensure their identified bugs are filed. Subagents do not inherit local skills automatically.

## Failure Protocol

- Missing Faz or a repository must not block read-only assistance. Complete the request and report any inability to file identified bugs; do not initialize a repository solely for tracking.
- If a faz command fails, correct syntax and retry.
- Use `faz <command> --help` only after a failure.
- Use `faz recap` for quick command reminders.
