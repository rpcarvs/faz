# PLANnn: Short plan title

Approval: Awaiting human approval
Implementation: Not started
Revision: 1
Last reviewed: YYYY-MM-DD

## Purpose and scope

State the approved problem, intended outcome, and boundaries. This plan describes outcomes, not a required Faz issue decomposition. Agents choose the appropriate issue types, hierarchy, and number of work items as implementation proceeds.

## Scope exclusions

List intentionally excluded behavior, deferred work, and assumptions that need separate approval before expansion.

## Completion summary

Whole-plan completion: Not verified

Retain the following close-out instructions in the generated plan. Implementation agents follow them without invoking the SDD skill. Do not infer completion from closed Faz tasks or mirror task counts.

### Close-out instructions

1. Keep `Implementation: Not started` until authorized implementation begins, then set `In progress`. Record dated evidence by outcome, including partial verification and remaining acceptance checks. Do not alter approved requirements while recording progress.
2. Verify every applicable acceptance criterion. Human validation is required when a criterion calls for it or the agent determines human testing or confirmation is needed. Record what to check and why. When ready for that review, set `Implementation: Awaiting human validation` and ask the human; the requested confirmation now gates completion.
3. If the human reports a failure before completion, record it, return to `In progress`, and create or reopen Faz work linked to this plan and outcome. Correct it, repeat relevant verification, and request human revalidation. Preserve earlier evidence and record corrections. Fixing approved behavior needs no scope revision; changed requirements need an approved revision to this unfinished plan.
4. Once all applicable criteria pass and any requested human validation is confirmed, record dated whole-plan evidence and the human confirmation, if applicable. Update `Whole-plan completion` to `Verified` with the date and `Implementation` to `Implemented` before reporting completion. Clear contextual confirmation such as "it works" confirms the requested checks, not unrelated unverified criteria. Without required or requested human validation, agent verification is sufficient.
5. `Approval` and `Revision` concern document-scope approval, not implementation acceptance; completion evidence alone does not change them. After `Implemented`, keep this plan as a historical record. New issues, including discrepancies found by reconciliation, require a new explicitly invoked SDD plan or ordinary Faz work outside this plan; do not reopen or edit this plan for them.

### Outcome completion records

Record each outcome's date, verification evidence, and remaining applicable criteria here. Include pending human checks and their eventual confirmation or reported failures.

## Requirements

- R01: Brief statement of the requirement this plan addresses.

## Work outcomes

Wxx entries describe outcomes, not Faz task slots. Agents choose and evolve the issue types, hierarchy, count, and dependencies required to deliver them. Several work items may contribute to one outcome, and implementation may add bugs or other necessary work.

List only real outcome prerequisites, not a preferred implementation order. Independent portions of dependent outcomes may proceed in parallel.

| Outcome | Depends on | Requirements |
| --- | --- | --- |
| [W01: Short outcome title](#w01-short-outcome-title) | None | R01 |

### W01 Short outcome title

Describe the product or technical outcome and its boundaries. Do not prescribe Faz task IDs, issue types, counts, or an epic structure.

Expected result: State the observable result when this outcome is complete.

Depends on: None

Requirements: R01

Verification: AC01

## Code boundaries

- Identify the expected components, packages, services, or external boundaries. These are guidance, not a prohibition on justified refactoring. Identify shared-file or shared-resource constraints that affect parallel implementation.

## Acceptance and verification

### AC01 Short acceptance criterion

State the observable acceptance condition and the test, inspection, or evidence that verifies it. Identify any required human validation and what the human must check.

## Review history

- Revision 1: Draft prepared YYYY-MM-DD. Awaiting approval.
