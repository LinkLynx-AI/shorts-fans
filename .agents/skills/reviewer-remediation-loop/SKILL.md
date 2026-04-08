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
  - required response format

## Procedure
1. Load the caller-supplied review profile.
2. Confirm local validation is complete enough to justify a reviewer run.
3. Run the required reviewer agent and explicitly require the profile's response format.
4. Validate that the reviewer output is consumable for gate classification.
5. If the reviewer output omits required finding fields or does not explicitly state `No findings. Gate clean.`:
   - do not treat the pass as clean
   - record `Reviewer output invalid for gate classification.`
   - rerun the same reviewer gate once with the response format restated
6. Classify findings using the profile's blocking rule only after the output is consumable.
7. If blocking findings exist:
   - fix the highest-severity findings first
   - rerun the targeted local validation affected by the fix
   - rerun the same reviewer gate
8. Repeat the loop until:
   - no blocking findings remain, or
   - an explicit external blocker or user decision prevents further progress
9. Record every review pass, finding disposition, and fix loop result in `Verification.md`.
10. Do not proceed to PR handoff while blocking findings remain.

## Guardrails
- A single reviewer pass is not enough unless the gate is clean.
- Do not downgrade a blocking finding to non-blocking without written rationale.
- If a finding is intentionally deferred, record why it is safe to defer and who accepted the risk.

## Output Contract
- Reviewer history is recorded in `Verification.md`.
- The final state is either:
  - gate passed with no blocking findings, or
  - explicitly blocked with a documented reason
- A gate pass requires either:
  - consumable findings with no blocking issue under the profile's rule, or
  - the exact clean response `No findings. Gate clean.`
- A reviewer response that fails the required format is not a gate pass.
