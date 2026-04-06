# backend

`backend/` は `Go + Gin + sqlc + pgx` を前提にした API / worker 基盤です。

## Requirements

- Go `1.24`
- Docker / Docker Compose

## Development

ローカル依存を起動します。

```bash
make backend-dev-up
```

API サーバーを起動します。

```bash
make backend-run
```

ローカル開発用の fixed mock data を投入します。

```bash
make backend-dev-seed
```

worker 骨格を起動します。

```bash
make backend-worker
```

停止するときは次を使います。

```bash
make backend-dev-down
```

## Database

migration を 1 つ進めます。

```bash
make backend-migrate-up
```

migration を 1 つ戻します。

```bash
make backend-migrate-down
```

dev seed を再投入するときは次を使います。

```bash
make backend-dev-seed
```

この seed は `creator-capable user 1人 + fan-only user 1人 + 公開済み short/main + follow/unlock/pin` を固定 UUID で idempotent に投入します。ローカル DB を作り直した直後の復旧や、手元確認の初期データとして使う想定です。

`sqlc` 生成を実行します。

```bash
make backend-generate
```

人間向けの YAML スキーマを生成します。

```bash
make backend-schema
./scripts/generate-db-schema.sh
```

このコマンドは `POSTGRES_DSN` が指す PostgreSQL サーバー上に一時 DB を作成し、そこへ migration を適用して `db/schema.generated.yaml` を生成します。実行ユーザーには一時 DB を作成・削除できる権限が必要です。

新しい migration を追加するときは、`backend/db/migrations/` に連番で `*.up.sql` と `*.down.sql` を対で作ります。
例:

```text
backend/db/migrations/000003_some_change.up.sql
backend/db/migrations/000003_some_change.down.sql
```

基本ルールは次のとおりです。

- `up` は forward-only で読めるように書き、`down` はその migration だけを 1 step 戻せる内容にする
- 既存 migration は原則書き換えず、新しい番号の migration を積む
- schema を変えたら `make backend-generate` も実行して、`sqlc` 生成が壊れていないことを確認する
- schema を変えたら `make backend-schema` で `db/schema.generated.yaml` も更新する
- DB で管理すべき不変条件は `FK`、`UNIQUE`、`CHECK`、必要なら trigger / view で表現する

新しい migration を作った後の確認手順は次を基準にします。

1. ローカル依存を起動する

```bash
make backend-dev-up
```

2. 現在の migration version を確認する

```bash
cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/migrate version
```

3. migration を適用する

```bash
make backend-migrate-up
```

4. 必要なら 1 step 戻す

```bash
make backend-migrate-down
```

5. 変更が大きいときは `up -> down -> up` まで確認する

```bash
make backend-migrate-up
make backend-migrate-down
make backend-migrate-up
```

6. backend の最低限検証を流す

```bash
make backend-schema
make backend-test
make backend-vet
make backend-generate
```

直接 `cmd/migrate` を使う場合は `POSTGRES_DSN` を渡して実行します。

```bash
cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/migrate up
cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/migrate down
cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/migrate version
cd backend && POSTGRES_DSN='postgres://shorts_fans:shorts_fans@localhost:5432/shorts_fans?sslmode=disable' go run ./cmd/schema
```

## Quality

```bash
make backend-fmt
make backend-test
make backend-coverage-check
make backend-vet
```

任意のしきい値で判定したいときは `BACKEND_COVERAGE_MIN` を渡します。

```bash
make backend-coverage-check BACKEND_COVERAGE_MIN=30
```

coverage check は既定で `cmd/*`、generated code の `internal/postgres/sqlc`、開発ツール用の `internal/dbschema` を除外します。entrypoint 配線や生成物、runtime request path 外の schema export ツールではなく、runtime の手書きロジックをしきい値の対象にするためです。

`BACKEND_COVERAGE_PROFILE` を指定すると、`go tool cover -html` などで再利用できる coverage profile を保存できます。

## Environment

通常はルート `Makefile` の既定値でローカル起動できます。外部環境に向ける場合は次の環境変数を上書きします。

- `APP_ENV`
- `API_ADDR`
- `POSTGRES_DSN`
- `REDIS_ADDR`
- `AWS_REGION`
- `SQS_QUEUE_URL`

`cmd/api` は `POSTGRES_DSN` と `REDIS_ADDR` を必須にします。`cmd/worker` は SQS 設定が未投入でも骨格起動できます。

## Structure

- `cmd/api`: API サーバーの entrypoint
- `cmd/worker`: worker 骨格の entrypoint
- `cmd/migrate`: migration 実行 entrypoint
- `cmd/schema`: migration から人間向け YAML スキーマを生成する entrypoint
- `internal/config`: 環境変数読込と validation
- `internal/dbschema`: 一時 DB を使った schema introspection と YAML 出力
- `internal/httpserver`: Gin router と graceful shutdown
- `internal/postgres`: `pgxpool` 初期化と readiness
- `internal/redis`: Redis client 初期化と readiness
- `internal/sqs`: SQS 設定と client factory の骨格
- `db/migrations`: `golang-migrate` 用 SQL
- `db/queries`: `sqlc` 用 query

今後の feature 追加時の backend 構成方針は `../docs/BACKEND_STRUCTURE.md` を参照してください。

## Notes

- この段階では business endpoint はまだ入れません。
- `GET /healthz` は process の生存確認用です。
- `GET /readyz` は PostgreSQL / Redis の疎通確認用です。
