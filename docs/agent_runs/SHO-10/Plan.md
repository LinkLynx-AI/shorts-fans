# Plan.md (Milestones + validations)

## Rules
- Stop-and-fix: if validation fails, repair it before moving to the next step.

## Milestones
### M1: Fix document shape and contract boundaries
- Acceptance criteria:
  - [ ] contract scope、non-goals、canonical SSOT references が固定されている。
  - [ ] 実装に委ねない判断と、あえて deferred にする判断が分かれている。
- Validation:
  - `rg -n "## Goals|## Non-goals|## Canonical Sources|## Deferred Decisions" docs/contracts/mvp-core-domain-contract.md`

### M2: Write the MVP core domain contract
- Acceptance criteria:
  - [ ] vocabulary、relationships、access boundaries、state transitions、publish/unlock preconditions が文書化されている。
  - [ ] `creator capability`、`creator profile`、`main unlock`、`submission package` の境界が一意に読める。
- Validation:
  - `rg -n "## Domain Vocabulary|## Relationship And Ownership Contract|## Access Boundary Contract|## State Transition Contract|## Publish And Unlock Preconditions" docs/contracts/mvp-core-domain-contract.md`

### M3: Sync indexes and run evidence
- Acceptance criteria:
  - [ ] `docs/README.md` と `AGENTS.md` に新規文書領域が反映されている。
  - [ ] run evidence が実際の判断と validation 状況を反映している。
- Validation:
  - `git diff -- docs/contracts/mvp-core-domain-contract.md docs/README.md AGENTS.md docs/agent_runs/SHO-10`
