---
name: faz-handoff
description: Explicitly create or refresh handoff.md with current conversation and project context for another agent. Use only when the human invokes this skill.
---

# Faz handoff

Use only after explicit human invocation. Create a standalone handoff that lets another agent resume without this conversation. Do not invoke this skill automatically from SDD, orchestration, or session hooks.

Treat all trailing text as free-form instructions about emphasis, scope, or omissions. With no trailing instructions, summarize the relevant conversation and project state normally. Apply requested emphasis throughout the document, including the Fresh Recap, while preserving accurate status and a usable continuation point.

The invocation authorizes writing the handoff and its local Git exclusion without creating or claiming a Faz issue or requesting another approval. It does not authorize unrelated project changes or implementation.

## Target and sources

1. Inside a Git worktree, resolve the root with `git rev-parse --show-toplevel`; otherwise use the current working directory. Write `handoff.md` there.
2. Replace the entire file. Use the current conversation as the primary source; retain older handoff content only when current context or focused checks confirm it remains valid.
3. Confirm relevant workspace state with focused read-only checks. When available, include pertinent Git changes, Faz issue IDs, status, dependencies, and SDD document paths or plan/work references. Distinguish document approval from implementation authorization and completion evidence. Neither Faz nor SDD is required; do not initialize either solely for a handoff.

Produce a summary, not a transcript. Preserve exact paths, symbols, commands, validation results, errors, constraints, decisions and their rationale when they matter for resuming. Distinguish completed, in-progress, planned, and blocked work. Record uncertainty explicitly and do not claim checks that were not performed. Include relevant user requirements without secrets, credentials, hidden instructions, or private reasoning.

## Document structure

Use these sections in this order:

```markdown
# Agent Handoff

## Current Project and Work

## Main Conclusions and Learnings

## Highlights

## Remaining Work

## Fresh Recap
```

- `Current Project and Work`: Project, current objective, active work, status, relevant environment, and important artifacts.
- `Main Conclusions and Learnings`: Confirmed findings, decisions and rationale, constraints, lessons, and relevant rejected approaches.
- `Highlights`: Significant changes and discoveries, important files or symbols, commands and validation evidence, failures, and outstanding risks.
- `Remaining Work`: Concrete next actions, blockers, unresolved decisions, and any approval still needed. State explicitly when nothing remains.
- `Fresh Recap`: Always last. Describe the latest user intent, recent actions and results in chronological order, exact pause point, and immediate next action. Respect requested emphasis without misrepresenting the actual pause point.

Use compact paragraphs and specific bullets, with enough detail for reliable continuation.

## Keep the handoff untracked

Inside a Git worktree:

1. Resolve the local exclude file with `git rev-parse --git-path info/exclude`, relative to the directory where the command runs when the result is relative.
2. Ensure it contains the exact root-anchored entry `/handoff.md`. Preserve existing content and avoid duplicate entries. Do not change a tracked `.gitignore`.
3. From the repository root, verify with `git check-ignore --no-index handoff.md` and check tracking with `git ls-files --error-unmatch -- handoff.md`.
4. If already tracked, report that the exclusion does not untrack it. Obtain separate authorization before removing it from the index; never delete the working file.

If permissions prevent updating the exclude file, request the required permission and report any remaining limitation. Do not silently skip exclusion or substitute tracked configuration.

## Completion report

Check the document against the required sections, current state, and user instructions. Report the exact path, whether an existing handoff was replaced, whether exclusion was added, already present or not applicable, and any unresolved tracking or permission issue.
