---
name: architecture-guard
description: Check the current diff for repository-specific architecture violations across frontend FSD boundaries, backend package boundaries, contract alignment, and policy constraints. Use during verification when code changes might drift from the repo's intended structure.
---

# architecture-guard

## Goal
- Catch repository-specific architectural drift before reviewer handoff.
- Keep the diff aligned with documented frontend, backend, and contract boundaries.

## Input Contract
- Current diff or touched-file set.
- Relevant repo documents for the touched area.

## Frontend Guard
- Read `docs/TYPESCRIPT.md` when frontend code is touched.
- Check:
  - FSD dependency direction
  - Public API only, no deep imports
  - no domain logic in `shared`
  - `useEffect` only for external synchronization
  - Zod parsing at external boundaries
  - no default exports except Next.js required entry files

## Backend Guard
- Read `docs/GO.md` and `docs/BACKEND_STRUCTURE.md` when backend code is touched.
- Check:
  - package responsibility and naming
  - `internal/` placement and dependency direction
  - exported symbol doc comments
  - `context.Context` as first argument, not stored on structs
  - explicit error handling and wrapping policy
  - concurrency ownership and shutdown clarity

## Contract and Scope Guard
- When API or behavior changes, confirm the diff still matches the relevant `docs/contracts/*.md`.
- If the docs do not specify the new behavior, record it as unresolved instead of filling the gap with a guess.
- If a fix would require a material architecture change outside the approved scope, stop and ask before doing it.

## Output Contract
- Produce concrete pass/fail notes tied to changed files.
- Surface architecture findings before reviewer gates treat them as downstream surprises.
