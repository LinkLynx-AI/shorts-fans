# dev media smoke runbook

## 位置づけ

- この文書は `SHO-26 media pipeline の疎通確認と運用メモを整える` の成果物です。
- 対象は `backend/cmd/media-smoke` による cycle1 の representative path 確認です。
- `import -> transcode -> linkage` の full workflow を end-to-end で保証するものではありません。

## この smoke が確認すること

`make backend-media-smoke` は次の順に representative path を確認します。

1. `MEDIA_JOBS_QUEUE_URL` に対する `GetQueueAttributes`
   - local principal が media jobs queue を参照できるかを確認します。
2. `MediaConvert ListQueues`
   - dev sandbox の region で MediaConvert API に到達できるかを確認します。
3. `MEDIA_SHORT_PUBLIC_BUCKET_NAME` への probe object 配置
   - `codex/media-smoke/<uuid>/short-probe.m3u8` を配置します。
4. `MEDIA_SHORT_PUBLIC_BASE_URL` 経由の public short 取得
   - short public delivery が `200 OK` を返し、配置した内容と一致するか確認します。
5. `MEDIA_MAIN_PRIVATE_BUCKET_NAME` への probe object 配置
   - `codex/media-smoke/<uuid>/main-probe.m3u8` を配置します。
6. signed URL 経由の private main 取得
   - `main` が private S3 + signed URL 前提で取得できるか確認します。
7. cleanup
   - probe object を削除します。prefix は `codex/media-smoke/` です。

実装の一次ソースは [backend/cmd/media-smoke/main.go](../../backend/cmd/media-smoke/main.go) と [backend/internal/media/probe.go](../../backend/internal/media/probe.go) です。

## この smoke が確認しないこと

- raw upload bucket への browser upload
- MediaConvert job の submit / complete / retry
- DB 更新、worker dequeue、`media_asset.processing_state` の遷移
- `short-main linkage` や review/publish 判定
- creator / fan UI からの playback

つまり、cycle1 の「AWS media sandbox に local backend から到達できるか」を見る smoke であって、「media workflow 実装が完成したか」を見るものではありません。

## 事前条件

- `SHO-24 dev 用 AWS media sandbox を Terraform で整備する` の apply が完了していること
- `SHO-25 local backend から dev AWS media sandbox に接続する` の env contract に従って Terraform output を backend env へ渡せること
- dev app access policy が local principal に手動 attach 済みであること
- local で AWS 認証済みであること
  - まず `aws sts get-caller-identity` が成功する状態にします
- `MEDIA_SHORT_PUBLIC_BASE_URL` が Terraform output の `short_public_base_url` を指していること
- `MEDIA_MAIN_PRIVATE_BUCKET_NAME` が Terraform output の `main_private_bucket_name` を指していること

必要な env 名は [backend/internal/config/config.go](../../backend/internal/config/config.go) の `ValidateMediaSmoke` に揃っています。

## セットアップ手順

Terraform state から output を読み、shell に export してから smoke を流します。

```bash
export BACKEND_AWS_REGION="$(terraform -chdir=infra/terraform/dev output -raw aws_region)"
export BACKEND_MEDIA_JOBS_QUEUE_URL="$(terraform -chdir=infra/terraform/dev output -raw media_jobs_queue_url)"
export BACKEND_MEDIA_RAW_BUCKET_NAME="$(terraform -chdir=infra/terraform/dev output -raw raw_bucket_name)"
export BACKEND_MEDIA_SHORT_PUBLIC_BUCKET_NAME="$(terraform -chdir=infra/terraform/dev output -raw short_public_bucket_name)"
export BACKEND_MEDIA_SHORT_PUBLIC_BASE_URL="$(terraform -chdir=infra/terraform/dev output -raw short_public_base_url)"
export BACKEND_MEDIA_MAIN_PRIVATE_BUCKET_NAME="$(terraform -chdir=infra/terraform/dev output -raw main_private_bucket_name)"
export BACKEND_MEDIACONVERT_SERVICE_ROLE_ARN="$(terraform -chdir=infra/terraform/dev output -raw mediaconvert_service_role_arn)"
```

ローカルに state がない場合は、Terraform output の値を安全な手段で取得して同じ環境変数に投入します。repo に固定値は残しません。

## 実行手順

1. AWS 認証状態を確認します。

```bash
aws sts get-caller-identity
```

2. smoke を実行します。

```bash
make backend-media-smoke
```

3. 成功時は `media smoke succeeded` とともに次の情報が出ます。
   - `queue_arn`
   - `mediaconvert_queue`
   - `short_public_url`
   - `short_object_key`
   - `main_object_key`
   - `main_signed_url_host`

`short_object_key` と `main_object_key` は cleanup 失敗時の手動削除にも使います。

## 失敗時の切り分け

### 設定不足で即失敗する

- 例: `missing required environment variables: ...`
- 原因:
  - Terraform output の読み込み漏れ
  - `MEDIA_JOBS_QUEUE_URL` ではなく古い alias だけ見ている
- 対応:
  - export をやり直す
  - `env | rg '^(BACKEND_AWS_REGION|BACKEND_MEDIA_)'` で投入済み値を確認する

### AWS 認証で失敗する

- 例: `Unable to locate credentials`
- 原因:
  - local で AWS 認証していない
  - profile / role 切り替えが終わっていない
- 対応:
  - `aws sts get-caller-identity` が通る状態にしてから再実行する

### queue / bucket / MediaConvert が `AccessDenied` になる

- 原因:
  - `media_app_access_policy_arn` が local principal に attach されていない
  - attach 先を誤っている
- 対応:
  - `docs/infra/dev-media-sandbox.md` の手動依存に従って principal attach を見直す
  - `GetQueueAttributes`、S3 `PutObject/DeleteObject`、MediaConvert `ListQueues` の権限が揃っているか確認する

### short public fetch が `403` / `404` になる

- 原因:
  - `MEDIA_SHORT_PUBLIC_BASE_URL` が別 distribution を指している
  - CloudFront distribution / OAC / bucket policy の反映前
  - short origin bucket 以外へ object を置いている
- 対応:
  - `short_public_base_url` を取り直す
  - CloudFront の deploy 完了を待って再実行する
  - bucket 名と `short_object_key` を照合する

### main signed URL fetch が `403` になる

- 原因:
  - `MEDIA_MAIN_PRIVATE_BUCKET_NAME` の指定誤り
  - local 時刻ずれや署名生成条件の不整合
  - object 配置までは通っているが取得権限に問題がある
- 対応:
  - bucket 名を再確認する
  - local clock を同期した上で再実行する
  - `main_object_key` が bucket に存在するか確認する

### cleanup に失敗して一時 object が残る

- 原因:
  - 実行途中で中断した
  - delete 権限がない
- 対応:
  - `codex/media-smoke/` prefix を手動で削除する
  - 削除後に再実行する

## 手動回復ポイント

- dev app access policy の principal attach は自動化していません。
- CloudFront の反映待ちは手動です。
- cleanup に失敗した probe object の削除は手動です。
- local でどの AWS profile / role を使うかは repo 外の access 運用に従います。

## quota / cost 注意点

- この smoke は MediaConvert job を投げません。
  - `ListQueues` だけなので、transcode コストは発生しません。
- 実際に発生するのは小さい probe object に対する S3 `PutObject/GetObject/DeleteObject`、CloudFront GET、SQS `GetQueueAttributes` です。
- HTTP fetch は retry を含めても最大 `5` 回です。
- raw bucket は使わないため、大きい media transfer を伴いません。

## cycle1 で未対応の部分

- `raw upload -> processing -> ready` の state 遷移そのもの
- actual MediaConvert job template / preset / queue 運用
- worker による dequeue と retry 制御
- `short public publishable` / `main unlockable` 判定
- private `main` の CloudFront 化

この文書は「次の実装者が dev sandbox の入口で迷わない」ことを目標にし、full production runbook は後続 cycle に残します。
