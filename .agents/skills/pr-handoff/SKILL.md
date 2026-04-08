---
name: pr-handoff
description: Produce a Japanese PR-ready handoff from `Prompt.md`, `Plan.md`, `Documentation.md`, and `Verification.md`, including title, summary, acceptance mapping, test results, and review notes. Use when a validated implementation is ready for branch/PR delivery.
---

# pr-handoff

## Goal
- Prepare branch and PR handoff material without re-reading the whole run.
- Keep PR communication aligned with repository policy.

## Input Contract
- `Prompt.md`
- `Plan.md`
- `Documentation.md`
- `Verification.md`
- Issue key, branch name, and target base branch

## Must-load Resources
- `assets/templates/PR.md`

## Operating Rules
- PR title and body must be written in Japanese.
- Explain both what changed and why it changed.
- Map each acceptance criterion to concrete implementation or test evidence.
- If the base branch is `main`, prepare for human review and do not assume auto-merge.
- If the base branch is not `main`, only prepare auto-merge-ready handoff after verification passes and the reviewer remediation loop finishes with no blocking findings.

## Procedure
1. Read the run artifacts instead of reconstructing the story from memory.
2. Create or refresh `./.local/codex-memo/<ISSUE>/PR.md`.
3. Fill:
   - title
   - summary of what changed
   - intent and rationale
   - acceptance criteria mapping
   - validation and smoke results
   - residual risks or follow-ups
   - review focus points
   - issue link
4. Keep the wording concise and reviewer-friendly.

## Output Contract
- `PR.md` exists under `./.local/codex-memo/<ISSUE>/`.
- PR title and body are in Japanese.
- Reviewers can understand scope, intent, evidence, and residual risk without reopening the full run log.
