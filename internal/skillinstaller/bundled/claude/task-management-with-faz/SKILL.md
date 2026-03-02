---
name: task-management-with-faz
description: Use when starting implementation work, planning tasks, creating features, fixing bugs, or any coding session that produces or modifies code
---

# Task Management with faz

## Overview

**No code without a tracked task. No task without faz.**

`faz` is the local CLI task tracker. Every coding session uses it. Other skills run their workflows, but faz runs in parallel with all of them. Skills tell you HOW to work. faz tracks WHAT you're working on. Both happen together, always.

**Violating the letter of this rule is violating the spirit of this rule.**

## The Iron Law

```
NO CODE WITHOUT `faz claim` FIRST
```

- Not one line. Not a "quick fix." Not "I'll register it after." Claim first, code second.
- Add meaningful and rich descriptions for the tasks so other agents will know what needs to be done by just reading the task.
- Prefer to create many tasks in smaller actionable units instead of aggregating many steps in only one task.

## Lifecycle: Orient -> Plan -> Execute -> Close

| Phase | Gate |
|-------|------|
| Orient | `faz onboard`, then `faz info` and `faz ready` |
| Plan | Epic + child tasks created in faz |
| Execute | `faz claim <id>` succeeds |
| Close | `faz close <id>` succeeds |

If `faz claim` fails (already claimed), pick another ready task.

## Quick Reference

Run `faz recap` for a full command overview.

| Action | Command |
|--------|---------|
| Get oriented | `faz onboard`, `faz info`, `faz ready` |
| Create epic | `faz create "Title" --type epic --priority 1 --description "..."` |
| Create task | `faz create "Title" --type task --priority 1 --parent <epic-id> --description "..."` |
| See task details | `faz show <id>` |
| List children | `faz children <epic-id>` |
| Claim before coding | `faz claim <id>` |
| Mark done | `faz close <id>` |
| Scope changed | `faz update <id> --description "..."` |
| New requirement | Create new task under epic first, then code |

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

