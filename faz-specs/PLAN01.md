# PLAN01: Explicit spec-driven development in Faz

Approval: Approved for implementation by the user
Revision: 4
Implementation: Implemented; local automated and scenario verification passed

## Completion summary

Reviewed: 2026-09-15

W01-W07 are implemented. Local checks passed; interactive discovery and invocation inside Codex and Claude remain unverified. No global provider installation was changed. This record distinguishes implementation delivery from complete provider-host acceptance.

- AC01-AC07: storage, service, CLI, and migration regressions pass. Existing databases upgrade through ordinary commands, including concurrent opens and rollback/retry after a failed migration. The new binary upgraded this repository's pre-SDD database without changing any prior issue fields or losing dependencies, verified against a SQLite backup before associations were added.
- AC08: isolated local/global installer tests pass for both providers, including explicit-only metadata, complete asset delivery, reruns, force behavior, configuration preservation, and symlink protection. The installed Codex skill passed the skill validator. Actual provider discovery/invocation was not exercised.
- AC09-AC14: independent worker-driven scenarios exercised a development brief, discrepancy reconciliation despite closed issues, and interrupted multi-issue outcome coverage. A fresh generation retest retained the agreed Work and completion guidance after an observed omission was corrected. Approval bookkeeping and CLI capability checks now precede issue mutation. These are scenario evaluations, not provider-host tests.
- AC15: existing monitor, Kanban, claim, and output-stream regression tests pass. A subprocess smoke test verified linked/unlinked context, exact and combined filters, inheritance, clearing, missing-document lifecycle, optional associations, and unchanged onboarding behavior.

Executed successfully: `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint run ./...`, repeated concurrent-upgrade/rollback tests, skill validation, and the subprocess CLI/install smoke test. Existing Faz work was associated with this plan without recreating issues or prescribing a permanent task manifest.

Normal implementation agents may append dated acceptance evidence without invoking SDD or changing approved requirements. Whole-plan verification requires all applicable acceptance checks; preserve this evidence if later review identifies discrepancies or approved supersession.

## Purpose and approval boundary

Add one human-invoked `faz-spec-driven-dev` skill that converts a development brief into specifications, architecture, and an approved high-level plan. Faz remains the local task tracker used during implementation.

The product decisions below were agreed in discussion. This document proposes implementation outcomes and verification. Approval of this implementation plan authorizes building the feature. The feature itself has a separate handoff: its users approve generated documents, receive the corresponding Faz issues, and separately instruct an agent to implement them.

This bootstrap plan does not prescribe its local tracking issue IDs or decomposition. Existing local issues can receive native associations after that capability exists, without recreating them.

## Product requirements

- R01: SDD starts only through explicit human invocation. Session hooks and ordinary task-management context must not automatically invoke it.
- R02: Arguments are the full natural-language development brief, including goals, principles, behavior, constraints, stack preferences, and architecture ideas. Clarify consequential gaps and conflicts using host question prompts, or direct questions when unavailable.
- R03: With arguments, develop the requested change through the full SDD process. Existing documents provide context rather than reducing the request to reconciliation.
- R04: Without arguments, compare documents with implementation, tests, and Faz state. Report discrepancies and propose changes for approval. Never silently adopt out-of-scope implementation as approved requirements.
- R05: Keep project-wide SPECS.md and ARCHITECTURE.md in tracked faz-specs/, alongside numbered PLAN01.md, PLAN02.md, and subsequent plans. Initially map existing projects at a high level and examine the requested feature deeply.
- R06: Plans describe high-level Wxx outcomes related to Rxx requirements, scope, exclusions, outcome dependencies, code boundaries, acceptance criteria, and test/validation strategy. They do not prescribe issue IDs, types, counts, or one-to-one task mappings.
- R07: Present document changes and the plan for approval. After approval, create or reconcile appropriate Faz issues and dependencies, report readiness, and stop without claiming or implementing.
- R08: Agents choose and evolve the issue breakdown, including new tasks and discovered bugs. Repeated SDD invocations inspect existing coverage rather than replaying a fixed task list. Product scope changes require approval; ordinary task decomposition does not.
- R09: Keep .faz/ local and ignored. Optional nullable plan_id and work_id fields support association and filtering; neither field nor their combination is unique. Existing Faz issue IDs remain the identities.
- R10: New children inherit both associations from their parent when omitted. Later parent edits do not propagate; existing child associations change only explicitly.
- R11: Show and successful claim context display populated SDD Plan and SDD Work fields and omit absent keys. Normal non-SDD workflows receive no SDD-specific query instructions or extra loop. Kanban remains unchanged.
- R12: Implementation agents record dated, evidence-backed completion atop a plan after acceptance verification. Closed issues alone do not prove completion. Explicit reconciliation preserves earlier completion history and distinguishes intentional supersession from discrepancies.

## Document contract

The bundled skill supplies focused templates:

- SPECS.md: purpose, boundaries, numbered requirements, constraints, acceptance criteria, and unresolved questions.
- ARCHITECTURE.md: current components, relevant interfaces/data flows, proposed changes, and decisions with rationale.
- PLANnn.md: scope, approval/revision, high-level Wxx outcomes and Rxx links, outcome dependencies, code boundaries, verification, completion, and review history.

Begin each Work section with a note that agents determine the Faz task breakdown. Include a summary table of outcomes, dependencies, and requirement references, followed by a subsection for each outcome. Each subsection describes intent and expected result, then references dependencies, requirements, and acceptance/verification criteria. Keep descriptions proportional to complexity and keep summary references consistent with the subsections. Do not put local Faz IDs, a task manifest, prescribed issue types, or mandatory epic suggestions into the document.

Use the next available numeric plan number, padded to at least two digits. Never renumber, reuse, or overwrite an unrelated plan. Renaming support and a separate immutable plan-ID system are out of scope. A new request normally creates a new plan; an amendment revises the relevant existing plan. Identify overlapping outcomes before creating additional work.

Distinguish approved requirements, observed behavior, and proposed changes. Keep approved sections intact while labeling proposed additions/replacements with the draft plan. Approval covers reviewed document changes and the plan revision. Material changes invalidate affected approval; task restructuring and verification bookkeeping do not.

Completion records describe what was checked, when, the result, and remaining acceptance gaps. They do not mirror task lists, IDs, counts, types, or leases. The evidence for outcomes determines completion, not the chosen task topology.

When no documents and no brief exist, propose a high-level baseline and clarify missing intent rather than inventing development scope. A fresh clone with no local Faz history is an unknown tracking state, not proof that every plan is unfinished. Inspect documents and code before proposing new local tracking work.

## Data and inheritance

Add nullable TEXT `plan_id` and `work_id` fields to issues and optional fields to the Go model.

- plan_id stores the filename identifier, for example PLAN01 for faz-specs/PLAN01.md at the Git root. No plans table or SQL-generated plan identity is needed.
- work_id stores the outcome identifier read from the document, for example W01. The CLI does not enforce heading syntax or numeric padding.
- Neither field nor their pair has a unique constraint. Many epics, tasks, features, chores, decisions, or bugs may share them.
- Both fields are optional associations, not identifiers or requirements for ordinary issue creation.
- At creation, omitted child fields inherit the corresponding parent values. Explicit child values override those defaults.
- Later parent changes and reparenting do not silently rewrite stored associations. Child changes remain explicit.
- Existing issue IDs, statuses, claims, hierarchy, and dependencies survive migrations and ordinary lifecycle operations.

Keep persistence in the repository and field validation in focused service helpers. Use nonunique indexes for filtering. Preserve existing unique public issue IDs.

Work filters match the exact stored identifier: W1 and W01 are distinct. The SDD skill reads and copies the actual identifier when assigning associations or querying. Do not silently normalize aliases or impose an additional naming standard. A work-only filter may match across plans; combine --plan and --work to select a specific plan outcome.

Validate explicitly assigned plan targets relative to the Git root without arbitrary path traversal or escaping symlinks. Missing documents must not obstruct inherited associations or ordinary reading, claiming, closing, and other lifecycle operations on existing issues. Work assignment/filtering does not require a Markdown heading parser.

Do not reject, merge, or overwrite issues because their associations match. Duplicate work is a semantic question: inspect descriptions, dependencies, status, and outcome coverage during explicit SDD sessions. Shared references cannot answer that question.

## CLI surface

Proposed usage:

```text
faz create "Implement requirement" --type task --plan PLAN01 --work W01 --description "..."
faz update <id> --plan PLAN01 --work W01
faz list --plan PLAN01
faz list --plan PLAN01 --work W01
faz list --all --plan PLAN01 --work W01
faz show <id>
faz claim <id>
```

Association flags and filters are optional. Updates must support explicit clearing and reject conflicting set/clear requests. Clearing syntax is an implementation detail to document alongside the commands; it must not silently change descendants.

Show and successful claim responses add only the populated fields:

```text
  SDD Plan: PLAN01 (faz-specs/PLAN01.md)
  SDD Work: W01
```

Absent fields produce no key or line. Existing unlinked claim output remains unchanged. Filters select rows without adding columns. List, monitor, kanban, and create retain their current layouts. Associations do not alter readiness, claim leases, claim safety, or status transitions.

The normal hooks/onboarding/task loop do not instruct agents to run SDD filters. SDD query guidance belongs in the explicitly invoked skill. Linked claim/show context identifies the plan; the plan contains scoped completion instructions.

## Skill and installation

Maintain one shared workflow with provider-specific explicit invocation metadata:

- Codex: agents/openai.yaml sets policy.allow_implicit_invocation to false. Read the brief from the explicit invocation message.
- Claude: skill frontmatter sets disable-model-invocation to true. Consume the entire invocation argument text.

These settings are documented in [OpenAI's skill documentation](https://learn.chatgpt.com/docs/build-skills) and [Claude's skill documentation](https://code.claude.com/docs/en/skills).

Extend the existing embedded installer to deliver both task-management-with-faz and faz-spec-driven-dev, with required metadata, templates, and references. Keep shared instructions maintained once. Preserve unrelated context/hooks and existing install conventions. Verify local/global installation, reruns, and explicit discovery using isolated destinations.

Keep the SDD skill self-contained. Use available project tools and respect project instructions without requiring multiple separately installed planning skills. Clarification and approval stay in the human's conversation.

Where generic task instructions need an exception for explicit drafting or already-approved work, keep it narrowly scoped. Do not add SDD query guidance or an SDD loop to ordinary non-SDD context. SessionStart behavior remains task management.

## Approval, issue creation, and handoff

1. Inspect relevant documents, code, tests, and existing work.
2. Develop the brief or reconcile the project according to invocation mode.
3. Draft high-level outcomes with requirements, acceptance criteria, and validation coverage.
4. Request approval of the concrete document changes and revision. Do not create implementation issues for the draft before approval.
5. Read the plan/work identifiers and inspect associated issues, including closed ones. Assess coverage; do not assume one issue per outcome.
6. Choose an appropriate current issue breakdown. Create required epics/issues with rich descriptions and blockers; associate related work and preserve existing claims.
7. Check that the initial execution work addresses the approved outcomes. If issue creation is interrupted, inspect partial results before continuing rather than blindly replaying creates.
8. Record approval, report readiness, and stop. Keep the task graph in Faz. A separate implementation instruction is required before claiming work.

During implementation or review, agents can split work, choose different issue types, create additional tasks, and report bugs. These execution choices do not require revising a task list in the plan. Reporting unrelated bugs does not silently expand approved product scope.

Completion bookkeeping verifies outcomes and updates the plan record without invoking SDD or changing requirements. The CLI close command does not write Markdown.

## Work outcomes

Wxx entries describe outcomes related to product requirements, not task slots. Agents choose and evolve the Faz issue types, hierarchy, count, and decomposition. Several epics or issues can contribute to one WORK, and new tasks or bugs are allowed. No one-to-one mapping or prescribed epic structure is implied. Outcome dependencies express technical ordering; issue blockers belong in Faz.

| Outcome | Depends on | Requirements |
| --- | --- | --- |
| [W01: Persist SDD associations](#w01-persist-sdd-associations) | None | R09 |
| [W02: Support inheritance and explicit edits](#w02-support-inheritance-and-explicit-edits) | W01 | R09, R10 |
| [W03: Expose optional CLI associations and context](#w03-expose-optional-cli-associations-and-context) | W02 | R09-R11 |
| [W04: Deliver the explicit SDD workflow](#w04-deliver-the-explicit-sdd-workflow) | W03 | R01-R08, R12 |
| [W05: Install the skill for supported providers](#w05-install-the-skill-for-supported-providers) | W04 | R01-R03 |
| [W06: Support execution handoff and completion reporting](#w06-support-execution-handoff-and-completion-reporting) | W04 | R07, R11, R12 |
| [W07: Document and verify the integrated workflow](#w07-document-and-verify-the-integrated-workflow) | W05, W06 | R01-R12 |

Verify each outcome during implementation. W07 integrates those checks rather than postponing all focused regression testing.

### W01 Persist SDD associations

Allow issues to retain optional associations with an SDD plan and work outcome. Multiple issues may share either or both associations; existing Faz issue IDs remain their identities. The storage model must accommodate SDD and ordinary tasks in the same database without requiring a plan registry or a fixed issue breakdown.

Expected result: associations survive database upgrades and normal persistence operations and can be queried without affecting unassociated issues or losing existing lifecycle, claim, hierarchy, or dependency data.

Depends on: None

Requirements: R09

Verification: AC01, AC02, AC03

### W02 Support inheritance and explicit edits

Let child issues receive their parent's plan and work context without requiring agents to repeat those fields on every creation. Inheritance is a creation-time default for each omitted field. Explicit child values take precedence, and later parent edits or reparenting do not silently alter stored child associations.

Expected result: agents can create, override, edit, or clear associations deliberately while normal claims and status transitions preserve them. Several issues can legitimately share the same associations, and unassociated work remains valid.

Depends on: W01

Requirements: R09, R10

Verification: AC02, AC05

### W03 Expose optional CLI associations and context

Make plan/work association and filtering available through ordinary CLI commands without turning SDD into a required workflow. Agents can select a plan or a specific outcome using the identifiers from its document. Show and successful claim responses provide populated SDD context so an implementation agent can locate the relevant plan; absent values add no output.

Expected result: optional flags support association and exact filtering from repository subdirectories, and linked issues carry useful claim context. Unlinked output, readiness, claim safety, monitor, and Kanban retain their existing behavior. Missing plan documents do not obstruct ordinary lifecycle operations.

Depends on: W02

Requirements: R09-R11

Verification: AC03, AC04, AC06, AC07, AC15

### W04 Deliver the explicit SDD workflow

Provide one human-invoked skill that interprets the entire development brief, asks focused clarification questions, and produces project-wide specifications and architecture plus a numbered outcome-based plan. With no brief, the skill assesses implementation evidence against existing approved documents. Proposed changes remain distinguishable from approved intent, and discrepancies are presented for review.

Expected result: the human can move from ideas to approved requirements, architecture, outcomes, and a validation strategy in one guided workflow. After approval, the agent chooses an appropriate initial Faz issue breakdown, assesses existing coverage, creates needed work, reports readiness, and stops. Subsequent calls preserve useful work and history without enforcing one issue per outcome or silently adopting scope changes.

Depends on: W03

Requirements: R01-R08, R12

Verification: AC09, AC10, AC11, AC12, AC13

### W05 Install the skill for supported providers

Make the complete skill available through Faz's Codex and Claude integrations, including its templates, references, and each provider's explicit invocation settings. Maintain shared workflow content once and preserve unrelated provider configuration and existing task-management hooks.

Expected result: users can install or update the integration locally or globally and explicitly invoke faz-spec-driven-dev with or without a brief. All referenced assets are present, repeated installation is supported, and installation does not enable automatic SDD invocation. Verification uses isolated destinations.

Depends on: W04

Requirements: R01-R03

Verification: AC08

### W06 Support execution handoff and completion reporting

Connect approved SDD work to normal Faz execution while keeping the two authorization steps distinct: plan approval prepares work; a separate instruction starts implementation. Linked plan context supplies the relevant requirements and completion guidance. Agents remain free to refine tasks and report bugs without rewriting a plan manifest.

Expected result: implementation agents can record dated acceptance evidence and a meaningful plan completion summary without invoking SDD or changing approved requirements. Later reconciliation preserves that history and distinguishes discrepancies from intentional supersession. Non-SDD sessions receive no extra query instructions or workflow steps.

Depends on: W04

Requirements: R07, R11, R12

Verification: AC11, AC14, AC15

### W07 Document and verify the integrated workflow

Explain how explicit invocation, argument-bearing development, no-argument reconciliation, optional associations, inheritance, and completion reporting fit together. Verify that the installed skill and CLI support those paths together, including existing databases, evolving task breakdowns, interrupted sessions, and missing local history.

Expected result: usage documentation matches the delivered behavior, automated regression checks cover the relevant CLI and storage invariants, and realistic skill scenarios assess approval and reconciliation behavior. Verification records distinguish executed checks from unavailable host tests. Compatibility with ordinary Faz usage is demonstrated alongside SDD behavior.

Depends on: W05, W06

Requirements: R01-R12

Verification: AC01-AC15

## Code boundaries

- W01: internal/model/issue.go, internal/db/migrate.go, internal/repo/issue_repo.go, and focused tests.
- W02: internal/service/issue_service.go, focused validation helpers where useful, and tests.
- W03: cmd/create.go, cmd/update.go, cmd/list.go, cmd/show.go, cmd/claim.go, root/path helpers where necessary, and tests.
- W04: internal/skillinstaller/bundled/faz-spec-driven-dev/ with SKILL.md and necessary resources.
- W05: internal/skillinstaller/installer.go, provider metadata, cmd/installskills/provider.go, and installation tests.
- W06: SDD skill/plan completion instructions; narrowly scoped generic exceptions in the existing task skill/context only if needed; non-SDD integration checks.
- W07: README.md, command help, integration fixtures, and recorded verification. Keep onboarding free of SDD query instructions.

These are expected implementation areas, not constraints against a justified internal refactor within approved scope.

## Acceptance and verification

### Automated behavior

- AC01: New and existing databases work; repeated upgrades preserve all legacy data and associations.
- AC02: All issue types accept optional associations. Every repository read path round-trips them without scan errors.
- AC03: Multiple issues share both fields while retaining distinct Faz IDs. Neither field nor the pair is unique. Filters return all matching rows; --all includes closed issues.
- AC04: Work matching is exact. W1 does not silently match W01. A combined plan/work query distinguishes identically named outcomes in separate plans.
- AC05: Children inherit both omitted fields. Explicit overrides work. Later parent edits and reparenting do not rewrite children. Lifecycle operations preserve associations.
- AC06: Explicit assignment handles Git-root-relative plans and invalid targets clearly. Missing documents do not block inherited references or ordinary lifecycle operations.
- AC07: Show and successful claim print only populated SDD Plan/SDD Work keys. Unlinked output, readiness, claims, monitor, and Kanban behavior remain compatible.
- AC08: Both provider installers deliver complete assets and explicit-only policy, preserve unrelated configuration, and support local/global reruns in isolated roots.

### Skill behavior

- AC09: A development brief produces scoped specs, architecture, high-level outcomes, acceptance criteria, and planned tests/validation. Meaningful gaps trigger focused questions.
- AC10: New briefs drive development planning while preserving unrelated approved content and earlier plans. Work sections provide a summary table and per-Wxx subsections with intent, expected result, dependency/requirement references, and verification references. Summary and subsection references agree. No Faz IDs, prescribed issue types/counts, or required epic suggestions appear.
- AC11: Approval precedes implementation issue creation. Agents choose appropriate initial decomposition and stop at readiness without claiming or coding. Product scope changes require renewed approval; task restructuring and newly discovered bugs remain possible.
- AC12: Repeated or interrupted sessions inspect existing coverage, preserve claims, and avoid blindly recreating equivalent work. Legitimate multiple issues for one outcome remain allowed. Missing local history is not proof of incomplete implementation.
- AC13: No-argument reconciliation proposes discrepancies rather than silently adopting out-of-scope behavior. Unmet criteria remain gaps despite closed tasks. An unchanged project causes no unnecessary document churn.
- AC14: Completion records contain dates and evidence. Later discrepancy reports preserve prior completion history and distinguish supersession from regression.
- AC15: Non-SDD sessions receive no SDD query instructions or additional SDD loop. SDD filters remain available only as optional CLI capabilities; linked context is conditional.

Use isolated Git/Faz fixtures for stateful checks. Run focused tests during implementation, followed by go test ./..., go vet ./..., golangci-lint run ./..., and relevant race checks. Validate skill packaging and provider metadata. Exercise realistic invocation scenarios in available hosts; structural checks alone do not prove agent behavior. Report unavailable verification honestly.

## Scope limits

- Keep .faz/ ignored and local. No task export, Git synchronization, automatic commits, or history-changing Git commands.
- No association uniqueness constraint, mandatory task manifest, rigid decomposition, or required SDD loop.
- No Kanban changes, automatic SDD invocation, automatic requirement adoption, or implementation after generated-plan approval.
- No additional remote services, plan registry, general specification engine, or system-level installation during implementation tests.

## Review history

- Revision 1: initial implementation proposal prepared; no implementation performed.
- Revision 2: skill renamed to faz-spec-driven-dev; rigid plan-to-issue mapping withdrawn for discussion.
- Revision 3: nonunique nullable plan_id/work_id, inheritance of both fields at creation only, optional exact filtering, conditional claim context, and outcome-based Work sections incorporated. Faz issue IDs/types removed from the plan. Code boundaries and acceptance/verification retained. Implementation awaits review.
- Revision 4: agreed Work format applied to all outcomes: linked summary table followed by concise descriptions, expected results, and dependency, requirement, and verification references. The document contract and acceptance checks require this structure in generated plans. No implementation code changed.
- Revision 4 approved for implementation. User authorized worker subagents for coding, with the primary agent responsible for orchestration, review, and quality. User requested context compaction before implementation starts.
