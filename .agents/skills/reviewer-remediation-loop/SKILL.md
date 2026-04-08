---
name: reviewer-remediation-loop
description: Run the required reviewer agent after local verification, fix blocking findings, and repeat until the review gate passes or an explicit external blocker remains. Use when a delivery flow must enforce reviewer approval instead of treating review as advisory.
---

# reviewer-remediation-loop

## Goal
- Enforce the reviewer gate after implementation.
- Prevent the common failure mode where review runs once and findings are left unresolved.

## Input Contract
- Current diff and branch state.
- `Verification.md` with local validation evidence.
- Review profile supplied by the caller.

## Review Profile Requirements
- The caller must provide:
  - reviewer agent name
  - blocking rule
  - expected review coverage

## Procedure
1. Load the caller-supplied review profile.
2. Confirm local validation is complete enough to justify a reviewer run.
3. Run the required reviewer agent.
4. Classify findings using the profile's blocking rule.
5. If blocking findings exist:
   - fix the highest-severity findings first
   - rerun the targeted local validation affected by the fix
   - rerun the same reviewer gate
6. Repeat the loop until:
   - no blocking findings remain, or
   - an explicit external blocker or user decision prevents further progress
7. Record every review pass, finding disposition, and fix loop result in `Verification.md`.
8. Do not proceed to PR handoff while blocking findings remain.

## Guardrails
- A single reviewer pass is not enough unless the gate is clean.
- Do not downgrade a blocking finding to non-blocking without written rationale.
- If a finding is intentionally deferred, record why it is safe to defer and who accepted the risk.

## Output Contract
- Reviewer history is recorded in `Verification.md`.
- The final state is either:
  - gate passed with no blocking findings, or
  - explicitly blocked with a documented reason
