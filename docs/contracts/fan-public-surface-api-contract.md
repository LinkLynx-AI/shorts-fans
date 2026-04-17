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
- `docs/ssot/business/data/data-strategy.md`

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
- `data.items`: `FeedItem[]`。item order 自体が resolved feed order を表します
- `meta.page`: `CursorPageInfo`

#### Ranking Semantics

- `recommended` は `public short` を候補集合にし、request 時点で解決された recommendation order を返します。
- `recommended` は auth optional のまま維持し、unauthenticated viewer に対しても閲覧可能な fallback order を返せます。newest-first 固定には戻しません。
- `following` は follow 中 creator の `public short` だけを候補集合にし、その中で解決された recommendation order を返します。
- この contract は concrete score formula や feature set を固定しませんが、`short -> main -> unlock` の主導線を壊さない順序解決を前提にします。

#### Order Interpretation

- resolved order は単純な recency 順を保証しません。consumer は同一 creator や同一 canonical main の item が返却順で間引かれたり、間隔を空けて並ぶ場合があることを前提にします。
- consumer は、viewer がすでに unlock 済みの canonical main に紐づく short や、server 側で十分視聴済みとして扱われる short が、他の eligible candidate より後ろに出る場合があることを前提にします。
- raw score、feature set、weight、threshold、suppression formula の具体値はこの contract では固定しません。

#### Cursor Semantics

- `cursor` は resolved order を継続取得するための opaque token です。
- cursor や payload から `publishedAt` keyset、rank key、recommendation reason を推測できることは保証しません。
- pagination は `recommended` / `following` の ordering semantics を維持したまま次ページへ進むために使います。

### `FeedItem`

| field | type | notes |
| --- | --- | --- |
| `short` | `ShortSummary` | feed 上の public short |
| `creator` | `CreatorSummary` | creator block 表示用 |
| `viewer.isPinned` | `boolean` | pin secondary action 表示用 |
| `viewer.isFollowingCreator` | `boolean` | current viewer の creator follow state |
| `unlockCta` | `UnlockCtaState` | 下部固定 CTA 用。price と main 長さはここから組み立てる |

- short pin relation の write contract は `docs/contracts/fan-short-pin-api-contract.md` を参照します。この read surface 自体は auth optional のまま維持します。

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
- short pin relation の write contract は `docs/contracts/fan-short-pin-api-contract.md` を参照します。この read surface 自体は auth optional のまま維持します。

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
- raw score、ranking explanation、recommendation reason、debug field は返さない
- `main` の direct listing は返さない
- recommendation や related creators は返さない

## Fixture Reference

- representative fixture は [fan-public-surfaces.json](fixtures/fan-public-surfaces.json) を参照します。
