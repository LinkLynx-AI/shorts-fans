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

`sqlc` 生成を実行します。

```bash
make backend-generate
```

## Quality

```bash
make backend-fmt
make backend-test
make backend-vet
```

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
- `internal/config`: 環境変数読込と validation
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
