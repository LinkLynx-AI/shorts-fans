---
name: linear-ticketing
description: Create and structure Linear issues for linklinx-AI using parent and child issue decomposition with explicit execution order, dependencies, acceptance criteria, and Do/Don't scope. Use when users ask to split features into actionable tasks, create parent and child issues, or prepare implementation-ready issue plans without writing code.
---

# linear-ticketing

## Goal
- Decompose large features into one parent feature issue and multiple child task issues.
- Fix execution order explicitly so long-running implementation does not drift.
- Keep child tasks reviewable at one task equals one PR size.
- Do not implement code in this skill.

## Required Input
- Feature summary with business purpose.
- Target area such as frontend, backend, or infra.

## Optional Input
- Team, project, label policies in Linear.
- Expected stack and constraints.

## Missing Information Policy
- Do not block issue creation for missing details.
- Record unresolved points in parent issue Open questions as TBD items.

## Output Contract
Create the following issue structure.

### Parent Feature Issue Required Sections
1. Overview
2. Goals with done conditions
3. Non-goals
4. Requirements with functional and non-functional points
5. Security and abuse considerations
6. Open questions
7. Execution plan with child list, order, and dependencies
8. Link to codex implementation operation rules

### Child Task Issue Required Sections
- Parent reference
- Order field, mandatory
- Do and Don't scope
- Acceptance Criteria in objectively testable form
- How to test with commands and steps
- Dependencies with blocked by and blocks
- Definition of Ready

Template text is available at `assets/templates.md`.

## Standard Decomposition for Messaging MVP
- [01] Database schema and migration
- [02] API create message
- [03] API list messages with pagination
- [04] Realtime publish via websocket
- [05] UI timeline view
- [06] UI composer send flow
- [07] End to end send and receive path
- [08] Observability and logging baseline

## How to Determine Order
- Prefer strong dependency direction: database, then API, then realtime, then UI, then end to end.
- Resolve blocked by issues first.
- If a child task cannot close in one PR, split it and reassign order.

## Linear Write Path
Pick one method.

1. Linear MCP preferred for local Codex execution.
- `codex mcp add linear --url https://mcp.linear.app/mcp`
- Or configure `~/.codex/config.toml` with `mcp_servers.linear`.

2. If MCP is unavailable.
- Output parent and child drafts in markdown including title, body, order, and dependencies for manual copy.

## Checklist
- Parent includes goals, non-goals, and open questions.
- Every child includes order, acceptance criteria, how to test, and do and don't scope.
- Every child is small enough for one PR.
- Parent issue clearly states full child execution order.
