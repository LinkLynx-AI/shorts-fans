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

dev AWS media sandbox への representative path を確認します。

```bash
make backend-media-smoke
```

詳細な前提、Terraform output からの env 設定、失敗時の切り分けは [`../docs/infra/dev-media-smoke.md`](../docs/infra/dev-media-smoke.md) を参照してください。

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
- `COGNITO_USER_POOL_CLIENT_ID`
- `MEDIA_JOBS_QUEUE_URL`
- `MEDIA_RAW_BUCKET_NAME`
- `MEDIA_SHORT_PUBLIC_BUCKET_NAME`
- `MEDIA_SHORT_PUBLIC_BASE_URL`
- `MEDIA_MAIN_PRIVATE_BUCKET_NAME`
- `MEDIACONVERT_SERVICE_ROLE_ARN`
- `CREATOR_AVATAR_UPLOAD_BUCKET_NAME`
- `CREATOR_AVATAR_DELIVERY_BUCKET_NAME`
- `CREATOR_AVATAR_BASE_URL`
- `CREATOR_REVIEW_EVIDENCE_BUCKET_NAME`

`MEDIA_JOBS_QUEUE_URL` は旧名 `SQS_QUEUE_URL` を後方互換 alias として受け付けますが、以後は `MEDIA_JOBS_QUEUE_URL` を正とします。

ルート `Makefile` 経由で `make backend-run` を使う場合は、shell から直接 `COGNITO_USER_POOL_CLIENT_ID` を export するのではなく、`BACKEND_COGNITO_USER_POOL_CLIENT_ID` を設定します。Makefile がそれを backend process 用の `COGNITO_USER_POOL_CLIENT_ID` へ引き渡します。`BACKEND_COGNITO_USER_POOL_ID` も従来どおり渡せますが、現状の `cmd/api` fail fast では必須ではありません。

`cmd/api` は `POSTGRES_DSN`、`REDIS_ADDR`、media sandbox 用 env 一式、および creator avatar upload / delivery / review evidence 用 env 一式を必須にします。creator upload endpoint、creator registration avatar upload endpoint、creator registration evidence upload endpoint が常時有効なため、`AWS_REGION`、media bucket / queue / role 設定、avatar bucket / base URL 設定、review evidence bucket 設定が不足している場合は fail fast します。`cmd/worker` は creator avatar env を要求しませんが、media sandbox を有効にして起動する場合は `POSTGRES_DSN` と media queue / bucket / role 設定が必要です。

`COGNITO_USER_POOL_CLIENT_ID` は `SHO-168` の fan auth endpoint wiring で API startup fail fast に組み込みました。`cmd/api` は `POSTGRES_DSN`、`REDIS_ADDR`、media / avatar env 一式に加えて、`AWS_REGION` と `COGNITO_USER_POOL_CLIENT_ID` が欠けている場合も起動しません。`COGNITO_USER_POOL_ID` は dev Cognito sandbox の CLI verification や将来の admin / issuer-based integration では引き続き有用ですが、現状の public Cognito API wiring では必須ではありません。値は `infra/terraform/dev` の `cognito_user_pool_client_id` output から受け取り、必要に応じて `cognito_user_pool_id` も併せて使います。

`cmd/media-smoke` は次を前提にします。

- dev app access policy が local principal に attach 済みであること
- `MEDIA_SHORT_PUBLIC_BASE_URL` が Terraform output の `short_public_base_url` を指していること
- `MEDIA_MAIN_PRIVATE_BUCKET_NAME` が Terraform output の `main_private_bucket_name` を指していること

smoke は `short_public` と `main_private` に一時 probe object を置き、`short` は public URL、`main` は signed URL で取得できることを確認した後に cleanup します。

## Structure

- `cmd/api`: API サーバーの entrypoint
- `cmd/media-smoke`: dev AWS media sandbox の representative path を検証する entrypoint
- `cmd/worker`: media processing worker の entrypoint
- `cmd/migrate`: migration 実行 entrypoint
- `cmd/schema`: migration から人間向け YAML スキーマを生成する entrypoint
- `internal/config`: 環境変数読込と validation
- `internal/dbschema`: 一時 DB を使った schema introspection と YAML 出力
- `internal/httpserver`: Gin router と graceful shutdown
- `internal/mediaconvert`: MediaConvert access check と materialization job 実行
- `internal/postgres`: `pgxpool` 初期化と readiness
- `internal/redis`: Redis client 初期化と readiness
- `internal/s3`: S3 upload / signed URL helper
- `internal/sqs`: SQS 設定、client factory、media wake-up helper
- `db/migrations`: `golang-migrate` 用 SQL
- `db/queries`: `sqlc` 用 query

今後の feature 追加時の backend 構成方針は `../docs/BACKEND_STRUCTURE.md` を参照してください。

## Notes

- この段階では business endpoint はまだ入れません。
- `GET /healthz` は process の生存確認用です。
- `GET /readyz` は PostgreSQL / Redis の疎通確認用です。
