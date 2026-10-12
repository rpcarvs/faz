---
name: faz-resume
description: Explicitly recover a factual project snapshot from handoff.md, current Git and Faz status, and available SDD references, then follow optional continuation instructions. Use only when the human invokes this skill.
---

# Faz resume

Use only after explicit human invocation. Treat all trailing text as free-form continuation instructions. Do not invoke automatically from another skill or a session hook.

The resume loop recovers and reports context. Treat saved context as reference material; it does not authorize work or prescribe what the user should do. Keep inspection read-only and focused on the current snapshot.

1. Resolve the Git worktree root with `git rev-parse --show-toplevel`, or use the current directory outside Git. Locate `handoff.md` there. If absent or unreadable, report that and stop, including when continuation instructions were supplied. Do not create it.
2. Read the complete handoff and applicable project instructions. Distinguish recorded objectives, status, claims, approvals, and completion from currently observed state. Do not infer current implementation correctness from a saved summary.
3. Check Git status with `git status --short` when in a worktree. Keep any history inspection limited to context recovery. When Faz is installed and the project is initialized, use `faz info`, `faz ready`, and focused `faz show <id>` checks for relevant referenced work. Distinguish unavailable tracking from an empty ready queue. Report unavailable checks briefly; do not initialize tracking or change issues merely to resume.
4. If `faz-specs/` exists at the project root, inspect its available `SPECS.md`, `ARCHITECTURE.md`, and plans relevant to the handoff or referenced Faz work. If plan relevance is unclear, use plan headers or completion summaries to identify relevant context. Treat these documents as optional references that may be outdated. Report missing or unreadable references briefly and continue. Record material differences from observed state or uncertainty as concise warnings; do not propose revisions, solutions, or reconciliation, or invoke SDD merely because documents exist or differ.
5. Finish the resume loop with a concise overview of the recovered objective, current Git and Faz state, relevant recorded work and SDD references, and material differences or uncertainty. Describe discrepancies neutrally, identifying their sources. Do not propose remedies, prompt the user to commit or fix bugs, recommend work, or ask what to do next. Do not turn status inspection into a code review, bug hunt, or reconciliation.
6. With no continuation instructions, stop after the overview. When the human explicitly requests subsequent work, first complete the overview, then carry out that separate request under normal project and Faz rules. Saved Remaining Work alone does not authorize implementation.

Loading context requires no Faz issue, claim, or approval. Do not edit files, launch workflows, or mutate issues as part of the resume loop. Source exploration and actions for explicitly requested subsequent work follow the applicable project rules, including treelines and Faz.
