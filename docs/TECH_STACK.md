# 技術選定

## この文書の目的

この文書は、このリポジトリで採用する技術スタックを明文化するための文書です。
ここに記載されたものを、現時点での実装前提として扱います。

## 確定している技術選定

### Frontend

- `Next.js`
- `TypeScript`
- `PWA`

### Backend

- `Go`
- `Gin`
- `sqlc`
- `pgx`

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

- `CCBill`

### Moderation

- 自動判定 + 人手審査

### Analytics

- まずは `PostgreSQL + S3` 集計
- 後で拡張

## 一覧

- Frontend: `Next.js + TypeScript + PWA`
- Backend: `Go + Gin + sqlc + pgx`
- DB: `PostgreSQL`
- Cache: `Redis`
- Queue: `SQS`
- Worker: `Go`
- Infra: `AWS ECS Fargate`
- Storage: `S3`
- Transcode: `MediaConvert`
- Delivery: `CloudFront`
- IaC: `Terraform`
- Payment: `CCBill`
- Moderation: `自動判定 + 人手審査`
- Analytics: `まずは PostgreSQL + S3 集計、後で拡張`
