---
name: faz-resume
description: Explicitly load handoff.md and current Git and Faz status, then follow optional continuation instructions. Use only when the human invokes this skill.
---

# Faz resume

Use only after explicit human invocation. Treat all trailing text as free-form continuation instructions. Do not invoke automatically from another skill or a session hook.

1. Resolve the Git worktree root with `git rev-parse --show-toplevel`, or use the current directory outside Git. Locate `handoff.md` there. If absent or unreadable, report that and stop, including when continuation instructions were supplied. Do not create it.
2. Read the complete handoff and applicable project instructions. Treat the handoff as saved context, checking current state before relying on recorded status, claims, approvals, or completion.
3. Check Git status with `git status --short` when in a worktree. When Faz is installed and the project is initialized, use `faz info`, `faz ready`, and focused `faz show <id>` checks for relevant referenced work. Distinguish unavailable tracking from an empty ready queue. Report unavailable checks briefly; do not initialize tracking or change issues merely to resume.
4. Without continuation instructions, briefly report the recovered objective, current state, material differences from the handoff, and next action, then stop. With instructions, use the recovered context to carry out that request under the normal project and Faz rules. Saved Remaining Work alone does not authorize implementation.

Loading context requires no Faz issue, claim, or approval. Keep checks focused; source exploration for subsequent work follows the applicable treelines workflow. Do not rewrite the handoff, invoke SDD, or start orchestration unless the human requests it.
