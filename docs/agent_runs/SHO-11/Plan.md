# Plan.md (Milestones + validations)

## Rules
- Stop-and-fix: validation が落ちたら次へ進まず修正する。

## Milestones
### M1: spec と schema 方針の固定
- Acceptance criteria:
  - [ ] table 境界と state 表現が SSOT と矛盾しない
  - [ ] media と submission package の scope boundary が固定されている
- Validation:
  - `rg` / `sed` で docs と既存 backend 構成を確認

### M2: migration と sqlc schema 設定の実装
- Acceptance criteria:
  - [ ] `000002` の up/down が追加されている
  - [ ] `backend/sqlc.yaml` が新 migration を読む
- Validation:
  - `git diff -- backend/db/migrations backend/sqlc.yaml`

### M3: backend 検証
- Acceptance criteria:
  - [ ] `go test` と `go vet` が通る
  - [ ] migration の apply / rollback / reapply が通る
  - [ ] `sqlc generate` が必要なら成功する
- Validation:
  - `GOCACHE=/tmp/go-build-sho11 go test ./...`
  - `GOCACHE=/tmp/go-build-sho11 go vet ./...`
  - local postgres 上で `go run ./cmd/migrate up/down/up`
  - `GOCACHE=/tmp/go-build-sho11 go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate`
