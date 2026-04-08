---
name: docs-sync
description: Check whether the current code change requires repository documentation updates, index updates, or explicit no-change notes, while respecting the `docs/ssot/` boundary. Use during verification or handoff whenever behavior, contracts, commands, or docs files changed.
---

# docs-sync

## Goal
- Keep repository documentation aligned with the delivered code.
- Avoid silent drift between behavior, contracts, and tracked docs.

## Input Contract
- Current diff or touched-file set.
- Relevant product, contract, and implementation docs.

## Operating Rules
- Never edit `docs/ssot/` directly from this repo.
- If a doc is added or renamed under `docs/`, update both `docs/README.md` and the root `AGENTS.md` index.
- If public behavior, contracts, or operational commands changed, either update the relevant docs or record why no doc change is needed.
- If the current run is intentionally code-only, record the explicit no-doc-change rationale.

## Procedure
1. Inspect the diff for behavior, contract, command, or documentation impact.
2. Update the relevant tracked docs when the change alters user-visible behavior, API contracts, setup, validation flow, or repo conventions.
3. Update `docs/README.md` and root `AGENTS.md` when the doc set itself changes.
4. If no doc update is required, record the rationale in the run log or verification evidence.

## Output Contract
- Documentation impact is resolved, not implicit.
- Index files stay in sync with the actual tracked docs tree.
