---
name: implementation-planner
description: Turn a frozen `Prompt.md` into an execution-ready `Plan.md` with ordered milestones, milestone-level acceptance criteria, validations, and rollback notes. Use when an implementation task is ready to be broken into a concrete delivery plan.
---

# implementation-planner

## Goal
- Convert one implementation spec into a small, auditable execution plan.
- Keep the plan sequential, checkpointed, and easy to resume.

## Input Contract
- `Prompt.md` for the current issue.
- Current repository state and branch context.

## Must-load Resources
- `assets/templates/Plan.md`

## Operating Rules
- Plan only one leaf issue.
- Each milestone must leave the repository in a valid state.
- Prefer outcome-based milestones over file-based checklists.
- Attach validation to every milestone.
- Include rollback or recovery notes when a milestone is risky or stateful.
- Update the plan if execution order changes materially.

## Procedure
1. Read `Prompt.md` and extract the concrete done conditions.
2. Break the work into the smallest meaningful milestones that can be validated independently.
3. For each milestone, define:
- goal
- acceptance criteria
- validation commands
- rollback notes if needed
4. Create or refresh `./.local/codex-memo/<ISSUE>/Plan.md`.
5. Order milestones so the highest-risk contract decisions happen early.
6. Keep plan language imperative and execution-ready.

## Output Contract
- `Plan.md` exists under `./.local/codex-memo/<ISSUE>/`.
- Every milestone has acceptance criteria and validation.
- The plan is specific enough to execute without re-deciding scope mid-run.
