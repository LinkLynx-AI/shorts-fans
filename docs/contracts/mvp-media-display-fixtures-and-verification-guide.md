# Media Display Fixtures And Verification Guide

## 位置づけ

- この文書は `SHO-157 fixture / smoke / docs を追従する` の成果物です。
- `SHO-158`、`SHO-159`、`SHO-154`、`SHO-155`、`SHO-156` で散って更新された media workflow / surface 表示の fixture と verification 根拠を、`upload complete -> processing -> ready -> public short -> main playback -> creator owner preview` の順で辿れるようにします。
- [fan-mvp-fixtures-and-integration-guide.md](fan-mvp-fixtures-and-integration-guide.md) は fan surface 専用の guide として維持し、この文書は creator upload と owner preview を含む media workflow の入口を補います。

## Goals

- upload 完了から各 surface 表示まで、どの contract / fixture / 既存検証を見ればよいかを固定する。
- `backend-media-smoke` が infra 到達性 smoke であり、workflow-level verification の代替ではないことを明示する。
- transport fixture と backend-internal verification の境界を混ぜない。

## Non-goals

- 新しい wire contract や新しい business rule の追加
- 新しい smoke binary、E2E route、Make target の追加
- `processing` のような internal boundary に対して transport fixture を新設すること

## Canonical Sources

- [creator-upload-api-contract.md](creator-upload-api-contract.md)
- [mvp-media-workflow-contract.md](mvp-media-workflow-contract.md)
- [media-display-access-contract.md](media-display-access-contract.md)
- [fan-public-surface-api-contract.md](fan-public-surface-api-contract.md)
- [fan-unlock-main-api-contract.md](fan-unlock-main-api-contract.md)
- [creator-workspace-owner-preview-api-contract.md](creator-workspace-owner-preview-api-contract.md)
- [fan-mvp-fixtures-and-integration-guide.md](fan-mvp-fixtures-and-integration-guide.md)
- [../infra/dev-media-smoke.md](../infra/dev-media-smoke.md)

## Flow Map

| phase | primary contract/docs | representative fixture | current verification entry | notes |
| --- | --- | --- | --- | --- |
| `upload complete` | `creator-upload-api-contract.md`, `mvp-media-workflow-contract.md` | `fixtures/creator-upload.json` | `cd backend && go test ./internal/creatorupload ./internal/httpserver` | completion request/validation/persistence と `processing_state = uploaded` の入口を確認する。現行 repo では backend contract test が主根拠。 |
| `processing -> ready -> auto-publish` | `mvp-media-workflow-contract.md`, `media-display-access-contract.md` | なし | `cd backend && go test ./internal/media` | `processing` は internal boundary なので transport fixture は置かない。worker、materializer、state 遷移、auto-publish bridge は backend test を正とする。 |
| `public short` | `fan-public-surface-api-contract.md`, `media-display-access-contract.md` | `fixtures/fan-public-surfaces.json` | `cd backend && go test ./internal/httpserver ./internal/media` / `cd frontend && pnpm test:e2e tests/e2e/route-shell.spec.ts` | public short の representative payload は fixture を正とし、frontend shell smoke は `frontend/scripts/mock-e2e-api-server.mjs` 経由で public surface fixture を読む。 |
| `main playback` | `fan-unlock-main-api-contract.md`, `media-display-access-contract.md` | `fixtures/fan-unlock-main.json` | `cd backend && go test ./internal/httpserver ./internal/media` / `cd frontend && pnpm test:unit -- src/features/unlock-entry/api/request-unlock-surface.test.ts src/widgets/main-playback-surface/api/request-main-playback-surface.test.ts 'src/app/api/fan/mains/[mainId]/access-entry/route.test.ts'` / `cd frontend && pnpm test:e2e tests/e2e/route-shell.spec.ts` | short から unlock / access-entry / playback までを分けて確認する。shell E2E は short detail から unlocked / owner preview 遷移の回帰確認に使う。 |
| `creator owner preview` | `creator-workspace-owner-preview-api-contract.md`, `media-display-access-contract.md` | `fixtures/creator-workspace-owner-preview.json` | `cd backend && go test ./internal/creator ./internal/httpserver ./internal/media` / `cd frontend && pnpm test:unit -- src/widgets/creator-mode-shell/api/get-creator-workspace-preview-collections.test.ts src/widgets/creator-mode-shell/api/get-creator-workspace-preview-detail.test.ts` | owner preview list/detail は transport fixture と fetcher/unit test が current consumer root。現時点で dedicated Playwright smoke はない。 |

## Verification Bundles

実装全体をまとめて確認するときは、次の bundle を基準にします。

### Backend workflow bundle

```bash
cd backend && go test ./internal/creatorupload ./internal/media ./internal/httpserver ./internal/creator
```

- upload acceptance、worker/materializer、public short、main playback、owner preview の backend 側根拠をまとめて確認します。
- `processing` 境界はこの bundle の `./internal/media` で確認し、transport fixture の追加対象にはしません。

### Frontend consumer bundle

```bash
cd frontend && pnpm test:unit -- src/features/unlock-entry/api/request-unlock-surface.test.ts src/widgets/main-playback-surface/api/request-main-playback-surface.test.ts src/widgets/creator-mode-shell/api/get-creator-workspace-preview-collections.test.ts src/widgets/creator-mode-shell/api/get-creator-workspace-preview-detail.test.ts 'src/app/api/fan/mains/[mainId]/access-entry/route.test.ts'
cd frontend && pnpm test:e2e tests/e2e/route-shell.spec.ts
```

- unit 側は unlock / main playback / creator workspace owner preview の consumer parsing を確認します。
- Playwright 側は feed から short detail、unlock、main playback、owner preview entry までの shell-level regression を確認します。

### Infra smoke

```bash
make backend-media-smoke
```

- `public short bucket` と `private main signed URL` に到達できるかを見る infra smoke です。
- upload completion、worker dequeue、`media_asset.processing_state` 遷移、auto-publish、workspace preview の代替には使いません。

## Usage Rules

- representative fixture は transport contract の canonical example として扱い、internal processing state の truth には使いません。
- `processing`、`ready`、`approved_for_publish`、`approved_for_unlock` の判定は backend test と workflow contract を正とします。
- shell smoke が通っても upload / materialization が成立したとはみなさず、backend workflow bundle を別で確認します。
- `fan-mvp-fixtures-and-integration-guide.md` には fan surface 専用の接続順を残し、この文書には creator upload / owner preview を含む cross-surface traceability を集約します。
