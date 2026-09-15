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

Plan identifiers are stable filename stems, such as `PLAN01`. Allocate a new identifier as the highest existing numeric suffix plus one, with at least two digits. Do not reuse holes, rename plans, or introduce a plan registry. A new brief normally creates a new plan; amend the relevant plan when the brief clearly revises approved work.

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
5. Preserve dated completion history. Distinguish an intentional, approved supersession from a regression or an unmet criterion. Do not change documents merely because Faz task counts or leases changed.

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

Every generated plan must also retain a `## Completion summary` near the top. It must direct normal implementation agents to record dated acceptance evidence there without invoking this skill or altering approved requirements. It must state that partial outcome evidence is recorded separately, whole-plan completion requires every applicable acceptance criterion, and later discrepancies preserve earlier evidence and distinguish supersession from regression.

## Approval, Faz preparation, and handoff

Document drafting under this explicitly invoked skill is allowed before an implementation task graph exists. Do not claim or start implementation while preparing the SDD documents.

After the human approves the document revision:

1. Persist the approval date, approved revision, and exactly reviewed document changes in the relevant plan's Review history. Promote only the reviewed proposals to approved content. Preserve unrelated content, unapproved proposals, and prior history intact.
2. Read the actual `PLANnn` and Wxx identifiers. Do not normalize identifiers: `W1` and `W01` are distinct.
3. Before any Faz issue mutation, verify that the installed CLI supports the plan/work association and coverage operations needed for the handoff, including create or update associations and `faz list --all --plan PLAN01 --work W01`. If support is missing, report an upgrade/readiness blocker and stop without creating unlinked work.
4. Inspect existing Faz coverage, including descriptions, dependencies, claims, status, and closed work. Shared plan/work associations are legitimate and do not prove duplicate work.
5. Choose an appropriate issue breakdown. Create or update only the epics, tasks, features, bugs, chores, and decisions that are needed, with rich descriptions and dependencies. Preserve existing claims and avoid blindly replaying interrupted creation. Associate related work with `--plan PLANnn` and `--work Wxx`. Many items may share either value. Work-only filtering can span plans, so combine both values for one plan outcome.
6. Confirm that the planned graph covers the approved outcomes and that its currently ready work is sufficient to begin. Not every outcome must be ready when dependencies are correctly represented. Report readiness and stop.

The approval covers the exactly reviewed document scope, so do not seek an extra task-graph approval for that scope. It authorizes preparation only. A separate human instruction is required before claiming or coding implementation work. Task restructuring, splitting work, or recording a newly discovered bug does not itself require revising a task list in the plan. A material product-scope change does require renewed document approval.

## During normal implementation

Normal Faz work remains independent of this skill. After a separate human implementation instruction, an implementation agent claims its linked Faz item, reads the linked plan, and implements normally without invoking this skill again. Record dated acceptance evidence in the plan's Completion summary after verifying an outcome. Mark the whole plan complete only when every applicable acceptance criterion is verified; record partially verified outcomes separately. Do not equate closed tasks with verified completion, mirror live task counts, or have `faz close` write Markdown. A later reconciliation can report new discrepancies while preserving the earlier completion record.
