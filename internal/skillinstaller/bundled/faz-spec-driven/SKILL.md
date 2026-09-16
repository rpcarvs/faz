---
name: faz-spec-driven
description: Explicitly develop or reconcile Faz-backed specifications, architecture, and outcome-based plans from a human development brief. Use only when the human invokes this skill.
---

# Faz spec-driven development

Use this skill only after an explicit human invocation. It is not part of the ordinary Faz task loop. Treat all text supplied with the invocation as one free-form development brief, not as positional arguments.

The skill maintains tracked SDD documents in `faz-specs/`:

```text
faz-specs/
  SPECS.md
  ARCHITECTURE.md
  PLAN01.md
  PLAN02.md
```

Plan identifiers are stable filename stems, such as `PLAN01`. Allocate a new identifier as the highest existing numeric suffix plus one, with at least two digits. Do not reuse holes, rename plans, or introduce a plan registry. A new brief normally creates a new plan; amend the relevant unfinished plan when the brief clearly revises approved work. Plans marked `Implemented` remain historical records; subsequent work needs a new explicitly invoked SDD plan or ordinary non-SDD Faz work.

## Select the mode

### Development brief

When the invocation contains a brief:

1. Inspect existing SDD documents and relevant implementation, tests, and Faz work. For a project without SDD documents, map the whole project at a high level and investigate the requested work in detail.
2. Ask focused questions through host question prompts when available, or ask them directly otherwise. Ask only about consequential gaps, such as unclear scope, constraints, compatibility, ownership, or acceptance evidence. Do not use questions to replace reasonable investigation.
3. Draft or revise the documents. Keep observed current behavior, approved intent, and proposed changes distinguishable. Preserve approved sections intact until review, and label proposed changes separately. Preserve unrelated approved content and earlier plan history.
4. Present the concrete document revision for human approval. Do not create implementation issues before that approval.

### Reconciliation

When the invocation has no brief:

1. If SDD documents do not exist, establish a high-level baseline proposal and ask the human to state the intended scope. Do not invent product scope or create implementation work.
2. Otherwise, read the approved documents, relevant code and tests, and existing Faz work, including closed work when it is associated with the plan.
3. Compare implementation evidence with approved requirements, architecture, outcomes, acceptance criteria, and recorded completion evidence.
4. If behavior exists outside approved scope, record it as a discrepancy and propose a document change for approval. Never silently rewrite requirements to make unapproved behavior intended.
5. Preserve dated completion history. Distinguish an intentional, approved supersession from a regression or an unmet criterion. Report discrepancies concerning an `Implemented` plan without reopening or editing it; propose follow-up work separately, without authorizing implementation. Do not change documents merely because Faz task counts or leases changed.

If there is no meaningful discrepancy, report that result without unnecessary document churn. A missing local `.faz/` database is not evidence that a plan was not implemented.

## Document contract

Use the templates in [assets](assets/) as the starting structure. Adapt their content to the project; do not copy placeholders into final documents.

- `SPECS.md` is project-wide. Assign stable requirement references such as `R01` and state acceptance criteria and validation evidence.
- `ARCHITECTURE.md` is project-wide. Describe current and proposed boundaries, data flow, integrations, constraints, and decisions relevant to the work.
- Each `PLANnn.md` is a high-level umbrella for approved work. It links requirements to outcomes and validation. It is not a task manifest.

Each plan must contain Code boundaries and Acceptance and verification sections. Its Work outcomes section has both:

1. A compact summary table with `Outcome`, `Depends on`, and `Requirements` columns.
2. One `### Wxx ...` subsection per outcome with a description, an `Expected result`, and explicit `Depends on`, `Requirements`, and `Verification` references.

Every generated plan must retain concise, actionable guidance directly under `## Work outcomes` that Wxx entries are outcomes rather than Faz task slots and that implementation agents choose the issue types, hierarchy, count, and dependencies. Keep the table and detailed subsections consistent. Wxx references name product outcomes, not implementation task slots. Do not put Faz IDs, prescribed issue types or counts, required epics, or a one-to-one task mapping in a plan. Agents may choose and revise the Faz breakdown, and may create bugs or other work discovered during implementation.

Every generated plan must also retain a `## Completion summary` near the top, including the template's actionable close-out instructions. These must remain in the generated document, not only in this skill: header transitions, dated outcome evidence, human-validation gates, the correction loop, and the boundary after `Implemented`. Implementation agents follow them without invoking this skill or altering approved requirements.

## Approval, Faz preparation, and handoff

Document drafting under this explicitly invoked skill is allowed before an implementation task graph exists. Do not claim or start implementation while preparing the SDD documents.

After the human approves the document revision:

1. Persist the approval date, approved revision, and exactly reviewed document changes in the relevant plan's Review history, and update its `Approval:` header to identify the approved revision. Promote only the reviewed proposals to approved content. Preserve unrelated content, unapproved proposals, and prior history intact.
2. Read the actual `PLANnn` and Wxx identifiers. Do not normalize identifiers: `W1` and `W01` are distinct.
3. Before any Faz issue mutation, verify that the installed CLI supports the plan/work association and coverage operations needed for the handoff, including create or update associations and `faz list --all --plan PLAN01 --work W01`. If support is missing, report an upgrade/readiness blocker and stop without creating unlinked work.
4. Inspect existing Faz coverage, including descriptions, dependencies, claims, status, and closed work. Shared plan/work associations are legitimate and do not prove duplicate work.
5. Choose an appropriate issue breakdown. Create or update only the epics, tasks, features, bugs, chores, and decisions that are needed, with rich descriptions and dependencies. Preserve existing claims and avoid blindly replaying interrupted creation. Associate related work with `--plan PLANnn` and `--work Wxx`. Many items may share either value. Work-only filtering can span plans, so combine both values for one plan outcome.
6. Confirm that the planned graph covers the approved outcomes and that its currently ready work is sufficient to begin. Not every outcome must be ready when dependencies are correctly represented. Report readiness and stop.

The approval covers the exactly reviewed document scope, so do not seek an extra task-graph approval for that scope. It authorizes preparation only. A separate human instruction is required before claiming or coding implementation work. Task restructuring, splitting work, or recording a newly discovered bug does not itself require revising a task list in the plan. A material product-scope change does require renewed document approval.

## During normal implementation

Normal Faz work remains independent of this skill. After a separate human implementation instruction, an implementation agent claims its linked Faz item, reads the linked plan, sets `Implementation: In progress`, and implements normally without invoking this skill again. `Not started` applies until implementation begins. Record dated acceptance evidence by outcome in the plan's Completion summary, including partial verification and remaining checks. Do not equate closed tasks with verified completion, mirror live task counts, or have `faz close` write Markdown.

## Post-implementation finalization

1. Verify every applicable acceptance criterion and record dated evidence. If verification fails or work remains, keep `Implementation: In progress`; do not claim whole-plan completion.
2. Require human validation when an acceptance criterion calls for it or the implementation agent determines human testing or confirmation is needed. Record what the human must check and why, set `Implementation: Awaiting human validation` when ready for that review, and ask the human. Once requested, this validation gates completion.
3. If the human reports a failure before completion, record it, return to `Implementation: In progress`, and create or reopen Faz work linked to the same plan and outcome. Correct the issue, repeat relevant verification, and request human revalidation. Corrections to approved behavior need no scope revision; changed requirements need an approved revision to the unfinished plan. Preserve prior evidence and the correction history.
4. When all applicable criteria pass and any requested human validation is confirmed, record that confirmation and a dated whole-plan completion summary, then set `Implementation: Implemented` before reporting completion. Clear contextual confirmation such as "it works" is sufficient for the requested checks, not for unrelated unverified criteria. If no human validation is required or requested, finalize after agent verification without adding a confirmation gate.
5. Keep `Approval:` and `Revision:` tied to document-scope approval, not implementation acceptance; completion evidence alone does not change them. After `Implemented`, do not reopen or edit the completed plan for new issues. Handle those through a new explicitly invoked SDD plan or ordinary Faz work outside that plan.
