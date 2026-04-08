# Core Policy

## Critical Merge Policy
- If PR base branch is `main`, do not auto-merge.
- For `main` base branch, mark PR ready for human review, include a review checklist in PR body, and stop.
- For non-`main` base branches, enable auto-merge only after required validations pass and the reviewer remediation loop finishes with no blocking findings.

## Main Branch Operation Permission
- Command execution is allowed by default.
- Any operation on `main` requires explicit user approval in advance.
- Examples: switching to `main`, merge/rebase/cherry-pick targeting `main`, push to `main`, or direct commit on `main`.

## Commit Cadence
- Commit in small, logical units.
- Keep one clear purpose per commit.
- Create commits at meaningful milestones before moving to the next major task.

## Execution Guardrails
- This leaf flow is an orchestration contract. Follow the required skill order and evidence rules.
- Never mix multiple issues into one branch or one PR.
- For sequential runs, never start the next issue before the current issue reaches merge step per policy.

## Linear Connection and Fallback
- Prefer Linear MCP for issue read and write operations.
- If MCP is unavailable, continue from prompt or branch context and keep manual sync notes in `Documentation.md`.
- After each milestone in fallback mode, prepare markdown-ready updates for manual posting.

## Required Memory Files
- Maintain `Prompt.md`, `Plan.md`, `Documentation.md`, `Verification.md`, and `PR.md` for long runs.
- Store them under the worktree-local ignored path `./.local/codex-memo/<LINEAR-IDENTIFIER>/`.
- Do not store run memory under tracked repository paths such as `docs/agent_runs/` or tracked config paths such as `.codex/`.

## PR Convention
- Branch format: `codex/<ISSUE-KEY>-<slug>` when a dedicated issue branch is required.
- PR body must include what and why, acceptance criteria mapping, test steps with results, migration or breaking notes, and Linear link.
- Include validation, reviewer loop, and UI check outcomes in PR body.

## Global Done Criteria
- Acceptance criteria are satisfied.
- Required validations pass.
- Runtime smoke gate passes for non-trivial changes.
- Reviewer remediation loop passes or an explicit blocker is recorded.
- PR handoff is prepared.
- For non-`main` base branches, auto-merge is enabled once required validations pass and the reviewer remediation loop finishes with no blocking findings.
- For `main` base branch, stop at human review required state.
