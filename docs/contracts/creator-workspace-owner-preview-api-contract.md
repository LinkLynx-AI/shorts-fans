# Creator Workspace Owner Preview API Contract

## 位置づけ

- この文書は creator owner が自分の `short` / `main` を preview する private contract を固定します。
- `/api/creator/workspace` overview 契約とは分離し、owner preview list/detail と short caption mutation だけを対象にします。
- review detail、analytics、売上、upload processing detail はこの leaf に含めません。

## Goals

- creator owner が delivery-ready asset を preview する最小 read surface を固定する。
- `short` preview と `main` preview の list/detail を cursor/list boundary と detail boundary で分ける。
- public short や fan main playback と混線しない owner-only access を明示する。
- short detail から caption を modal で更新する最小 mutation surface を固定する。

## Non-goals

- `/api/creator/workspace` overview の拡張
- upload queue、processing state、review detail
- sales / analytics / moderation 指標
- public publish / unlock eligibility の mutation
- short caption 以外の post action mutation (`price`, `非公開`, `削除`)

## Canonical Sources

- `docs/contracts/media-display-access-contract.md`
- `docs/contracts/submission-package-review-contract.md`
- `docs/contracts/creator-workspace-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-media-workflow-contract.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/creator/workspace/shorts` | required | owner preview 用 short list |
| `GET` | `/api/creator/workspace/mains` | required | owner preview 用 main list |
| `GET` | `/api/creator/workspace/shorts/{shortId}/preview` | required | owner preview 用 short detail |
| `GET` | `/api/creator/workspace/mains/{mainId}/preview` | required | owner preview 用 main detail |
| `PUT` | `/api/creator/workspace/shorts/{shortId}/caption` | required | owner preview short の caption 更新 |

## Surface-specific Payloads

### `PreviewCardVideoAsset`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | asset identifier |
| `kind` | `"video"` | video 固定 |
| `posterUrl` | `string` | preview card 表示用 poster |
| `durationSeconds` | `number` | rounded-up seconds |

### `WorkspacePreviewShortItem`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | short identifier |
| `canonicalMainId` | `string` | linked canonical main |
| `media` | `PreviewCardVideoAsset` | poster-only preview asset |
| `previewDurationSeconds` | `number` | short length |

### `WorkspacePreviewMainItem`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | main identifier |
| `leadShortId` | `string` | preview card の lead short context |
| `media` | `PreviewCardVideoAsset` | poster-only preview asset |
| `durationSeconds` | `number` | main length |
| `priceJpy` | `number` | reference price |

## Request Contract

### `GET /api/creator/workspace/shorts`

#### Query

| field | type | required |
| --- | --- | --- |
| `cursor` | `string` | no |

#### Response

- `data.items`: `WorkspacePreviewShortItem[]`
- `meta.page`: `CursorPageInfo`

### `GET /api/creator/workspace/mains`

#### Query

| field | type | required |
| --- | --- | --- |
| `cursor` | `string` | no |

#### Response

- `data.items`: `WorkspacePreviewMainItem[]`
- `meta.page`: `CursorPageInfo`

### `GET /api/creator/workspace/shorts/{shortId}/preview`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Response

- `data.preview.short`: `ShortSummary`
- `data.preview.creator`: `CreatorSummary`
- `data.preview.access`: `MainAccessState`
- `data.preview.access.status` は `owner` 固定
- `meta.page = null`

### `GET /api/creator/workspace/mains/{mainId}/preview`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Response

- `data.preview.main.id`: `string`
- `data.preview.main.media`: `VideoDisplayAsset`
- `data.preview.main.durationSeconds`: `number`
- `data.preview.main.priceJpy`: `number`
- `data.preview.creator`: `CreatorSummary`
- `data.preview.access`: `MainAccessState`
- `data.preview.access.status` は `owner` 固定
- `data.preview.entryShort`: `ShortSummary`
- `meta.page = null`

### `PUT /api/creator/workspace/shorts/{shortId}/caption`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `caption` | `string` | yes | modal input の raw string をそのまま送る |

#### Response

- `data.short.id`: `string`
- `data.short.caption`: `string`
- `meta.page = null`

## Response Rules

- caller は authenticated viewer である必要があります。
- caller は approved creator capability を持つ必要があります。
- caller は preview 対象の owner 自身である必要があります。
- list endpoint は `delivery-ready` な owner 自身の `short` / `main` を返し、public publish / unlock state とは独立に判定して構いません。
- owner preview は private surface であり、`submission package ready` のような readiness predicate と、`pending review` / `approved for publish` / `approved for unlock` のような review submit・decision state・publish-unlock gate とは独立して扱い、review 承認を必須条件にしません。
- owner preview では `MainAccessState.reason = owner_preview` を使い、`unlocked` と混ぜません。
- list endpoint は poster 中心の preview card を返し、detail endpoint だけ full playback 用 `url` を返します。
- caption mutation は request body の前後空白を trim して保存し、trim 後に空文字なら DB では `null` として扱います。
- caption mutation の read surface では `null` を返さず、empty string として返します。
- caption mutation 成功後は caller が short detail を再取得して最新値に同期する前提にします。

## Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | invalid JSON、unknown field、`caption` 欠落 |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | preview 対象を解決できない、または caller 所有の short ではない |
| `500` | `internal_error` | unexpected failure |

## Guardrails

- analytics、gross revenue、unlock count を返しません。
- review reason、processing state、storage ref を返しません。
- public creator profile の代替にしません。
- `/api/creator/workspace` overview response に list/detail payload を混ぜません。
- caption mutation に `price`, `非公開`, `削除` を混ぜません。

## Fixture Reference

- representative fixture は [creator-workspace-owner-preview.json](fixtures/creator-workspace-owner-preview.json) を参照します。
