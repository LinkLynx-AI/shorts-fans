# Review Gates (Full Reviewer Stack)

## Required Review Agents
- `reviewer` as meta review entrypoint.

## `reviewer` Gate Contract
- Meta reviewer orchestrates specialist reviewers in parallel and consolidates findings.
- Specialist coverage: security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.
- No separate UI review gate is required. UI-impact validation remains covered by touched-area validations and runtime smoke evidence in delivery flow.

## Gate Failure Handling
- If any required gate fails, fix issues and return to implementation step.
