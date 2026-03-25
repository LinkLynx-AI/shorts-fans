---
name: linear-implementation-parent
description: Execute linklinx-AI parent Linear issues by delivering child issues sequentially with one issue equals one PR. Use when the request is a parent issue that has child issues or explicitly asks to run full parent-to-child implementation flow. This variant uses the full reviewer stack with meta review (`reviewer`) and specialist reviewers.
---

# linear-implementation-parent

## Goal
- Execute one parent issue by completing child issues in strict order with one issue equals one PR delivery.
- Preserve existing merge policy, review gates, and runtime smoke requirements.

## Input Contract
- Parent Linear issue identifier or equivalent issue context.
- Child issue list, or enough context to resolve children from Linear.
- Repository branch context and acceptance criteria for each child.

## Routing and Scope Boundary
- Use this skill only for parent issue runs.
- Do not use this skill for a single child issue or standalone issue request.
- If it is unclear whether the input is a parent issue, ask exactly one clarifying question before execution.

## Must-load References
- `references/core-policy.md`
- `references/delivery-flow.md`
- `references/review-gates.md`
- `assets/memory_templates/Prompt.md`
- `assets/memory_templates/Plan.md`
- `assets/memory_templates/Implement.md`
- `assets/memory_templates/Documentation.md`

## Hard-stop Guardrails
- Follow delivery flow in order. Do not skip mandatory evidence.
- Keep one active child issue at a time. Never batch multiple child issues into one PR.
- If interruption happens, resume the in-progress child issue first.
- If policy conflicts arise, follow repository or developer policy first and log the decision in `Documentation.md`.

## Done Criteria
- Every child issue in the parent scope reaches merge step per policy.
- Validation, review gates, UI gates (when applicable), and runtime smoke gate are recorded for each child.
- PR status and merge policy outcome are recorded for each child.
