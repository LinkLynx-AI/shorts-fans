# Prompt.md (Spec / Source of truth)

## Goals
- MVP core 永続化作業向けの実装用ドメイン契約を 1 文書追加する。
- `docs/ssot/` は承認済み product source として変更しない。
- `SHO-11` 以降の DB 作業で必要になる identity、ownership、linkage、access、review state の曖昧さをなくす。

## Non-goals
- DB schema 設計
- SQL や `sqlc` 契約の設計
- API 契約の設計
- review tooling UI の詳細
- `subscription` 設計

## Deliverables
- `docs/contracts/mvp-core-domain-contract.md`
- Index updates in `docs/README.md` and `AGENTS.md`
- Run evidence files under `docs/agent_runs/SHO-10/`

## Done when
- [ ] 1 つの contract 文書に vocabulary、relationships、access boundary、state transitions、preconditions が揃っている。
- [ ] SSOT を変更せず、参照元と実装上の delta note が明示されている。
- [ ] 索引が同期されている。

## Constraints
- Perf: docs-only なので N/A
- Security: 承認済み SSOT を超えて access behavior を広げない
- Compatibility: `SHO-11+` の実装者が追加 product 判断なしで読めること
