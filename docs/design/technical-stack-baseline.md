# 技術スタック前提

## この文書の目的

この文書は、このサービスをどの技術スタックで実装していくかの前提を固定するための設計書です。
画面仕様や API 詳細、運用詳細を詰める前に、まず大枠の構成を明文化します。

この文書に書かれたものだけを現時点の確定事項として扱います。
書かれていないものや、本文中で保留としたものは未確定です。

## 全体方針

- Frontend は `Next.js + TypeScript` を採用します。
- Backend は `Go` を採用します。
- 配信形態は `Web / PWA` を主体にします。
- インフラは `AWS` を本体とし、`Cloudflare` を前段に置きます。
- 成人向けサービスとしての BAN 耐性を考慮し、動画本体は AWS 側に保持します。
- 現時点のリポジトリ構成は `frontend/` と `backend/` を前提にします。

## 現時点で確定していること

### Frontend

- Frontend の主要フレームワークは `Next.js` です。
- Frontend の実装言語は `TypeScript` です。
- 一般ユーザー向け画面、クリエイター向け画面、管理画面は、基本的に Next.js で構築します。
- 認証後 UI、動画フィード表示、サブスク導線も Frontend の責務に含みます。
- `Web / PWA` を前提に画面体験を設計します。
- `SSR / CSR` は画面や機能に応じて使い分けます。

### Backend

- Backend の実装言語は `Go` です。
- 外部公開 API は基本的に `REST` とします。
- API 契約は `OpenAPI` ベースで管理する前提です。
- 初期構成は `モジュラモノリス` とします。
- 重い処理だけを必要に応じて worker に分離します。

### Backend の責務

- 認証
- ユーザー管理
- クリエイター管理
- 投稿メタデータ管理
- フィード取得
- サブスク状態管理
- entitlement（視聴権限）判定
- 通報・モデレーション関連 API
- アップロード開始 API
- アップロード完了 API

### インフラ全体

- インフラ前段には `Cloudflare` を置きます。
- 本体インフラには `AWS` を使います。

### Cloudflare の役割

- DNS
- WAF
- Bot 対策
- Rate limiting
- Web / API の前段保護

Cloudflare は動画本体 CDN の主軸ではなく、前段保護を主目的として使います。

### AWS の主要構成

#### アプリ実行基盤

- `ECS Fargate`
  - Go API
  - Next.js フロント
  - 必要に応じて worker

#### ロードバランサ

- `ALB`

#### データベース

- `PostgreSQL (RDS)`
  - user
  - creator
  - subscription
  - entitlement
  - post metadata
  - moderation state
  - その他の正本データ

#### キャッシュ / 一時データ

- `Redis`
  - cache
  - rate limit 補助
  - hot data
  - 一時状態
  - 冪等制御補助

#### オブジェクトストレージ

- `S3`
  - 元動画
  - 変換後動画
  - サムネイル
  - その他メディア資産

#### 動画変換

- `AWS Elemental MediaConvert`
  - HLS 変換
  - 複数解像度生成
  - サムネイル生成

#### 動画配信

- `CloudFront`
  - 動画配信 CDN
  - signed URL / signed cookies 前提

#### 非同期基盤

- `SQS`
  - 変換後処理
  - モデレーションジョブ
  - 通知ジョブ
  - 集計系処理

#### 検索

- `OpenSearch`
  - 動画検索
  - クリエイター検索
  - タグ検索
  - 管理画面検索

#### 監視

- `CloudWatch`
- `OpenTelemetry`
- `Sentry`

#### IaC

- `Terraform`

### 動画処理の基本方針

- クライアントは API から upload 許可を得ます。
- 動画アップロードは `S3 presigned URL` による直接アップロードを前提にします。
- upload 完了後、Go API がジョブを投入します。
- `MediaConvert` で HLS へ変換します。
- 変換後動画は `S3 + CloudFront` で配信します。

### データの責務分担

- `PostgreSQL`: 正本データ
- `Redis`: キャッシュ / 一時状態
- `OpenSearch`: 検索用派生データ
- `S3`: 動画・画像実体
- `SQS`: 非同期ジョブ
- `CloudFront`: 配信
- `Cloudflare`: 前段保護

### 運用方針

- 環境は `dev / staging / prod` を分離します。
- DB migration は `forward-only` です。
- `feature flag` 前提で開発します。
- 動画本体は AWS 側に置きます。
- Cloudflare は動画の主ホスティングではなく前段保護中心で使います。

### 現時点で採用候補から外したもの

- NestJS
- Kubernetes / EKS
- Kafka
- 自前 FFmpeg 基盤中心の初期構成
- Cloudflare を動画本体 CDN の主軸にする構成

## 現時点ではまだ固定しないこと

この段階では、以下は未確定です。

- 認証基盤の具体プロダクト
- 決済基盤の具体プロダクト
- 年齢確認 / 本人確認の具体プロダクト
- OpenAPI の生成運用フロー
- モジュラモノリス内のモジュール分割単位
- worker の分離粒度と実行方式の詳細
- Redis のマネージドサービス詳細
- OpenSearch の index 設計詳細
- Cloudflare 側の各機能の具体的な商品プラン
- PWA としてどこまで offline 対応するか
- どの画面を SSR にし、どの画面を CSR にするか
- Secrets 管理を `AWS Secrets Manager` にするか `SSM Parameter Store` にするか
- observability のメトリクス / トレース / エラー収集の責務分担詳細
- Terraform の module 分割方針

## この文書の使い方

- 新しい技術選定を行うときは、まずこの文書の確定事項と矛盾しないかを確認します。
- 実装都合だけで別技術を既成事実化しません。
- 詳細な設計が必要なものは、この文書の前提を変えずに別文書へ分離します。
- この文書で未確定のものを決めたときは、ここへ追記してから詳細文書へ展開します。
