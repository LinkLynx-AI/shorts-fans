---
name: linear-implementation-simple-review-leaf
description: Orchestrate one leaf issue delivery by routing through repo-scoped implementation skills from spec freeze to PR handoff. Use when the request starts from a child issue or a standalone smallest-unit issue. This variant enforces the unified reviewer gate (`reviewer_simple`) until the gate is clean.
---

# linear-implementation-simple-review-leaf

## Goal
- Orchestrate one leaf issue run from spec freeze to PR handoff.
- Stay thin: this skill routes work to focused implementation skills instead of owning every procedure directly.
- Cover both start modes: child issue start and standalone smallest-unit issue start.

## Input Contract
- Child issue identifier or standalone issue identifier.
- Acceptance criteria and target repository context.
- Branch context to determine whether a dedicated branch is required.

## Routing and Scope Boundary
- Use this skill for leaf runs only.
- Supported starts:
  - Child issue start (issue belongs to a parent and is requested directly).
  - Standalone smallest-unit start (no parent-child decomposition required).
- If the issue type is unclear, ask exactly one clarifying question before execution.

## Review Selection Policy
- This is the default leaf skill for repository work.
- Prefer this skill when the change is local and does not need specialist review decomposition.
- Do not escalate to `linear-implementation-leaf` unless at least one of the repository review escalation conditions is met or the user explicitly requests full review.

## Must-load Resources
- `references/orchestration-policy.md`
- `references/core-policy.md`
- `references/delivery-flow.md`
- `references/review-gates.md`
- `references/reviewer-profile.md`
- `../implementation-spec-writer/SKILL.md`
- `../implementation-planner/SKILL.md`
- `../implementation-runbook/SKILL.md`
- `../code-change-verification/SKILL.md`
- `../reviewer-remediation-loop/SKILL.md`
- `../pr-handoff/SKILL.md`

## Orchestration Order
1. Resolve start mode and branch handling from `references/orchestration-policy.md`.
2. Run `implementation-spec-writer` and produce `Prompt.md`.
3. Run `implementation-planner` and produce `Plan.md`.
4. Run `implementation-runbook` and keep `Documentation.md` current during implementation.
5. Run `code-change-verification` and produce `Verification.md`.
6. Run `reviewer-remediation-loop` with `references/reviewer-profile.md`.
7. Run `pr-handoff` and produce `PR.md`.

## Hard-stop Guardrails
- Deliver exactly one issue in this run.
- Do not start sibling issues unless explicitly requested.
- Do not skip local verification before reviewer gates.
- Do not treat the reviewer gate as advisory. Keep looping until `references/reviewer-profile.md` passes or an explicit blocker is documented.
- If required evidence is missing, stop and complete evidence before closing the run.
- If scope must expand beyond the approved leaf issue, stop and ask before continuing.

## Done Criteria
- `Prompt.md`, `Plan.md`, `Documentation.md`, `Verification.md`, and `PR.md` exist for the issue.
- Local validation and runtime smoke evidence are recorded.
- The reviewer remediation loop ends with no blocking findings or an explicit documented blocker.
- The target leaf issue reaches PR handoff state for the current branch policy.
