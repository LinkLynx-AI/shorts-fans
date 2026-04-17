# dev Cognito sandbox

## 位置づけ

- この文書は `SHO-198 Cognito を実際に動かすための Terraform / infra 設定を整備する` の成果物です。
- 対象は `infra/terraform/dev` に追加した fan auth 用 Cognito slice です。
- `docs/contracts/fan-auth-api-contract.md` が固定した `email + password` / custom modal 前提を、AWS コンソール手作業なしで再現できるようにすることを目的にします。
- media / avatar sandbox と同じ Terraform root を使いますが、mail verification / password reset / auth runtime handoff だけをこの文書で扱います。

## 固定した前提

- Cognito Hosted UI は primary flow にしません。
- App Client は `generate_secret = false` の secretless 構成です。
- mail delivery は Cognito user pool の `DEVELOPER + SES` を最終形とし、SES sender identity verify 完了までは `COGNITO_DEFAULT` のまま維持できる staged rollout にします。
- 最小 branding は `From: shortsfans <no-reply@...>` に限定し、mail subject / body の custom template は今回入れません。
- social login / MFA / passkey / custom attribute は今回の対象外です。
- internal user / viewer / session の canonical state は引き続き app backend が持ちます。

## この Terraform が作るもの

- fan auth mail 用 SES sender identity
  - `cognito_email_from_address` に指定した email address を SES identity として作成
  - SES verification mail の承認完了後に Cognito `DEVELOPER` mode へ切り替え可能
- fan auth 用 Cognito User Pool
  - sign-in attribute は `email`
  - email auto verify を有効化
  - self-service sign-up を許可
  - case-insensitive username handling
  - account recovery は `verified_email` のみ
  - email confirmation は code 方式
  - deletion protection は dev destroy を妨げない `INACTIVE`
  - `cognito_use_ses_developer_email = true` のとき、`shortsfans <...>` と SES `source_arn` を使って送信
- fan auth 用 Cognito App Client
  - backend custom UI 連携向け
  - client secret なし
  - `prevent_user_existence_errors = ENABLED`
  - `ALLOW_USER_PASSWORD_AUTH` のみ許可

## この issue で作らないもの

- Hosted UI domain
- OAuth callback / logout URL
- client secret の保管先
- SES configuration set / dedicated IP / domain DKIM automation
- Cognito trigger Lambda
- social provider 設定

## 使い方

1. 既存の tfvars を用意します。

```bash
cd infra/terraform/dev
cp terraform.tfvars.example terraform.tfvars
```

2. `terraform.tfvars` に少なくとも次を設定します。

```hcl
aws_region = "ap-northeast-1"

cognito_email_from_address      = "no-reply@example.com"
cognito_use_ses_developer_email = false
```

- `cognito_email_from_address` は最終的に `shortsfans <...>` の `...` 部分として使います。
- `cognito_use_ses_developer_email` は sender identity verify 前は `false` のままにします。

3. 初期化します。

```bash
terraform init
```

4. まず sender identity 作成用の plan を確認します。

```bash
terraform plan -var-file=terraform.tfvars
```

5. 問題がなければ apply して SES sender identity を作成します。

```bash
terraform apply -var-file=terraform.tfvars
```

6. `cognito_email_from_address` に届く SES verification mail の承認リンクを開き、sender identity を verify します。

7. SES console または CLI で sender identity の `VerificationStatus = SUCCESS` を確認します。

```bash
aws sesv2 get-email-identity \
  --region "$(terraform output -raw aws_region)" \
  --email-identity "no-reply@example.com"
```

- `--email-identity` には `cognito_email_from_address` に設定した実値を使います。

8. verify 完了を確認したら `terraform.tfvars` の `cognito_use_ses_developer_email = true` へ変更し、Cognito user pool を `DEVELOPER` mode に切り替えます。

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

9. 変更を戻したい場合は `cognito_use_ses_developer_email = false` に戻して再 apply します。環境ごと消す場合だけ destroy を使います。

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
- `cognito_email_sending_account`
- `cognito_email_from_address`
- `cognito_ses_email_identity_arn`
- `cognito_ses_email_identity_verification_status`

`cognito_ses_email_identity_verification_status` は Terraform state 上の最終 refresh 値です。SES verification mail の承認直後は自動で更新されないため、再度 `terraform plan` / `terraform apply` か state refresh を行ってから確認します。

backend fan auth 実装が読む env 名は次に揃えます。

| Terraform output | backend env |
| --- | --- |
| `aws_region` | `AWS_REGION` |
| `cognito_user_pool_id` | `COGNITO_USER_POOL_ID` |
| `cognito_user_pool_client_id` | `COGNITO_USER_POOL_CLIENT_ID` |

`COGNITO_CLIENT_SECRET` は存在しません。CLI でも backend でも `SECRET_HASH` は不要です。

## mail prerequisite と制約

- `cognito_use_ses_developer_email = true` の apply には、`cognito_email_from_address` の SES verify 完了が前提です。
- Terraform でも `cognito_use_ses_developer_email = true` の apply 前に `VerificationStatus = SUCCESS` を要求し、未完了のまま user pool 更新へ進まないようにしています。
- `DEVELOPER` mode に切り替えると、Cognito の送信元は `shortsfans <cognito_email_from_address>` になります。
- mail subject / body は Cognito default のままで、今回の branding は sender のみです。
- `DEVELOPER + SES` に切り替えると Cognito default mail の quota ではなく SES 側の quota が使われます。
- ただし SES account が sandbox のままなら、送信先は verified address のみに制限され、上限も SES sandbox の範囲に留まります。
- mail delivery の成否確認は AWS コンソール前提にせず CLI verification で行いますが、SES sandbox 中は送信先も verify 済み address か mailbox simulator に制限されます。

## SES production access

- 本番相当の送信上限に上げるには、`DEVELOPER + SES` へ切り替えたうえで SES production access を別途申請します。
- `使い方` の流れどおり `infra/terraform/dev` にいる前提で、AWS CLI から申請する場合の例:

```bash
aws sesv2 put-account-details \
  --region "$(terraform output -raw aws_region)" \
  --production-access-enabled \
  --mail-type TRANSACTIONAL \
  --website-url "https://replace-with-your-site.example.com" \
  --contact-language JA \
  --use-case-description "Transactional Cognito sign-up verification and password reset emails for shorts-fans." \
  --additional-contact-email-addresses "replace-with-ops@example.com"
```

- `website-url` と contact email は repo で固定していないため、実運用値に置き換えます。
- AWS 側の審査は非同期で、request を送っても即時には上限が上がりません。

## CLI verification

まず repo root へ戻り、Terraform output を shell に取り込みます。

```bash
cd /path/to/shorts-fans

export BACKEND_AWS_REGION="$(terraform -chdir=infra/terraform/dev output -raw aws_region)"
export BACKEND_COGNITO_USER_POOL_ID="$(terraform -chdir=infra/terraform/dev output -raw cognito_user_pool_id)"
export BACKEND_COGNITO_USER_POOL_CLIENT_ID="$(terraform -chdir=infra/terraform/dev output -raw cognito_user_pool_client_id)"
export BACKEND_COGNITO_ISSUER_URL="$(terraform -chdir=infra/terraform/dev output -raw cognito_user_pool_issuer_url)"
export BACKEND_COGNITO_EMAIL_SENDING_ACCOUNT="$(terraform -chdir=infra/terraform/dev output -raw cognito_email_sending_account)"
export BACKEND_COGNITO_FROM_EMAIL_ADDRESS="$(terraform -chdir=infra/terraform/dev output -raw cognito_email_from_address)"
```

`BACKEND_COGNITO_EMAIL_SENDING_ACCOUNT` が `DEVELOPER`、`BACKEND_COGNITO_FROM_EMAIL_ADDRESS` が `shortsfans <...>` になっていることを先に確認します。

verification 用の一時値を用意します。

```bash
export TEST_EMAIL="replace-with-verified-recipient@example.com"
export TEST_PASSWORD="TempPass123!"
export NEW_TEST_PASSWORD="TempPass456!"
```

- SES sandbox のまま検証する場合、`TEST_EMAIL` は SES で verify 済み recipient に置き換えます。
- production access 承認後は任意の実受信 inbox や使い捨て mail で同じ手順を使えます。

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

- Cognito slice の Terraform 定義を変えたあとで元に戻すときは、差分を戻すか `cognito_use_ses_developer_email = false` に戻して `terraform apply -var-file=terraform.tfvars` を再実行します。
- 開発環境全体を破棄するときだけ `terraform destroy -var-file=terraform.tfvars` を使います。
- destroy すると media / avatar / Cognito slice が同じ root からまとめて消えるため、必要な output を事前に控えておきます。

## 後続 task への handoff

- `SHO-168` の current backend runtime は `AWS_REGION` と `COGNITO_USER_POOL_CLIENT_ID` を使って public Cognito API へ接続します。
- `COGNITO_USER_POOL_ID` は CLI verification の `admin-delete-user` や将来の issuer / admin API integration では引き続き使えるため、Terraform output として維持します。
- App Client が secretless なので、backend は `USER_PASSWORD_AUTH` / sign-up / password reset 系の public Cognito API をそのまま扱う前提です。
- `DEVELOPER + SES` では Cognito が SES を代理送信するため、backend 側の public API call や request/response contract の変更は不要です。
- issuer/JWKS を使う token validation を後続で入れる場合は `cognito_user_pool_issuer_url` を起点にします。
