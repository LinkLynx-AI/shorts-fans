# Backend ディレクトリ構成

## この文書の目的

この文書は、`backend/` 配下のディレクトリ構成と package 配置方針を明文化するためのものです。
`docs/GO.md` の Go 実装ルールを前提にしつつ、このリポジトリで backend 機能を追加するときの具体的な配置ルールを定めます。

## 位置づけ

- この文書は backend の構成方針を定めます。
- `backend/README.md` は起動方法と現状の骨格説明を扱います。
- `docs/GO.md` は Go 全般の実装ルールを扱います。
- この文書に書かれていない詳細は `docs/GO.md` に従います。

## 現時点で確定している前提

- backend の module root は `backend/` とします。
- 実行単位は `cmd/` 配下に分けます。
- migration と SQL ソースは `db/` 配下に置きます。
- アプリケーション内部の package は `internal/` 配下に置きます。
- 現在の基盤 package である `config`、`httpserver`、`postgres`、`redis`、`sqs` は `internal/` 配下に置きます。

## 基本方針

### 1. `cmd/` は entrypoint と配線だけにする

- `cmd/api`、`cmd/worker`、`cmd/migrate` には `main.go` と、その実行単位に必要な最小限の wiring だけを置きます。
- business logic は `cmd/` に置きません。

### 2. 共有技術 package と業務 package を分ける

- PostgreSQL、Redis、SQS、設定、認証、ロギングなどの技術基盤は `internal/` 配下の共有 package に置きます。
- `creator`、`shorts`、`subscription`、`feed`、`media` などの業務ロジックは `internal/<domain>/` に置きます。
- 共有技術 package に業務判断を混ぜません。

### 3. 業務ロジックは domain 単位で閉じる

- 新しい business logic は、まず `internal/<domain>/` を作ってそこに閉じ込めます。
- top-level で `handler/`、`service/`、`repository/` のような横断ディレクトリは作りません。
- HTTP や DB の都合より、プロダクトの業務用語を優先して package を切ります。

### 4. SQL の責務を分ける

- SQL ソースの root は `db/queries/` に置きます。
- domain ごとの query が増えたら `db/queries/<domain>/` に分けます。
- skeleton 段階の共通 query や bootstrap 用 query は `db/queries/` 直下に置いて構いません。
- migration は `db/migrations/` に置きます。
- `sqlc` の生成コードは `internal/postgres/sqlc/` に置きます。
- 手書きの repository adapter は generated code と分け、各 domain package 側で責務を持たせます。

### 5. HTTP は transport に留める

- `internal/httpserver` は router、middleware、request/response の変換、graceful shutdown など transport の責務に留めます。
- 認可判定、公開可否判定、課金導線判定などの業務判断は各 domain package に置きます。
- handler から `sqlc` generated code を直接呼ばない構造を優先します。

## 推奨ディレクトリ例

以下は feature が増えたときの推奨例です。存在していない directory も含みますが、追加時はこの方針に従います。

```text
backend/
  cmd/
    api/
    worker/
    migrate/

  db/
    migrations/
    queries/
      bootstrap.sql
      creator/
      shorts/
      subscription/
      feed/
      media/

  internal/
    config/
    httpserver/
    postgres/
      sqlc/
      tx.go
    redis/
    sqs/
    auth/
    logging/

    creator/
      create.go
      repository.go
      errors.go

    shorts/
      create.go
      publish.go
      policy.go
      repository.go

    subscription/
      entitlement.go
      repository.go

    feed/
      ranker.go
      repository.go

    media/
      repository.go

    jobs/
      transcode/
      moderation/
      thumbnail/
```

## ファイル命名ルール

- 惰性で `service.go` や `model.go` を増やさないこと。
- 可能な限り `create.go`、`publish.go`、`entitlement.go`、`policy.go`、`ranker.go` のように振る舞いベースで分けます。
- `repository.go` は repository 抽象が実際に必要なときだけ置きます。
- interface は利用側で定義します。

## 依存方向

- `cmd/*` は `internal/*` に依存してよいです。
- `internal/httpserver` は各 domain package に依存してよいです。
- 各 domain package は共有技術 package に依存してよいですが、HTTP transport には依存しません。
- domain package から `gin` など transport 固有型を露出させません。
- handler から DB generated code へ直接依存する構造は避けます。

## 未確定事項

以下はまだ確定していません。必要になった時点で別途判断します。

- `internal/httpserver` を将来 `internal/http/` に分割するか
- 共有技術 package を将来 `internal/platform/` 配下に再編するか
- `sqlc` 生成物を将来 context ごとに分割出力するか

## 運用ルール

- backend に新しい feature を追加するときは、まずこの文書に反していないかを確認します。
- この文書を更新した場合は、`AGENTS.md` と `docs/README.md` の索引も合わせて更新します。
