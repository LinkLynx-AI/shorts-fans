---
name: implementation-spec-writer
description: Freeze one repo task into a run-local `Prompt.md` by reading only the relevant repository source-of-record docs, contracts, and issue context before code changes begin. Use when an implementation needs explicit goals, non-goals, constraints, acceptance criteria, and source references.
---

# implementation-spec-writer

## Goal
- Turn the current request into a concrete implementation spec before writing code.
- Keep repository knowledge in docs and contracts as the system of record.
- Make missing or undecided behavior explicit instead of guessing.

## Input Contract
- User request, issue text, or equivalent task context.
- Repository area that will be touched.
- Relevant contract or policy documents when they exist.

## Must-load Resources
- `assets/templates/Prompt.md`

## Operating Rules
- Read only the source-of-record documents that are relevant to the touched area.
- Prefer repository docs over memory. Treat undocumented behavior as unresolved.
- Capture product requirements, constraints, and acceptance criteria. Do not invent implementation details here.
- If a requirement is still ambiguous after checking the repo, mark it as `未確定` or an open question.
- If the task conflicts with repository policy, stop and surface the conflict before implementation.

## Source-of-Record Routing
- Frontend UI or state work:
  - `docs/TYPESCRIPT.md`
  - Relevant `docs/contracts/*.md`
- Backend or API work:
  - `docs/GO.md`
  - `docs/BACKEND_STRUCTURE.md`
  - Relevant `docs/contracts/*.md`
- Infra or media workflow work:
  - Relevant `docs/infra/*.md`
  - Relevant `docs/contracts/*.md`
- Product or behavior clarification:
  - `docs/README.md`
  - `docs/ssot/LOCAL_INDEX.md`

## Procedure
1. Resolve the leaf issue boundary. Confirm the run covers exactly one deliverable issue.
2. Read only the minimal set of repository documents needed to define scope and constraints.
3. Create or refresh `./.local/codex-memo/<ISSUE>/Prompt.md`.
4. Fill `Prompt.md` with:
- goals
- non-goals
- deliverables
- done conditions
- constraints
- exact source references
- explicit open questions or unresolved items
5. If a material ambiguity remains and cannot be resolved from repo context, stop and ask once before implementation.

## Output Contract
- `Prompt.md` exists under `./.local/codex-memo/<ISSUE>/`.
- Every acceptance criterion traces back to user input or a repository document.
- Unspecified behavior is marked as unresolved, not silently decided.
