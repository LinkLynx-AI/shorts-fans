# Review Gates (Full Reviewer Stack)

## Required Review Agent
- `reviewer` as meta review entrypoint.

## Gate Contract
- Meta reviewer orchestrates specialist reviewers in parallel and consolidates findings.
- Specialist coverage: security, correctness, performance, test quality, and coding rules.
- Blocking rule: block when at least one `P1` or higher finding has confidence `>= 0.65`.
- No separate UI review gate is required. UI-impact validation remains covered by touched-area validations and runtime smoke evidence in delivery flow.

## Remediation Loop
- Review is complete only when the gate is clean.
- If any blocking finding is returned, fix it, rerun the impacted local validation, and rerun `reviewer`.
