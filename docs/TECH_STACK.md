# 技術選定

## この文書の目的

この文書は、このリポジトリで採用する技術スタックを明文化するための文書です。
ここに記載されたものを、現時点での実装前提として扱います。

## 現時点の実装前提

### Frontend

- `Next.js`
- `TypeScript`
- `PWA`

### Backend

- `Go`
- `Gin`
- `sqlc`
- `pgx`

### Auth

- `Amazon Cognito User Pools`
- custom UI を primary にし、Hosted UI は fan auth の正規主導線にしません。

### Data / Infra

- `PostgreSQL`
- `Redis`
- `SQS`
- `AWS ECS Fargate`
- `S3`
- `MediaConvert`
- `CloudFront`
- `Terraform`

### Infra 管理方針

- インフラは `Terraform` で管理します。

### Worker

- `Go`

### Payment

- `Payment provider` は未確定
- first-pass candidate は `CCBill / Segpay / Verotel`

### Moderation

- `MVP` は `manual-heavy` を前提にした人手審査中心
- 自動判定は後で拡張

### Analytics

- まずは `PostgreSQL + S3` 集計
- 後で拡張

## 一覧

- Frontend: `Next.js + TypeScript + PWA`
- Backend: `Go + Gin + sqlc + pgx`
- Auth: `Amazon Cognito User Pools (fan auth, custom UI primary)`
- DB: `PostgreSQL`
- Cache: `Redis`
- Queue: `SQS`
- Worker: `Go`
- Infra: `AWS ECS Fargate`
- Storage: `S3`
- Transcode: `MediaConvert`
- Delivery: `CloudFront`
- IaC: `Terraform`
- Payment: `未確定 (first-pass candidate: CCBill / Segpay / Verotel)`
- Moderation: `MVP は人手審査中心、後で自動判定を拡張`
- Analytics: `まずは PostgreSQL + S3 集計、後で拡張`
