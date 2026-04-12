# shorts-fans

短尺動画の連続視聴から `main unlock` につなぐ fan / creator 向けサービスの開発リポジトリです。

## 起動方法

### 前提

- Docker / Docker Compose が使えること
- Go `1.24` が入っていること
- Node.js `20.9.0` 以上と `pnpm@9.6.0` が使えること
- AWS CLI で `main-admin` profile を使えること

### 1. Notion から `.env` と `frontend/.env` をコピーする

Notion にある root 用 `.env` の内容をコピーして、repo root で次を実行します。

```bash
pbpaste > .env
```

Notion にある frontend 用 `frontend/.env` の内容をコピーして、repo root で次を実行します。

```bash
pbpaste > frontend/.env
```

### 2. AWS に `main-admin` でアカウントを作成する

この手順だけは手動で行ってください。

### 3. AWS に `main-admin` でログインする

```bash
aws sso login --profile main-admin
```

### 4. 依存を入れてローカル依存サービスを起動する

```bash
(cd frontend && pnpm install)
make backend-dev-up
```

### 5. 3 つのプロセスを別ターミナルで起動する

ターミナル 1:

```bash
make backend-run
```

ターミナル 2:

```bash
make backend-worker
```

ターミナル 3:

```bash
cd frontend && pnpm dev
```

起動後の確認先:

- frontend: `http://localhost:3000`
- backend API: `http://localhost:8080`

停止するときは repo root で次を実行します。

```bash
make backend-dev-down
```
