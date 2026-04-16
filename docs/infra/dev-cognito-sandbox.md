# dev Cognito sandbox

## 位置づけ

- この文書は `SHO-198 Cognito を実際に動かすための Terraform / infra 設定を整備する` の成果物です。
- 対象は `infra/terraform/dev` に追加した fan auth 用 Cognito slice です。
- `docs/contracts/fan-auth-api-contract.md` が固定した `email + password` / custom modal 前提を、AWS コンソール手作業なしで再現できるようにすることを目的にします。
- media / avatar sandbox と同じ Terraform root を使いますが、mail verification / password reset / auth runtime handoff だけをこの文書で扱います。

## 固定した前提

- Cognito Hosted UI は primary flow にしません。
- App Client は `generate_secret = false` の secretless 構成です。
- mail delivery は `COGNITO_DEFAULT` を使い、SES identity や `source_arn` はこの issue では作りません。
- social login / MFA / passkey / custom attribute は今回の対象外です。
- internal user / viewer / session の canonical state は引き続き app backend が持ちます。

## この Terraform が作るもの

- fan auth 用 Cognito User Pool
  - sign-in attribute は `email`
  - email auto verify を有効化
  - self-service sign-up を許可
  - case-insensitive username handling
  - account recovery は `verified_email` のみ
  - email confirmation は code 方式
  - deletion protection は dev destroy を妨げない `INACTIVE`
- fan auth 用 Cognito App Client
  - backend custom UI 連携向け
  - client secret なし
  - `prevent_user_existence_errors = ENABLED`
  - `ALLOW_USER_PASSWORD_AUTH` のみ許可

## この issue で作らないもの

- Hosted UI domain
- OAuth callback / logout URL
- client secret の保管先
- SES identity / sender address / configuration set
- Cognito trigger Lambda
- social provider 設定

## 使い方

1. 既存の tfvars を用意します。

```bash
cd infra/terraform/dev
cp terraform.tfvars.example terraform.tfvars
```

2. `terraform.tfvars` の `aws_region` を対象リージョンに合わせます。

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

6. 変更を戻したい場合は定義を差し戻して再 apply します。環境ごと消す場合だけ destroy を使います。

```bash
terraform apply -var-file=terraform.tfvars
terraform destroy -var-file=terraform.tfvars
```

## 後続 task が参照する output

- `aws_region`
- `cognito_user_pool_id`
- `cognito_user_pool_arn`
- `cognito_user_pool_client_id`
- `cognito_user_pool_issuer_url`

backend fan auth 実装が読む env 名は次に揃えます。

| Terraform output | backend env |
| --- | --- |
| `aws_region` | `AWS_REGION` |
| `cognito_user_pool_id` | `COGNITO_USER_POOL_ID` |
| `cognito_user_pool_client_id` | `COGNITO_USER_POOL_CLIENT_ID` |

`COGNITO_CLIENT_SECRET` は存在しません。CLI でも backend でも `SECRET_HASH` は不要です。

## mail prerequisite と制約

- この slice は `COGNITO_DEFAULT` mail を前提にするため、SES identity の apply は不要です。
- 一方で、対象 AWS アカウントで Cognito default mail が利用可能であることは前提です。
- verification code / password reset code の送信元や branding を制御したい場合は、後続 issue で SES 連携を追加します。
- mail delivery の成否確認は AWS コンソール前提にせず、使い捨て email と CLI verification で行います。

## CLI verification

まず repo root へ戻り、Terraform output を shell に取り込みます。

```bash
cd /path/to/shorts-fans

export BACKEND_AWS_REGION="$(terraform -chdir=infra/terraform/dev output -raw aws_region)"
export BACKEND_COGNITO_USER_POOL_ID="$(terraform -chdir=infra/terraform/dev output -raw cognito_user_pool_id)"
export BACKEND_COGNITO_USER_POOL_CLIENT_ID="$(terraform -chdir=infra/terraform/dev output -raw cognito_user_pool_client_id)"
export BACKEND_COGNITO_ISSUER_URL="$(terraform -chdir=infra/terraform/dev output -raw cognito_user_pool_issuer_url)"
```

verification 用の一時値を用意します。

```bash
export TEST_EMAIL="replace-with-disposable-email@example.com"
export TEST_PASSWORD="TempPass123!"
export NEW_TEST_PASSWORD="TempPass456!"
```

1. sign up を開始します。

```bash
aws cognito-idp sign-up \
  --region "$BACKEND_AWS_REGION" \
  --client-id "$BACKEND_COGNITO_USER_POOL_CLIENT_ID" \
  --username "$TEST_EMAIL" \
  --password "$TEST_PASSWORD" \
  --user-attributes Name=email,Value="$TEST_EMAIL"
```

2. mailbox で confirmation code を受け取り、confirm します。

```bash
export SIGN_UP_CODE="replace-with-sign-up-code"

aws cognito-idp confirm-sign-up \
  --region "$BACKEND_AWS_REGION" \
  --client-id "$BACKEND_COGNITO_USER_POOL_CLIENT_ID" \
  --username "$TEST_EMAIL" \
  --confirmation-code "$SIGN_UP_CODE"
```

3. backend 想定の auth flow を確認します。

```bash
aws cognito-idp initiate-auth \
  --region "$BACKEND_AWS_REGION" \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id "$BACKEND_COGNITO_USER_POOL_CLIENT_ID" \
  --auth-parameters USERNAME="$TEST_EMAIL",PASSWORD="$TEST_PASSWORD"
```

4. password reset を開始します。

```bash
aws cognito-idp forgot-password \
  --region "$BACKEND_AWS_REGION" \
  --client-id "$BACKEND_COGNITO_USER_POOL_CLIENT_ID" \
  --username "$TEST_EMAIL"
```

5. mailbox で reset code を受け取り、new password を確定します。

```bash
export PASSWORD_RESET_CODE="replace-with-reset-code"

aws cognito-idp confirm-forgot-password \
  --region "$BACKEND_AWS_REGION" \
  --client-id "$BACKEND_COGNITO_USER_POOL_CLIENT_ID" \
  --username "$TEST_EMAIL" \
  --confirmation-code "$PASSWORD_RESET_CODE" \
  --password "$NEW_TEST_PASSWORD"
```

6. reset 後の sign-in を確認します。

```bash
aws cognito-idp initiate-auth \
  --region "$BACKEND_AWS_REGION" \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id "$BACKEND_COGNITO_USER_POOL_CLIENT_ID" \
  --auth-parameters USERNAME="$TEST_EMAIL",PASSWORD="$NEW_TEST_PASSWORD"
```

7. 検証用 user を削除します。

```bash
aws cognito-idp admin-delete-user \
  --region "$BACKEND_AWS_REGION" \
  --user-pool-id "$BACKEND_COGNITO_USER_POOL_ID" \
  --username "$TEST_EMAIL"
```

## rollback / destroy

- Cognito slice の Terraform 定義を変えたあとで元に戻すときは、差分を戻して `terraform apply -var-file=terraform.tfvars` を再実行します。
- 開発環境全体を破棄するときだけ `terraform destroy -var-file=terraform.tfvars` を使います。
- destroy すると media / avatar / Cognito slice が同じ root からまとめて消えるため、必要な output を事前に控えておきます。

## 後続 task への handoff

- `SHO-168` の current backend runtime は `AWS_REGION` と `COGNITO_USER_POOL_CLIENT_ID` を使って public Cognito API へ接続します。
- `COGNITO_USER_POOL_ID` は CLI verification の `admin-delete-user` や将来の issuer / admin API integration では引き続き使えるため、Terraform output として維持します。
- App Client が secretless なので、backend は `USER_PASSWORD_AUTH` / sign-up / password reset 系の public Cognito API をそのまま扱う前提です。
- issuer/JWKS を使う token validation を後続で入れる場合は `cognito_user_pool_issuer_url` を起点にします。
