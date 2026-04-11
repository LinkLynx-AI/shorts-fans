# Creator Workspace API Contract

## 位置づけ

- この文書は `/creator` の approved creator workspace における `summary` と `top performers` の read 契約を固定します。
- `GET /api/creator/workspace` は `creator info / overview metrics / revision requested summary` に限定し、`top performers` 自体は別 endpoint に分離します。
- actual transport 実装より先に、frontend mock と backend read boundary の責務を揃えることを目的にします。

## Goals

- `/creator` header に表示する creator 自身の情報を private workspace 契約として固定する。
- overview の最小指標を UI copy ではなく raw 値で返す。
- `Top main` / `Top short` セクションを summary endpoint と分離した dedicated read surface で返す。
- review / analytics / media collection が混ざりやすい `/creator` mock state から、今回の対象を明確に切り出す。

## Non-goals

- `main` / `shorts` grid
- item detail modal
- upload / linkage / review queue 自体の API
- creator registration 時の `handle` 入力仕様

## Canonical Sources

- `docs/contracts/mvp-core-domain-contract.md`
- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/viewer-creator-entry-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/ssot/product/account/account-permissions.md`
- `docs/ssot/product/account/identity-and-mode-model.md`
- `docs/ssot/product/creator/creator-workflow.md`
- `docs/ssot/product/creator/creator-analytics-minimum.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/creator/workspace` | required | creator private workspace の overview と creator info |
| `GET` | `/api/creator/workspace/top-performers` | required | creator private workspace の top main / top short |

## `GET /api/creator/workspace` Response Contract

- `meta.page = null`
- 正常系は `error = null`
- `CurrentViewer` の `id / activeMode / canAccessCreatorMode` は返しません。app shell は引き続き `GET /api/viewer/bootstrap` を正とします。

### `data.workspace.creator`

- shape は `docs/contracts/fan-mvp-common-transport-contract.md` の `CreatorSummary` と同じです。
- `id / displayName / handle / avatar / bio` を返します。
- `avatar` は custom avatar がない場合 `null` を返し、client は既存の platform default avatar / initials fallback を描画します。
- `/creator` では `bio` を本人紹介文として使い、workspace 固有の別説明文は持ちません。

### `WorkspaceOverviewMetrics`

| field | type | notes |
| --- | --- | --- |
| `grossUnlockRevenueJpy` | `number` | creator owner の gross unlock revenue 合計 |
| `unlockCount` | `number` | unlock 件数 |
| `uniquePurchaserCount` | `number` | unlock 済み viewer のユニーク人数 |

- 金額は `priceJpy` と同様に JPY 整数で返し、`¥120,000` のような表示は frontend が組み立てます。
- `K` 表記や localized label は backend が返しません。

### `RevisionRequestedSummary`

| field | type | notes |
| --- | --- | --- |
| `totalCount` | `number` | `main + short` の差し戻し総数 |
| `mainCount` | `number` | `main.state = revision_requested` 件数 |
| `shortCount` | `number` | `short.state = revision_requested` 件数 |

- 差し戻しが 0 件のときは `null` を返します。
- `badge`、`label`、`detail` のような UI 固有文言は返しません。
- `review_reason_code` の語彙はこの leaf では固定しないため、summary は count のみを扱います。

### Success Example

```json
{
  "data": {
    "workspace": {
      "creator": {
        "id": "creator_mina_rei",
        "displayName": "Mina Rei",
        "handle": "@minarei",
        "avatar": {
          "id": "asset_creator_mina_avatar",
          "kind": "image",
          "url": "https://cdn.example.com/creator/mina/avatar.jpg",
          "posterUrl": null,
          "durationSeconds": null
        },
        "bio": "quiet rooftop と hotel light の preview を軸に投稿。"
      },
      "overviewMetrics": {
        "grossUnlockRevenueJpy": 120000,
        "unlockCount": 238,
        "uniquePurchaserCount": 164
      },
      "revisionRequestedSummary": {
        "totalCount": 1,
        "mainCount": 0,
        "shortCount": 1
      }
    }
  },
  "meta": {
    "requestId": "req_creator_workspace_001",
    "page": null
  },
  "error": null
}
```

## `GET /api/creator/workspace/top-performers` Response Contract

- `meta.page = null`
- 正常系は `error = null`
- `topMain` / `topShort` は、それぞれ preview 可能な候補がない場合 `null` を返します。

### `data.topPerformers.topMain`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | main public id |
| `media` | `PreviewCardVideoAsset` | preview card 表示用 poster |
| `unlockCount` | `number` | canonical main の unlock 件数 |

### `data.topPerformers.topShort`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | short public id |
| `media` | `PreviewCardVideoAsset` | preview card 表示用 poster |
| `attributedUnlockCount` | `number` | MVP 暫定として canonical main の unlock 件数を proxy attribution として返す |

- `Top main` は `main_unlocks` 件数の降順、同率時は `created_at DESC, id DESC` で決めます。
- `Top short` は true attribution ではなく、linked canonical main の unlock 件数を proxy attribution として降順で決めます。同率時は `created_at DESC, id DESC` を使います。
- display 不可な candidate は skip し、次順位候補を使います。

### Top Performers Success Example

```json
{
  "data": {
    "topPerformers": {
      "topMain": {
        "id": "main_quiet_rooftop",
        "media": {
          "id": "asset_main_quiet_rooftop",
          "kind": "video",
          "posterUrl": "https://signed.example.com/mains/quiet-rooftop.jpg",
          "durationSeconds": 720
        },
        "unlockCount": 238
      },
      "topShort": {
        "id": "short_quiet_rooftop",
        "media": {
          "id": "asset_short_quiet_rooftop",
          "kind": "video",
          "posterUrl": "https://cdn.example.com/shorts/quiet-rooftop.jpg",
          "durationSeconds": 16
        },
        "attributedUnlockCount": 238
      }
    }
  },
  "meta": {
    "requestId": "req_creator_workspace_top_performers_001",
    "page": null
  },
  "error": null
}
```

## Response Rules

- caller は authenticated viewer である必要があります。
- caller は approved creator capability を持つ必要があります。
- workspace creator は current viewer 自身に固定し、path parameter で他 creator の private workspace を読む形にはしません。
- `handle` は `/creator` header での表示前提で必須とします。
- `avatar = null` は error ではなく、creator registration 時に custom avatar を設定しなかったか、既存 avatar が未設定であることを表します。
- ただし `handle` を creator registration 時にいつ確定させるかは別 PR の責務とし、この文書は approved creator workspace read に必要な shape だけを固定します。
- `GET /api/creator/workspace/top-performers` も同じ auth / creator capability 制約を使います。
- `GET /api/creator/workspace/top-performers` は `managedCollections` や detail view state を返しません。

## Error Contract

| status | code | notes |
| --- | --- | --- |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | creator capability はあるが workspace creator profile を解決できない不整合 |
| `500` | `internal_error` | unexpected failure |

## Boundary Guardrails

- `GET /api/creator/workspace` は `topPerformers`、`managedCollections`、`posters`、detail view 用 state を返しません。
- `GET /api/creator/workspace/top-performers` は `managedCollections`、`posters`、detail view 用 state を返しません。
- `activeMode` や mode switch CTA 情報は返しません。
- review detail list や analytics breakdown は返しません。
- owner preview list/detail は `docs/contracts/creator-workspace-owner-preview-api-contract.md` に分離します。
- owner preview からの main price / short caption mutation は `docs/contracts/creator-workspace-metadata-edit-api-contract.md` に分離します。
- public creator profile をこの endpoint で代替しません。

## Fixture Reference

- representative fixture は [creator-workspace.json](fixtures/creator-workspace.json) を参照します。
