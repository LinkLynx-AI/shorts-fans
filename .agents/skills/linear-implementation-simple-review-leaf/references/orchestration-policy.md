# Orchestration Policy

## Start Mode Decision
- Child issue start: issue is a child and requested directly.
- Standalone smallest-unit start: no parent-child decomposition is needed.

If start mode is unclear, ask one clarifying question before execution.

## Branch Rule by Start Mode
- Child issue start: create a dedicated branch `codex/<ISSUE-KEY>-<slug>`.
- Standalone smallest-unit start: use the current branch as-is.

## Run Artifacts
- Maintain run-local artifacts under `./.local/codex-memo/<ISSUE>/`.
- Required artifacts:
  - `Prompt.md`
  - `Plan.md`
  - `Documentation.md`
  - `Verification.md`
  - `PR.md`

## Scope and Stop Conditions
- Deliver exactly one leaf issue in this run.
- Do not start sibling issues unless explicitly requested.
- Stop and ask before expanding scope, changing branch policy, or relying on undocumented behavior.
