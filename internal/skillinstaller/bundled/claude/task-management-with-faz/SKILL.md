---
name: task-management-with-faz
description: Use when starting work, research, analysis, planning tasks, creating features, fixing bugs, or any session that requires you to act
---

# Task Management with faz

## Overview

**No work without a tracked task. No task without faz.**

`faz` is the local CLI task tracker. Every session uses it. Other skills run their workflows, but faz runs in parallel with all of them. Skills tell you HOW to work. faz tracks WHAT you're working on. Both happen together, always.

**Violating the letter of this rule is violating the spirit of this rule.**

## The Iron Law

```
NO WORK WITHOUT `faz claim` FIRST
```

- Not one line. Not a "quick fix." Not "I'll register it after." Claim first, work second.
- Add meaningful and rich descriptions for the tasks so other agents will know precisely what needs to be done by just reading the task.

## Lifecycle: Orient -> Plan -> Execute -> Close

| Phase | Gate |
|-------|------|
| Orient | `faz onboard`, then `faz info` and `faz ready` |
| Plan | Epic + child tasks created in faz |
| Execute | `faz claim <id>` succeeds |
| Close | `faz close <id>` succeeds |

If `faz claim` fails (already claimed), pick another ready task.

## Mandatory Task Rules

- Tasks MUST be atomic units! Do not aggregate a lot of work into a task. Remember, atomic tasks!
- Add rich descriptions to the atomic tasks so ANY other Coding Agent or Human will understand know what to do!
- You first must created ALL the necessary epics and children before start working.
- Set blockers ALWAYS! You must be explicit when a task is being blocked by another task. That is non-negotiable. Use `faz dep` to manage blockers.
- After ALL this is done, communicate to the user and wait for approval.

Example Case: User asks you to develop a tool.
 - Given the request, you may create 1 epic with topic A. It requires 3 atomic tasks.
 - Do not start coding yet. The plan may need another epic with topic B with 5 atomic tasks.
 - After all tasks are created you now set the blockers / dependencies with `faz dep`.
 - Now you ask user to approve the setup.

## Quick Reference

Run `faz recap` for a full command overview.

| Action | Command |
|--------|---------|
| Get oriented | `faz onboard`, `faz info`, `faz ready` |
| Create epic | `faz create "Title" --type epic --priority 1 --description "..."` |
| Create task | `faz create "Title" --type task --priority 1 --parent <epic-id> --description "..."` |
| See task details | `faz show <id>` |
| List children | `faz children <epic-id>` |
| Set A is blocked by B | `faz dep add <A> <B>` |
| Remove the blocker | `faz dep remove <A> <B>` |
| List task blockers | `faz dep list <A>` |
| List what the task blocks | `faz dep list <B>` |
| Claim before coding | `faz claim <id>` |
| Mark done | `faz close <id>` |
| Scope changed | `faz update <id> --description "..."` |
| New requirement | Create new task under epic first, then code |

## Execution Pattern

Use this pattern whenever the user asks for work.

1. State a short draft plan in chat.
2. Run orientation commands.
3. Reuse or create epic plus child tasks.
4. Share selected task IDs.
5. Claim a non-epic task.
6. Execute the task.
7. Close finished tasks and report final statuses.

## Integration with Other Skills

faz is not an alternative to other skills. It layers on top:

| Skill | What it does | faz's role |
|-------|-------------|------------|
| brainstorming | Explores requirements | faz tracks the resulting work items |
| writing-plans | Creates implementation plan | Plan steps become faz tasks |
| executing-plans | Runs plan with subagents | Each subagent claims its faz task |
| TDD | RED-GREEN-REFACTOR cycle | faz claim before RED, faz close after GREEN |

**The failure mode:** Another skill's workflow feels complete on its own, so you skip faz. This is wrong. The other skill handles process. faz handles tracking. Both are required.

## Common Rationalizations - STOP

If you catch yourself thinking any of these, stop and run faz commands before continuing.

| Excuse | Reality |
|--------|---------|
| "I'll register in faz after coding" | After = never. Claim first, code second. |
| "Too small to track" / "Just prototyping" | If it changes code, it gets a task. No size exceptions. |
| "Another skill already tracks this" | Skills handle process. faz handles state. Both required. |
| "I'm mid-workflow, faz will break my flow" | faz runs in parallel with every skill. Not instead of. |
| "faz is overhead" | Skipping tasks breaks your user's trust. |

## Subagent Enforcement

When spawning subagents via the Agent tool, always include the full content of this skill in the subagent's prompt. Subagents do not inherit skills automatically. If a subagent will write or modify code, it must receive faz instructions and the task ID it should work on. No exceptions.

## Only Claim Non-Epic Items

You can only `faz claim` items of type task, bug, feature, chore, or decision. Never claim an epic directly. Epics are containers. Claim their children.

## Close Epics after Done

- Close epics after all children are closed and the epic is done.
- DO NOT end the sessions leaving open epics with no children.

