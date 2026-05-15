# dev AWS app environment

## 位置づけ

- この文書は `infra/terraform/dev` に追加した full dev app environment の運用入口です。
- 既存の media / avatar / review evidence / Cognito sandbox と同じ Terraform root を使い、backend API、worker、frontend、PostgreSQL、Redis を AWS dev 上で動かします。
- production 環境ではありません。custom domain、HTTPS certificate、WAF、remote state、CI deploy は今回の対象外です。

## この Terraform が作るもの

- VPC
  - public subnet 2 つ、private subnet 2 つ
  - NAT Gateway は作りません
- ALB
  - `http://...` の public entrypoint
  - `/api/*`、`/healthz`、`/readyz` は backend API へ forward
  - その他の path は frontend へ forward
- ECS Fargate
  - backend API service
  - backend worker service
  - frontend service
  - migration / seed 用 one-off maintenance task definition
- ECR
  - backend image repository
  - frontend image repository
- RDS PostgreSQL 16
  - private subnet 配置
  - dev 用の single instance
- ElastiCache Redis
  - private subnet 配置
  - dev 用の single node
- CloudWatch Logs
  - backend / worker / frontend / maintenance の log group
- Secrets Manager / Parameter Store
  - `POSTGRES_DSN` と `ADMIN_API_TOKEN` は Secrets Manager
  - non-secret runtime handoff は SSM Parameter Store
- IAM
  - ECS execution role
  - app task role
  - 既存の media / avatar / review evidence access policy を app task role に attach
  - Cognito public auth flow に必要な最小権限

## deploy 手順

1. tfvars を用意します。

```bash
cp infra/terraform/dev/terraform.tfvars.example infra/terraform/dev/terraform.tfvars
```

2. `infra/terraform/dev/terraform.tfvars` の `aws_region`、`admin_api_token`、`dev_allowed_cidr_blocks` を dev 用に変更します。

`dev_allowed_cidr_blocks` は dev ALB に到達できる送信元 CIDR です。HTTP-only の dev 環境なので `0.0.0.0/0` は許可せず、自分の固定 IP や VPN の CIDR に絞ります。

3. Terraform を初期化します。

```bash
terraform -chdir=infra/terraform/dev init
```

4. ECR repository を Terraform 管理下で先に作り、backend / frontend image を push します。

```bash
make aws-dev-build-push AWS_DEV_IMAGE_TAG=dev
```

5. full dev environment を apply します。

```bash
make aws-dev-apply
```

6. migration と seed を ECS one-off task で流します。

```bash
make aws-dev-migrate
make aws-dev-seed
```

7. app URL を確認します。

```bash
terraform -chdir=infra/terraform/dev output -raw dev_app_url
```

## 代表 smoke

`DEV_APP_URL` に `dev_app_url` output を入れて確認します。

```bash
export DEV_APP_URL="$(terraform -chdir=infra/terraform/dev output -raw dev_app_url)"

curl -fsS "$DEV_APP_URL/healthz"
curl -fsS "$DEV_APP_URL/readyz"
curl -fsS "$DEV_APP_URL/"
curl -fsS "$DEV_APP_URL/api/fan/feed?tab=recommended"
```

## image 更新

同じ tag を再利用して image を push した場合、ECS service は自動では新 image を pull しません。再 deploy だけ行う場合は次を使います。

```bash
make aws-dev-build-push AWS_DEV_IMAGE_TAG=dev
make aws-dev-force-deploy
```

tag を変える場合は `terraform.tfvars` の `backend_image_tag` / `frontend_image_tag` と `AWS_DEV_IMAGE_TAG` を揃えてから `make aws-dev-build-push` と `make aws-dev-apply` を実行します。

## Guardrail

- dev ALB は HTTP only です。production 相当の TLS / WAF はこの環境では扱いません。
- RDS と Redis は private subnet に置き、ECS task security group からのみ許可します。
- ECS task は public subnet に置き、public IP を割り当てます。これは NAT Gateway を避ける dev コスト優先の判断です。
- frontend は browser runtime では `NEXT_PUBLIC_API_BASE_URL=__same_origin__` で ALB 同一 origin の `/api` に向け、server runtime では `API_BASE_URL_INTERNAL` で ECS service discovery の backend private DNS に向けます。
- AWS-hosted frontend の ALB origin は Terraform が S3 direct upload CORS に自動追加します。localhost origin は local 開発用に残します。
- `terraform destroy` は dev 環境全体を消す場合だけ使います。
