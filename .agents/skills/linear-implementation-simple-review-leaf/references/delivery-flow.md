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
2. Implement scoped changes.
3. Commit in small logical units.
4. Run repository validations for the touched areas and any issue-specific checks:
- backend changes: `cd backend && go test ./... && go vet ./...`
- frontend changes: `cd frontend && pnpm lint && pnpm typecheck && pnpm test:unit`
- combined changes: run both backend and frontend validations
- do not run `Playwright` or `pnpm test:e2e` unless the user explicitly requests it
5. Execute review gates defined in `references/review-gates.md`.
6. If review or validation gate fails, fix and repeat from implementation.
7. For non-trivial changes, run runtime smoke gate:
- start the affected app(s) in `backend/` and/or `frontend/` if local runtime is available
- verify changed route(s) or endpoint(s) and check for fatal runtime errors
8. If runtime smoke fails, fix and repeat from implementation.
9. Open one PR to `main`.
10. Apply merge policy from `references/core-policy.md`.
11. Do not start sibling issues unless explicitly requested.

## Required Evidence
- Leaf issue key and branch name.
- Start mode decision (child issue start or standalone start).
- Touched-area validation results and additional issue-specific validation results.
- Review gate result with blocking finding disposition.
- Runtime smoke gate result for non-trivial changes, or explicit skip rationale for trivial changes.
- PR URL, base branch, and merge policy state.
