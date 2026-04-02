# Plan.md (Milestones + validations)

## Rules
- Stop-and-fix: validation が落ちたら次へ進まず修正する。

## Milestones
### M1: spec と schema 方針の固定
- Acceptance criteria:
  - [x] table 境界と state 表現が SSOT と矛盾しない
  - [x] media と submission package の scope boundary が固定されている
  - [x] DB で強制する境界と backend に残す workflow の線引きが決まっている
- Validation:
  - `rg` / `sed` で docs と既存 backend 構成を確認

### M2: migration と DB guard の実装
- Acceptance criteria:
  - [x] `000002` の up/down が current schema に追随している
  - [x] approval/public/purchase boundary を守る trigger/view が追加されている
  - [x] creator draft/public profile の境界が table で分離されている
- Validation:
  - `git diff -- backend/db/migrations docs/agent_runs/SHO-11`

### M3: backend 検証
- Acceptance criteria:
  - [x] `go test` と `go vet` が通る
  - [x] reviewer gate で blocking finding が解消されている
  - [x] migration の apply / rollback / reapply は実行可否が記録されている
- Validation:
  - `GOCACHE=/tmp/go-build-sho11 go test ./...`
  - `GOCACHE=/tmp/go-build-sho11 go vet ./...`
  - `git diff --check`
  - `reviewer` / `reviewer_ui_guard`
