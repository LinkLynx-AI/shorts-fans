# Review Gates (Full Reviewer Stack)

## Required Review Agent
- `reviewer` as meta review entrypoint.

## Selection Policy
- Do not use this gate as the repository default.
- Use this gate only when at least one repository review escalation condition is met:
- auth, session, permission, access control, payment, or unlock changes
- DB schema, SQL, index, migration, cache, concurrency, infra, or CI changes
- cross-frontend-backend changes, or changes that cross multiple layers or bounded contexts
- shared contract, common platform, architecture, FSD boundary, or package boundary changes
- explicit user request for full review

## Gate Contract
- Meta reviewer orchestrates specialist reviewers in parallel and consolidates findings.
- Specialist coverage: security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.
- Reviewer output must be findings only.
- Each finding must include severity, confidence, file path, and rationale.
- If the diff is clean, the reviewer must return exactly `No findings. Gate clean.`
- A response that does not satisfy this format is invalid for gate classification and does not count as a pass.
- No separate UI review gate is required. UI-impact validation remains covered by touched-area validations and runtime smoke evidence in delivery flow.
- The meta reviewer must prioritize specification alignment first, then regression risk, design integrity, validation adequacy, and readability before specialist optimization advice.

## Remediation Loop
- Review is complete only when the gate is clean.
- If any blocking finding is returned, fix it, rerun the impacted local validation, and rerun `reviewer`.
