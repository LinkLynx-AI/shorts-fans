# Prompt.md (Spec / Source of truth)

## Goals
- `docs/contracts/mvp-core-domain-contract.md` と `docs/ssot/` に沿って MVP core entity の初期 migration を作る。
- `user` を root identity にしたまま、`creator capability`、creator draft/public profile、`media`、`short`、`main`、`unlock`、fan-side follow/pin/progress を DB 上で表現する。
- `1 canonical main : 複数 short`、public `short`、locked/purchased `main`、private fan hub state を永続化で表せるようにする。
- high-risk boundary は backend だけでなく DB でも壊れないようにする。

## Non-goals
- `submission package` 専用 schema の導入。
- `subscription` や analytics schema の導入。
- query/repository/API 実装。
- SHO-23 で決める media workflow 詳細の先取り。
- docs 未確定な `consumer settings` の具体列定義。

## Deliverables
- `backend/db/migrations/000002_mvp_core_entities.up.sql`
- `backend/db/migrations/000002_mvp_core_entities.down.sql`
- reviewer/validation evidence を含む `docs/agent_runs/SHO-11/*`

## Done when
- [ ] empty DB から `up` で core schema を作れる
- [ ] `down` で `000002` を戻せる
- [ ] `1 main : n short` と `main unlock` が表現できる
- [ ] approval/public/purchase boundary の主要不変条件が DB でも守られる
- [ ] backend validation と review gate の結果を残す

## Constraints
- Perf: MVP 主線に不要な index や抽象化を入れすぎない
- Security: ownership access と purchase access を混ぜない
- Compatibility: `sqlc` と `golang-migrate` がそのまま扱える PostgreSQL schema にする
