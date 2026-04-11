---
name: code-change-verification
description: Run touched-area validation, runtime smoke checks, architecture guard checks, and documentation sync checks, then record the results in `Verification.md`. Use after implementation and before reviewer gates or PR handoff.
---

# code-change-verification

## Goal
- Verify the changed code locally before reviewer gates.
- Make validation evidence explicit and auditable.
- Pull architecture and docs checks into the same verification pass.

## Input Contract
- Current diff or touched-file set.
- `Prompt.md`, `Plan.md`, and `Documentation.md`.
- Current branch context.

## Must-load Resources
- `assets/templates/Verification.md`
- `../architecture-guard/SKILL.md`
- `../docs-sync/SKILL.md`

## Validation Matrix
- Backend changes:
  - `cd backend && go test ./... && go vet ./...`
- Backend review-ready coverage when the touched backend packages are inside the coverage target defined by `scripts/check-backend-coverage.sh`:
  - `make backend-coverage-check BACKEND_COVERAGE_MIN=80 BACKEND_COVERAGE_PROFILE=.artifacts/backend-coverage.out`
- Backend concurrency changes:
  - `cd backend && go test -race ./...`
- Backend dependency changes:
  - `cd backend && go mod tidy && govulncheck ./...`
- Frontend changes:
  - `cd frontend && pnpm lint && pnpm typecheck && pnpm test:unit`
- Frontend review-ready coverage when logic or tests under `src/shared`, `src/entities`, `src/features`, or `src/widgets` changed:
  - `cd frontend && pnpm test:coverage:check`
- Docs-only changes:
  - Run docs-sync and any doc-targeted checks only.

## Runtime Smoke Rules
- For non-trivial backend behavior changes, exercise the affected endpoint or local runtime path.
- For non-trivial frontend UI changes, run a smoke flow on the affected route.
- Do not use Playwright unless the user explicitly requested E2E validation.
- If a smoke check is skipped, record why the change is trivial enough to skip it.

## Procedure
1. Determine the touched areas from the diff.
2. Create or refresh `./.local/codex-memo/<ISSUE>/Verification.md`.
3. Run the required command set from the validation matrix.
4. Load `architecture-guard` and record its result.
5. Load `docs-sync` and either update docs or record a no-change rationale.
6. Run runtime smoke checks when the change is non-trivial.
7. Record every command, result, artifact, and skip rationale in `Verification.md`.
8. Do not start reviewer gates until local verification is green or blocked by an explicit external dependency.

## Guardrails
- Never hide a failing command.
- Prefer rerunning the exact failed command after a fix before broadening the validation scope again.
- If backend coverage is skipped because the diff only touches coverage-excluded backend paths, record the exact skip rationale.
- If a required tool is unavailable, record that explicitly and explain the residual risk.

## Output Contract
- `Verification.md` exists under `./.local/codex-memo/<ISSUE>/`.
- Validation coverage matches the actual touched areas, including backend or frontend coverage gates when applicable.
- Architecture and docs impact are accounted for before reviewer gates.
