# Fan Profile API Contract

## 位置づけ

- この文書は `SHO-19 fan profile private hub の API 契約とモックデータを定義する` の成果物です。
- `Following / Pinned Shorts / Library / Settings` を private consumer hub として読む contract を固定します。
- DTO と response envelope は `docs/contracts/fan-mvp-common-transport-contract.md` を参照します。

## Canonical Sources

- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/ssot/product/fan/consumer-state-and-profile.md`
- `docs/ssot/product/fan/fan-profile-and-engagement.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/account/account-permissions.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/fan/profile` | required | private hub overview |
| `GET` | `/api/fan/profile/following` | required | following creator 一覧 |
| `GET` | `/api/fan/profile/pinned-shorts` | required | pinned short 一覧 |
| `GET` | `/api/fan/profile/library` | required | unlocked main 一覧 |
| `GET` | `/api/fan/profile/settings` | required | minimum read-only settings sections |

## Surface-specific Payloads

### `PinnedShortItem`

| field | type | notes |
| --- | --- | --- |
| `short` | `ShortSummary` | pinned short |
| `creator` | `CreatorSummary` | short owner |

### `LibraryItem`

| field | type | notes |
| --- | --- | --- |
| `main.id` | `string` | canonical main identifier |
| `main.title` | `string` | main title |
| `main.durationSeconds` | `number` | main length |
| `creator` | `CreatorSummary` | owner |
| `entryShort` | `ShortSummary` | main へ戻るときの entry context として使う linked short |
| `access` | `MainAccessState` | `purchased` または `owner` |

### `FollowingItem`

| field | type | notes |
| --- | --- | --- |
| `creator` | `CreatorSummary` | followed creator |
| `viewer.isFollowing` | `true` | private hub 上では true 固定 |

### `SettingsSection`

| field | type | notes |
| --- | --- | --- |
| `key` | `"account" \| "payment" \| "safety" \| "mode_switch"` | settings section identifier |
| `label` | `string` | section label |
| `available` | `boolean` | 現時点で選択可能か |

## Request Contract

### `GET /api/fan/profile`

#### Response

- `data.fanProfile.title`: `string`
- `data.fanProfile.currentMode`: `"fan" \| "creator"`
- `data.fanProfile.counts.following`: `number`
- `data.fanProfile.counts.pinnedShorts`: `number`
- `data.fanProfile.counts.library`: `number`
- `data.preview.pinnedShorts`: `PinnedShortItem[]`
- `data.preview.library`: `LibraryItem[]`

#### Overview Rules

- preview arrays は各 dedicated endpoint と同じ並び順の先頭 `3` 件まで返します。
- overview では following の count は返しますが、creator list 自体は `/following` に分けます。

### `GET /api/fan/profile/following`

#### Query

| field | type | required |
| --- | --- | --- |
| `cursor` | `string` | no |

#### Response

- `data.items`: `FollowingItem[]`
- `meta.page`: `CursorPageInfo`

### `GET /api/fan/profile/pinned-shorts`

#### Query

| field | type | required |
| --- | --- | --- |
| `cursor` | `string` | no |

#### Response

- `data.items`: `PinnedShortItem[]`
- `meta.page`: `CursorPageInfo`

### `GET /api/fan/profile/library`

#### Query

| field | type | required |
| --- | --- | --- |
| `cursor` | `string` | no |

#### Response

- `data.items`: `LibraryItem[]`
- `meta.page`: `CursorPageInfo`

### `GET /api/fan/profile/settings`

#### Response

- `data.currentMode`: `"fan" \| "creator"`
- `data.creatorModeAvailable`: `boolean`
- `data.sections`: `SettingsSection[]`
- `meta.page = null`

## HTTP States

| endpoint | populated | empty | not_found | unauthenticated |
| --- | --- | --- | --- | --- |
| `GET /api/fan/profile` | `200` | `200` | `404` | `401` |
| `GET /api/fan/profile/following` | `200` | `200` | no | `401` |
| `GET /api/fan/profile/pinned-shorts` | `200` | `200` | no | `401` |
| `GET /api/fan/profile/library` | `200` | `200` | no | `401` |
| `GET /api/fan/profile/settings` | `200` | no | no | `401` |

## Out-of-scope Guardrails

- public fan profile は返さない
- `like / comment / public activity` は返さない
- full watch history は返さない
- creator dashboard や creator private analytics は返さない

## Fixture Reference

- representative fixture は [fan-profile.json](fixtures/fan-profile.json) を参照します。
