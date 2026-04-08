---
name: implementation-runbook
description: Execute one implementation plan with tight scope control, existing-pattern reuse, small diffs, and continuous run logging in `Documentation.md`. Use when code changes are ready to be made from an approved `Prompt.md` and `Plan.md`.
---

# implementation-runbook

## Goal
- Apply the plan without widening scope.
- Keep execution legible through a live `Documentation.md`.
- Reuse repository patterns before introducing new abstractions.

## Input Contract
- `Prompt.md`
- `Plan.md`
- Current branch and working tree state

## Must-load Resources
- `assets/templates/Documentation.md`

## Operating Rules
- Inspect existing implementations before introducing a new pattern.
- Follow `Plan.md` as the default execution order.
- Keep diffs small and reversible.
- Do not mix unrelated cleanup into the same run.
- If scope must expand, stop and ask before continuing.
- This skill is the runbook. Do not rely on a separate per-run `Implement.md`.

## Procedure
1. Read `Prompt.md` and `Plan.md`.
2. Inspect the touched area for existing architecture and naming patterns.
3. Create or refresh `./.local/codex-memo/<ISSUE>/Documentation.md`.
4. Execute milestones in order.
5. After each milestone, update `Documentation.md` with:
- current status
- next step
- decisions and rationale
- demo or run notes
- known issues or follow-ups
6. If implementation reveals a material plan change, update `Plan.md` and log the reason in `Documentation.md`.
7. Run the milestone validation before moving on.

## Output Contract
- Code changes stay within the frozen leaf scope.
- `Documentation.md` is current enough for a handoff or resume.
- Plan deviations are documented with reasons.
