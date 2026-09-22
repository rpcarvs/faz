---
name: faz-orchestration
description: Explicitly orchestrate subagents to implement existing Faz tasks, supervise verification, and close out approved work, with or without SDD documents. Use only when the human invokes this skill.
---

# Faz orchestration

Use only after explicit human invocation. The invocation authorizes execution of the selected existing Faz scope, not creation of an initial task graph or expansion of requirements. Treat trailing text as additional instructions and context. User instructions override skill defaults, subject to higher-priority restrictions; clarify scope-changing requests before execution.

This skill does not require SDD and is not part of ordinary Faz sessions. Use the normal Faz tracking rules throughout execution; do not invoke the SDD skill automatically.

## Establish scope and readiness

1. Read Faz work, descriptions, dependencies, and claims using `faz info`, `faz ready`, `faz list`, and `faz show <id>`. Do not rely on previous chat context. Apply any task, epic, or plan selection supplied by the user.
2. If there are no unfinished executable issues, stop and ask the user to prepare the epics, tasks, and dependencies before invoking this skill. Do not create an initial graph. Distinguish an empty or completed scope from work that exists but is blocked or claimed; an empty ready list alone does not mean no tasks exist. Report a missing or inaccessible Faz store instead of treating it as an empty queue.
3. Inspect dependencies and potential editing conflicts. Proceed when scope is clear. If independent workstreams make the intended scope ambiguous, confirm with the user even if those workstreams do not collide. Work without specs is equally valid. Do not expand into unrelated ready work or unapproved newly discovered issues.
4. Read relevant project instructions and, for linked SDD work, the active plan and needed specifications. Stop affected work if requirements conflict or scope lacks approval. An unresolved `decision` issue requires a human decision, not an implementation worker.
5. Confirm subagent execution is available. If unavailable, report the limitation and stop rather than silently implementing alone.

## Delegate and supervise

Use up to 10 concurrent subagents unless the user specifies another limit, always bounded by the environment's capacity. This is a ceiling, not a target. Schedule only independent ready work; serialize overlapping edits or resolve ownership before dispatch. Workers must not spawn additional agents without coordinator authorization and capacity allocation.

Give each worker:

- One exact non-epic Faz issue ID, its intended outcome, dependencies, and approved scope.
- Relevant project instructions and the Faz tracking rules; do not assume workers inherit skills or conversation context.
- Editing ownership, known concurrent work, relevant context paths, and verification expectations.
- Instructions to read and successfully `faz claim <id>` before coding. If the claim fails, report it and do not work on that issue or take over another claim.
- The decision and defect rules below, and a prohibition on editing any `faz-specs/` files. Workers may read those files when needed.
- A requirement to report changes, verification results, remaining risks, and any blocking issues. Report readiness for review; leave closure to the orchestrator after verification.

Keep workers within their assigned boundaries. Do not overwrite another worker's or the user's changes. Recheck readiness as dependencies complete and maintain claim ownership while work or review remains active. Do not schedule an issue twice or treat an expired claim as proof its previous worker has stopped.

## Defects, conflicts, and decisions

Routine implementation choices within approved requirements are allowed. Neither coordinator nor workers may autonomously change requirements, architecture, constraints, or agreed implementation direction.

A worker may correct a defect introduced by its own assigned implementation only when the fix remains in scope, does not interfere with another agent's work, and presents no identified risk elsewhere. Record discovered bugs in Faz. If uncertain, ask the orchestrator before acting. The orchestrator may clarify approved scope, but decisions requiring new authority go to the user.

Record unrelated bugs, conflicts, and unresolved decisions in Faz with enough context for the user to act; use a `bug` or `decision` issue as appropriate and link blockers explicitly. If clarification is unavailable, document the issue rather than guessing. Stop affected work, continue only safe independent work, and communicate the need to the user. Ask immediately when the decision blocks further progress; otherwise include it in the final report. Reporting an issue does not authorize its implementation.

## Review and close out

1. Inspect each worker's changes and verification evidence against its task. Run relevant integration checks; a worker's completion message alone is insufficient. Return in-scope defects for correction under the same boundaries.
2. Close verified Faz issues, then completed container epics. Recheck dependencies and dispatch newly ready work within the selected scope. Stop when that scope is complete or remaining work requires human input or external action; report blockers accurately.
3. For SDD-linked work, only the orchestrator may update the active plan's implementation status and dated completion evidence. Follow its close-out instructions: human validation gates completion when criteria require it or the orchestrator requests it. Record what to check and why, wait for confirmation, and return to in-progress correction work if validation fails. Never equate task closure with whole-plan completion.
4. Do not change approved requirements, architecture, or completed plans. Issues found after a plan is `Implemented` belong to new explicitly requested SDD planning or ordinary Faz work outside that plan. Partial execution of a plan cannot mark the whole plan implemented.
5. Report completed scope, verification performed, pending human validation, and outstanding bugs or decisions. Do not claim success when verification is incomplete or required confirmation remains pending.
