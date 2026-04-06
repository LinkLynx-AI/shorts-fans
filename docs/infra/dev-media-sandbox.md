# dev AWS media sandbox

## 位置づけ

- この文書は `SHO-24 dev 用 AWS media sandbox を Terraform で整備する` の実装内容をまとめるためのものです。
- 対象は cycle1 の local 開発から使う最小の dev media 基盤です。
- `short` の public delivery には `CloudFront` を含めます。
- remote state、GitHub Actions 連携、principal への policy attach 自動化はこの文書の対象外です。

## 依存前提

- AWS アカウント準備と root / MFA / budget / quota の初期 guardrail は `SHO-22` 完了を前提にします。
- media workflow 上の `delivery-ready`、retry `3` 回、`short public` と `main private` の境界は [contracts/mvp-media-workflow-contract.md](../contracts/mvp-media-workflow-contract.md) に従います。
- Terraform CLI `1.9+` と AWS CLI がローカルに入っていることを前提にします。

## この Terraform が作るもの

- private raw upload bucket
  - creator upload の受け口
  - public access は全面 block
  - `14` 日で expire
- short public origin bucket
  - `CloudFront` の S3 origin
  - bucket 自体は private のままにする
- short public CloudFront distribution
  - public short 再生の入口
  - origin access control で short origin bucket だけを読む
- private main delivery bucket
  - paid main 向け
  - direct private S3 + signed URL 前提
- media jobs queue と dead-letter queue
  - `maxReceiveCount = 3`
  - workflow 契約の retry 上限に合わせる
- MediaConvert service role
  - raw bucket 読み取りと short/main bucket 書き込みを許可
- dev app access policy
  - S3 / SQS / MediaConvert job 操作に必要な最小権限をまとめる
  - attach は手動で行う

## 入力変数

- `aws_region`
  - 必須
  - 例: `ap-northeast-1`
- `allowed_app_origins`
  - 任意
  - default は `http://localhost:3000` と `http://127.0.0.1:3000`
  - `main` の private S3 delivery bucket CORS に使います

## 使い方

1. example から tfvars を作ります。

```bash
cd infra/terraform/dev
cp terraform.tfvars.example terraform.tfvars
```

2. `terraform.tfvars` の `aws_region` を dev 用 AWS リージョンに合わせます。

3. 初期化します。

```bash
terraform init
```

4. plan を確認します。

```bash
terraform plan -var-file=terraform.tfvars
```

5. 問題がなければ apply します。

```bash
terraform apply -var-file=terraform.tfvars
```

6. 使い終わったら必要に応じて destroy します。

```bash
terraform destroy -var-file=terraform.tfvars
```

## 後続 task が参照する output

- `raw_bucket_name`
- `short_public_bucket_name`
- `short_public_base_url`
  - public short の CloudFront base URL として使います
- `short_public_cloudfront_distribution_id`
- `short_public_cloudfront_distribution_arn`
- `short_public_cloudfront_domain_name`
- `main_private_bucket_name`
- `media_jobs_queue_url`
- `media_jobs_queue_arn`
- `media_jobs_dlq_url`
- `mediaconvert_service_role_arn`
- `media_app_access_policy_arn`

特に `SHO-25` では `AWS_REGION` と `media_jobs_queue_url`、bucket 名、MediaConvert role ARN が接続前提になります。

backend から読む env 名は次に揃えます。

| Terraform output | backend env |
| --- | --- |
| `aws_region` | `AWS_REGION` |
| `media_jobs_queue_url` | `MEDIA_JOBS_QUEUE_URL` |
| `raw_bucket_name` | `MEDIA_RAW_BUCKET_NAME` |
| `short_public_bucket_name` | `MEDIA_SHORT_PUBLIC_BUCKET_NAME` |
| `short_public_base_url` | `MEDIA_SHORT_PUBLIC_BASE_URL` |
| `main_private_bucket_name` | `MEDIA_MAIN_PRIVATE_BUCKET_NAME` |
| `mediaconvert_service_role_arn` | `MEDIACONVERT_SERVICE_ROLE_ARN` |

`SQS_QUEUE_URL` は旧 skeleton との互換 alias としてのみ残し、新しい backend 実装では `MEDIA_JOBS_QUEUE_URL` を使います。

## Guardrail

- `short` の public delivery は `CloudFront + private S3 origin` に固定します。
- `main` は引き続き direct private S3 + signed URL 前提です。
- raw/main bucket は public access block を有効化します。
- short origin bucket も private のままにし、bucket policy は対象の CloudFront distribution だけを許可します。
- すべての bucket policy で HTTPS 以外のアクセスを deny します。
- すべての bucket で SSE-S3 を使います。
- queue は SQS managed SSE を有効化します。
- raw object は lifecycle で自動削除し、不要な蓄積コストを抑えます。

## 手動依存と未対応事項

- dev app access policy の principal attach は手動です。
  - どの IAM user / role に attach するかは `SHO-22` の access 運用に従います。
- remote state backend は未対応です。
  - この root module は local state 前提です。
- MediaConvert custom queue、job template、preset、notification 設定は未対応です。
  - cycle1 は default queue 前提に留めます。
- `short` の custom domain、TLS 証明書、WAF、invalidation 運用は未対応です。
- `main` の CloudFront 化は未対応です。
- browser から raw bucket へ直接 upload する CORS / presigned upload flow は未定義です。
  - 必要なら後続 task で追加します。
