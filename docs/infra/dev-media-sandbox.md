# dev AWS media sandbox

## 位置づけ

- この文書は `SHO-24 dev 用 AWS media sandbox を Terraform で整備する` の実装内容をまとめるためのものです。
- 対象は cycle1 の local 開発から使う最小の dev media 基盤です。
- `CloudFront`、remote state、GitHub Actions 連携、principal への policy attach 自動化はこの文書の対象外です。

## 依存前提

- AWS アカウント準備と root / MFA / budget / quota の初期 guardrail は `SHO-22` 完了を前提にします。
- media workflow 上の `delivery-ready`、retry `3` 回、`short public` と `main private` の境界は [contracts/mvp-media-workflow-contract.md](../contracts/mvp-media-workflow-contract.md) に従います。
- Terraform CLI `1.9+` と AWS CLI がローカルに入っていることを前提にします。

## この Terraform が作るもの

- private raw upload bucket
  - creator upload の受け口
  - public access は全面 block
  - `14` 日で expire
- public short delivery bucket
  - dev では direct S3 public object を許可
  - 匿名 `GetObject` だけを bucket policy で許可
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
  - short/main の delivery bucket CORS に使います

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
- `main_private_bucket_name`
- `media_jobs_queue_url`
- `media_jobs_queue_arn`
- `media_jobs_dlq_url`
- `mediaconvert_service_role_arn`
- `media_app_access_policy_arn`

特に `SHO-25` では `AWS_REGION` と `media_jobs_queue_url`、bucket 名、MediaConvert role ARN が接続前提になります。

## Guardrail

- `CloudFront` をここでは作りません。
  - dev は direct S3 delivery に留めます。
  - workflow 契約でも `CloudFront` は必須ではありません。
- raw/main bucket は public access block を有効化します。
- short bucket も ACL では public にせず、bucket policy で `GetObject` だけを匿名公開します。
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
- browser から raw bucket へ直接 upload する CORS / presigned upload flow は未定義です。
  - 必要なら後続 task で追加します。
