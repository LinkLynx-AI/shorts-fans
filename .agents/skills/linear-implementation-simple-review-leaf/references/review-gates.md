# Review Gates (Simple Reviewer)

## Required Review Agents
- `reviewer_simple` as unified review entrypoint.

## `reviewer_simple` Gate Contract
- Unified review covers security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.
- UI-impact handling stays inside `reviewer_simple`. No separate `reviewer_ui_guard` or `reviewer_ui` gate is required.

## Gate Failure Handling
- If any required gate fails, fix issues and return to implementation step.
