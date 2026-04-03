# Fan MVP Common Transport Contract

## 位置づけ

- この文書は `SHO-16 fan MVP 共通 DTO と response 契約を定義する` の成果物です。
- `feed / short detail / creator profile / unlock / main player / fan profile` で共有する transport contract をここで固定します。
- ここで定義した DTO 名、state 名、response envelope を後続 surface 文書で再定義しません。

## Goals

- fan MVP surface で共通に使う DTO と response envelope を固定する。
- `public / locked / purchased / owner / empty / not_found` の scenario vocabulary を統一する。
- frontend が UI copy を組み立てやすい raw 値を返し、backend が UI 固有文字列を持ち込まない境界を作る。

## Non-goals

- endpoint ごとの request / response 詳細
- mutation API
- DB schema や storage 都合の field
- ranking、recommendation、payment provider 固有の metadata

## Canonical Sources

- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/fan/fan-journey.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/content/content-model.md`
- `docs/ssot/product/content/short-main-linkage.md`
- `docs/ssot/product/monetization/billing-and-access.md`
- `docs/ssot/product/fan/consumer-state-and-profile.md`
- `docs/ssot/product/account/account-permissions.md`

## Response Envelope

- すべての fan MVP read endpoint は次の envelope を返します。
- list endpoint だけ `meta.page` を持ちます。single resource endpoint では `meta.page = null` とします。
- 正常系は `error = null`、異常系は `data = null` とします。

```json
{
  "data": {
    "id": "short_aoi_softlight"
  },
  "meta": {
    "requestId": "req_feed_recommended_001",
    "page": {
      "nextCursor": "feed:recommended:cursor:001",
      "hasNext": true
    }
  },
  "error": null
}
```

```json
{
  "data": null,
  "meta": {
    "requestId": "req_short_not_found_001",
    "page": null
  },
  "error": {
    "code": "not_found",
    "message": "short was not found"
  }
}
```

### Error Codes

| code | meaning |
| --- | --- |
| `auth_required` | authentication が必要な surface に未認証 viewer が来た |
| `not_found` | creator / short / main が存在しない、または公開対象ではない |
| `main_locked` | main は存在するが viewer は再生可能 state を持たない |

## Scenario Vocabulary

- ここでの語彙は fixture 名と state matrix の統一用です。すべてが wire field になるわけではありません。

| scenario | meaning |
| --- | --- |
| `public` | public surface をそのまま閲覧できる正常系 |
| `locked` | canonical main は存在するが viewer は再生できない |
| `purchased` | viewer が canonical main の paid access を持つ |
| `owner` | viewer が creator owner として preview access を持つ |
| `empty` | request 自体は成功したが items が 0 件 |
| `not_found` | 対象 resource が存在しない、または見せるべきではない |

## Common DTOs

### `MediaAsset`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | asset identifier |
| `kind` | `"image" \| "video"` | MVP では avatar は `image`、short / main は `video` |
| `url` | `string` | public または signed URL。client は opaque に扱う |
| `posterUrl` | `string \| null` | `video` の preview poster。`image` では `null` |
| `width` | `number` | pixel width |
| `height` | `number` | pixel height |
| `durationSeconds` | `number \| null` | `video` だけ値を持つ |

### `CreatorSummary`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | creator identifier |
| `displayName` | `string` | public display name |
| `handle` | `string` | creator search 用の public identifier。core domain minimum を広げる目的ではなく transport 用にだけ持つ |
| `avatar` | `MediaAsset` | `kind = "image"` |
| `bio` | `string` | public creator bio |

### `ShortSummary`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | short identifier |
| `canonicalMainId` | `string` | `1 short : 1 canonical main` の target |
| `creatorId` | `string` | short owner |
| `title` | `string` | short title |
| `caption` | `string` | public caption |
| `media` | `MediaAsset` | `kind = "video"` |
| `previewDurationSeconds` | `number` | short 自身の長さ |

### `ShortDetail`

| field | type | notes |
| --- | --- | --- |
| `short` | `ShortSummary` | short 本体 |
| `creator` | `CreatorSummary` | short owner の public 情報 |
| `viewer.isPinned` | `boolean` | private pin state |
| `viewer.isFollowingCreator` | `boolean` | follow state |
| `unlockCta` | `UnlockCtaState` | short detail から見える CTA 状態 |

### `PurchaseState`

| field | type | notes |
| --- | --- | --- |
| `mainId` | `string` | canonical main identifier |
| `status` | `"not_purchased" \| "purchased"` | creator ownership はここに混ぜない |

### `MainAccessState`

| field | type | notes |
| --- | --- | --- |
| `mainId` | `string` | canonical main identifier |
| `status` | `"locked" \| "purchased" \| "owner"` | playback 可否に使う state |
| `reason` | `"purchase_required" \| "purchased_access" \| "owner_preview"` | `status` の解釈を固定する最小 field |

### `UnlockCtaState`

| field | type | notes |
| --- | --- | --- |
| `state` | `"unlock_available" \| "setup_required" \| "continue_main" \| "owner_preview" \| "unavailable"` | frontend はこの state から CTA label を組み立てる |
| `priceJpy` | `number \| null` | `setup_required` と `unlock_available` で必須 |
| `mainDurationSeconds` | `number \| null` | `setup_required` と `unlock_available` で必須 |
| `resumePositionSeconds` | `number \| null` | `continue_main` で使う。その他は `null` |

### `CursorPageInfo`

| field | type | notes |
| --- | --- | --- |
| `nextCursor` | `string \| null` | 次ページがない場合は `null` |
| `hasNext` | `boolean` | next page の有無 |

## Formatting Rules

- 金額は `priceJpy` の整数値で返し、`¥2,200` のような UI copy は frontend が組み立てます。
- 長さや進捗は `number` 秒で返し、`12分` や `5:12 left` のような表示文字列は frontend が組み立てます。
- backend は `Unlock` や `Continue main` のような UI 固有文言を返しません。

## Fixture Reference

- DTO 例と envelope 例は [fan-mvp-common.json](fixtures/fan-mvp-common.json) を canonical example とします。
