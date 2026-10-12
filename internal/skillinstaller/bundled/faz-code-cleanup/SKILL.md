---
name: faz-code-cleanup
description: Explicitly assess unnecessary code, tests, complexity, and organizational friction, then prepare evidence-backed Faz cleanup work for approval. Use only when the human invokes this skill; it does not perform cleanup automatically.
---

# Faz code cleanup

Use only after direct human invocation. Treat trailing text as the requested scope and constraints. The invocation authorizes inspection and Faz planning, not implementation. Do not invoke this skill automatically after SDD, during ordinary reviews, or from another workflow.

Propose behavior-preserving cleanup, not new features or an architectural rewrite. Do not modify project files, run autofixes, or rewrite approved specifications or completed plans during this assessment.

## Establish context

1. Read project instructions, `handoff.md`, and relevant `faz-specs/` documents when present. Verify their statements against current implementation; missing handoff or specs files do not block inspection.
2. Use the treelines exploration workflow before reading Go, Python, or Rust source: index and wait for completion, inspect the overview, then narrow with symbol and relationship queries before selective source reads. For other languages, use available language-appropriate inspection tools. Report exploration limitations rather than claiming unsupported certainty.
3. Inspect existing Faz issues and claims to avoid duplicate proposals. Read-only investigation needs no task or claim. If Faz or a repository is unavailable, report findings and the inability to prepare issues; do not initialize a repository solely for tracking.

## Find worthwhile cleanup

- Unreachable code, obsolete compatibility layers, unused dependencies, duplicate implementations, and abandoned artifacts.
- Tests that exercise obsolete code, duplicate coverage without distinct value, or assert implementation details without protecting meaningful behavior.
- Unnecessary indirection, confusing names, and avoidable complexity that reduce readability.
- Comments and docstrings that repeat the code or contain obsolete workflow metadata. Prefer concise technical explanations, preserving rationale, constraints, and non-obvious behavior. Remove Faz/SDD identifiers that serve only as task annotations; preserve tooling directives, legal notices, and meaningful references.
- Organization that obscures responsibilities or creates unnecessary coupling. Suggest lightweight domain-oriented organization only when it addresses demonstrated navigation, ownership, or coupling problems. Respect language and framework conventions; do not add DDD layers or patterns merely for conformity.

Before proposing removal, check runtime entry points, public consumers, indirect usage, supported configurations, and compatibility obligations. A missing treelines caller or test-only usage is not proof of dead code. Public APIs and legitimate test helpers may have no ordinary production callers.

Judge tests by the behavior, contract, or regression they protect, not their size or count. For a proposed test removal, explain what protection would disappear and whether remaining or replacement coverage preserves it. Passing tests alone do not prove a deletion is safe.

## Prepare findings and Faz work

1. For each actionable finding, identify its location, evidence, proposed change, expected benefit, risk, and verification needed. Separate confirmed cleanup from uncertain candidates. Do not invent findings to fill a quota; report when no worthwhile changes are supported.
2. Automatically create or reuse issues for identified bugs, including bugs reported by the user and confirmed during inspection. Filing a bug does not authorize fixing it.
3. Prepare or reuse atomic cleanup tasks or chores with rich descriptions carrying the finding and acceptance criteria. Group epics by distinct domain or deliverable where useful, never merely by the whole project. Small scope may need one or no epics. Record only real prerequisites, explain blockers, and define non-conflicting file ownership for parallel work. Preserve existing claims and approved scope.
4. Keep uncertain removal proposals out of ready implementation work until their evidence is resolved. Record consequential choices as decision issues for the user; do not make product, compatibility, or architecture decisions autonomously.
5. Report the findings, prepared issue IDs, verification performed, limitations, and decisions needed. Stop for human approval without claiming or starting cleanup. Subsequent approved implementation follows the normal Faz workflow; this skill does not start orchestration automatically.
