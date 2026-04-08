# Review Gates (Simple Reviewer)

## Required Review Agent
- `reviewer_simple` as unified review entrypoint.

## Gate Contract
- Unified review covers security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.
- Reviewer output must be findings only.
- Each finding must include severity, confidence, file path, and rationale.
- If the diff is clean, the reviewer must return exactly `No findings. Gate clean.`
- A response that does not satisfy this format is invalid for gate classification and does not count as a pass.
- UI-impact handling stays inside `reviewer_simple`. No separate `reviewer_ui_guard` or `reviewer_ui` gate is required.

## Remediation Loop
- Review is complete only when the gate is clean.
- If any blocking finding is returned, fix it, rerun the impacted local validation, and rerun `reviewer_simple`.
