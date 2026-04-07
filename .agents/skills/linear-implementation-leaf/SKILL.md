---
name: linear-implementation-leaf
description: Execute leaf issue runs for this repository with one issue equals one PR delivery. Use when the request starts from a child issue or from a standalone smallest-unit issue that does not require parent-child decomposition. This variant uses the full reviewer stack with meta review (`reviewer`) and specialist reviewers.
---

# linear-implementation-leaf

## Goal
- Execute one leaf issue run with one issue equals one PR delivery.
- Cover both start modes: child issue start and standalone smallest-unit issue start.

## Input Contract
- Child issue identifier or standalone issue identifier.
- Acceptance criteria and target repository context.
- Branch context to determine whether a dedicated branch is required.

## Routing and Scope Boundary
- Use this skill for leaf runs only.
- Supported starts:
  - Child issue start (issue belongs to a parent and is requested directly).
  - Standalone smallest-unit start (no parent-child decomposition required).
- If the issue type is unclear, ask exactly one clarifying question before execution.

## Must-load References
- `references/core-policy.md`
- `references/delivery-flow.md`
- `references/review-gates.md`
- `assets/memory_templates/Prompt.md`
- `assets/memory_templates/Plan.md`
- `assets/memory_templates/Implement.md`
- `assets/memory_templates/Documentation.md`

## Hard-stop Guardrails
- Deliver exactly one issue in this run.
- Do not start sibling issues unless explicitly requested.
- Follow the branch policy in delivery flow based on start mode.
- If required evidence is missing, stop and complete evidence before closing the run.

## Done Criteria
- The target leaf issue reaches PR step and merge status follows policy for the base branch.
- Validation, review gate, and runtime smoke gate evidence are recorded.
