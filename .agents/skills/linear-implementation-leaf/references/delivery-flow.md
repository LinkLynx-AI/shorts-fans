# Delivery Flow (Leaf Issue)

## Start Mode Decision
- Child issue start: issue is a child and requested directly.
- Standalone smallest-unit start: no parent-child decomposition needed.

If start mode is unclear, ask one clarifying question before execution.

## Branch Rule by Start Mode
- Child issue start: create a dedicated child issue branch.
- Standalone smallest-unit start: use the current branch as-is and do not create extra branch.

## Mandatory Delivery Loop
1. Confirm scope is exactly one leaf issue.
2. Run `implementation-spec-writer` and freeze the issue into `Prompt.md`.
3. Run `implementation-planner` and create `Plan.md`.
4. Run `implementation-runbook` and implement scoped changes while keeping `Documentation.md` current.
5. Run `code-change-verification` for touched-area validation, architecture checks, docs sync, and runtime smoke.
6. Run `reviewer-remediation-loop` with the active reviewer profile from `references/reviewer-profile.md`.
7. If the reviewer or validation gate fails, fix and repeat the affected implementation and verification steps.
8. Run `pr-handoff` and prepare `PR.md`.
9. Apply merge policy from `references/core-policy.md`.
10. Do not start sibling issues unless explicitly requested.

## Required Evidence
- Leaf issue key and branch name.
- Start mode decision (child issue start or standalone start).
- `Prompt.md`, `Plan.md`, `Documentation.md`, `Verification.md`, and `PR.md`.
- Touched-area validation results and additional issue-specific validation results.
- Reviewer remediation loop result and blocking finding disposition for every remediation loop pass.
- Runtime smoke gate result for non-trivial changes, or explicit skip rationale for trivial changes.
- PR URL or PR handoff state, base branch, and merge policy state.
