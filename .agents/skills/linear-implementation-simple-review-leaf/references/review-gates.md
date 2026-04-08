# Review Gates (Simple Reviewer)

## Required Review Agent
- `reviewer_simple` as unified review entrypoint.

## Selection Policy
- Use this gate by default for leaf issue delivery.
- Escalate to the full reviewer stack only when at least one repository review escalation condition is met:
- auth, session, permission, access control, payment, or unlock changes
- DB schema, SQL, index, migration, cache, concurrency, infra, or CI changes
- cross-frontend-backend changes, or changes that cross multiple layers or bounded contexts
- shared contract, common platform, architecture, FSD boundary, or package boundary changes
- explicit user request for full review

## Gate Contract
- Unified review covers security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.
- Reviewer output must be findings only.
- Each finding must include severity, confidence, file path, and rationale.
- If the diff is clean, the reviewer must return exactly `No findings. Gate clean.`
- A response that does not satisfy this format is invalid for gate classification and does not count as a pass.
- UI-impact handling stays inside `reviewer_simple`. No separate `reviewer_ui_guard` or `reviewer_ui` gate is required.
- Review order inside this gate must prioritize specification alignment first, then regression risk, design integrity, validation adequacy, and readability.

## Remediation Loop
- Review is complete only when the gate is clean.
- If any blocking finding is returned, fix it, rerun the impacted local validation, and rerun `reviewer_simple`.
