# Review Gates (Full Reviewer Stack)

## Required Review Agents
- `reviewer` as meta review entrypoint.
- `reviewer_ui_guard` to detect UI-impact changes.
- `reviewer_ui` only when UI guard says UI changes exist.

## `reviewer` Gate Contract
- Meta reviewer orchestrates specialist reviewers in parallel and consolidates findings.
- Specialist coverage: security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.

## UI Gate Contract
- If UI guard is false, mark UI checks as skipped with rationale.
- If UI guard is true, run UI review and include result as pass or fail evidence.

## Gate Failure Handling
- If any required gate fails, fix issues and return to implementation step.
