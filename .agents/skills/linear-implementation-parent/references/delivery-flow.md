# Delivery Flow (Parent Issue)

## Parent Issue Handling
1. Collect all child issues.
2. Determine execution order by:
- explicit order markers in title or body
- dependency relations
- fallback layer order: contract/domain, backend, frontend data/state, frontend UI, end-to-end
3. Execute child issues sequentially in that order.
4. Treat a child issue as complete only when all mandatory delivery loop steps and evidence are complete.

## Mandatory Per-child Delivery Loop
1. Create a dedicated branch for the child issue.
2. Run sub-agents for exploration, implementation, and validation.
3. Execute review gates defined in `references/review-gates.md`.
4. If review or validation gate fails, fix and repeat from implementation.
5. For non-trivial changes, run runtime smoke gate:
- start the affected app(s) in `backend/` and/or `frontend/` if local runtime is available
- verify changed route(s) or endpoint(s) and check for fatal runtime errors
6. If runtime smoke fails, fix and repeat from implementation.
7. Open PR from child branch to the parent branch.
8. Apply merge policy from `references/core-policy.md`.
9. Start next child issue only after merge step and evidence are complete.

## Required Per-child Evidence
- Child issue key and branch name.
- Touched-area validation results and additional issue-specific validation results.
- Review gate result and blocking finding disposition.
- Runtime smoke gate result for non-trivial changes, or explicit skip rationale for trivial changes.
- PR URL, base branch, and merge or auto-merge status.

## Interruption Recovery
1. Detect in-progress child issue from branch, commits, and `Documentation.md`.
2. Complete missing steps for that child issue first.
3. Only then start the next child issue.
