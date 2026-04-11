# Fan Public Surface API Contract

## 位置づけ

- この文書は `SHO-17 feed / short detail / creator profile の API 契約とモックデータを定義する` の成果物です。
- primary loop の public surface である `feed / short detail / creator search / creator profile` の read contract を固定します。
- DTO と response envelope は `docs/contracts/fan-mvp-common-transport-contract.md` を参照します。

## Canonical Sources

- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/ssot/product/fan/fan-journey.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/content/content-model.md`
- `docs/ssot/product/content/short-main-linkage.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/fan/feed` | `tab=recommended` は optional、`tab=following` は required | short feed |
| `GET` | `/api/fan/shorts/{shortId}` | optional | short detail |
| `GET` | `/api/fan/creators/search` | optional | creator search only |
| `GET` | `/api/fan/creators/{creatorId}` | optional | public creator profile header |
| `GET` | `/api/fan/creators/{creatorId}/shorts` | optional | public creator short grid |

## Request Contract

### `GET /api/fan/feed`

#### Query

| field | type | required | notes |
| --- | --- | --- | --- |
| `tab` | `"recommended" \| "following"` | yes | `following` は認証済み viewer 前提 |
| `cursor` | `string` | no | cursor pagination |

#### Response

- `data.tab`: request と同じ tab
- `data.items`: `FeedItem[]`
- `meta.page`: `CursorPageInfo`
- `recommended` は `published_at DESC, id DESC` の newest-first で返します。
- `following` も同じ cursor 順序で返します。

### `FeedItem`

| field | type | notes |
| --- | --- | --- |
| `short` | `ShortSummary` | feed 上の public short |
| `creator` | `CreatorSummary` | creator block 表示用 |
| `viewer.isPinned` | `boolean` | pin secondary action 表示用 |
| `viewer.isFollowingCreator` | `boolean` | current viewer の creator follow state |
| `unlockCta` | `UnlockCtaState` | 下部固定 CTA 用。price と main 長さはここから組み立てる |

#### HTTP States

| case | status |
| --- | --- |
| `recommended` success | `200` |
| `following` success | `200` |
| `following` empty | `200` |
| `following` unauthenticated | `401` + `auth_required` |

### `GET /api/fan/shorts/{shortId}`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Response

- `data.detail`: `ShortDetail`
- `meta.page = null`

#### HTTP States

| case | status |
| --- | --- |
| normal `public` | `200` |
| `unlocked` | `200` |
| `owner` | `200` |
| `not_found` | `404` + `not_found` |

### `GET /api/fan/creators/search`

#### Query

| field | type | required | notes |
| --- | --- | --- | --- |
| `q` | `string` | no | empty string または未指定なら recent creators を返す |
| `cursor` | `string` | no | cursor pagination |

#### Matching Rule

- search target は `displayName` と `handle` に限定します。
- `short caption / hashtag / full-text search` はこの contract に入れません。

#### Response

- `data.query`: resolved query string
- `data.items`: `CreatorSearchResult[]`
- `meta.page`: `CursorPageInfo`

### `CreatorSearchResult`

| field | type | notes |
| --- | --- | --- |
| `creator` | `CreatorSummary` | search result item |

#### HTTP States

| case | status |
| --- | --- |
| normal | `200` |
| empty | `200` |

### `GET /api/fan/creators/{creatorId}`

#### Path

| field | type | required |
| --- | --- | --- |
| `creatorId` | `string` | yes |

#### Response

- `data.profile.creator`: `CreatorSummary`
- `data.profile.stats.shortCount`: `number`
- `data.profile.stats.fanCount`: `number`
- `data.profile.viewer.isFollowing`: `boolean`
- `meta.page = null`
- follow relation の write contract は `docs/contracts/fan-creator-follow-api-contract.md` を参照します。この read surface 自体は auth optional のまま維持します。

#### Guardrail

- creator profile は `main` の direct list surface に昇格させません。

#### HTTP States

| case | status |
| --- | --- |
| normal | `200` |
| `not_found` | `404` + `not_found` |

### `GET /api/fan/creators/{creatorId}/shorts`

#### Path

| field | type | required |
| --- | --- | --- |
| `creatorId` | `string` | yes |

#### Query

| field | type | required | notes |
| --- | --- | --- | --- |
| `cursor` | `string` | no | short grid keyset pagination |

#### Response

- `data.items`: `CreatorProfileShortGridItem[]`
- `meta.page`: `CursorPageInfo`

### `CreatorProfileShortGridItem`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | short identifier |
| `canonicalMainId` | `string` | canonical main target |
| `creatorId` | `string` | short owner |
| `media` | `MediaAsset` | `kind = "video"` |
| `previewDurationSeconds` | `number` | short 自身の長さ |

#### Guardrail

- grid item には `unlockCta`、価格、`mainDurationSeconds` を入れません。
- grid item には `title`、`caption` を入れません。
- creator profile は `main` の direct list surface に昇格させません。

#### HTTP States

| case | status |
| --- | --- |
| normal | `200` |
| empty short grid | `200` |
| `not_found` | `404` + `not_found` |

## State Matrix

| endpoint | public | unlocked | owner | empty | not_found |
| --- | --- | --- | --- | --- | --- |
| `GET /api/fan/feed` | yes | yes | no | yes | no |
| `GET /api/fan/shorts/{shortId}` | yes | yes | yes | no | yes |
| `GET /api/fan/creators/search` | yes | no | no | yes | no |
| `GET /api/fan/creators/{creatorId}` | yes | no | no | no | yes |
| `GET /api/fan/creators/{creatorId}/shorts` | yes | no | no | yes | yes |

## Out-of-scope Guardrails

- `like` と `comment` は返さない
- ranking explanation や debug field は返さない
- `main` の direct listing は返さない
- recommendation や related creators は返さない

## Fixture Reference

- representative fixture は [fan-public-surfaces.json](fixtures/fan-public-surfaces.json) を参照します。
