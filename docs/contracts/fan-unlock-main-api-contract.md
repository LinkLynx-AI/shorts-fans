# Fan Unlock And Main API Contract

## 位置づけ

- この文書は `SHO-18 unlock / main player の API 契約とモックデータを定義する` の成果物です。
- `short -> unlock -> main` の display/access contract を固定し、初回 setup 要否、unlock 済み access、owner preview の差分を揃えます。
- この leaf では payment provider 決済は行わず、`access-entry` 成功時に canonical `main` の unlock access を永続記録します。
- DTO と response envelope は `docs/contracts/fan-mvp-common-transport-contract.md` を参照します。

## Canonical Sources

- `docs/contracts/media-display-access-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/ssot/product/monetization/billing-and-access.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/fan/fan-journey.md`
- `docs/contracts/mvp-core-domain-contract.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/fan/shorts/{shortId}/unlock` | required | main access entry 前の setup / unlock state 読み出し |
| `POST` | `/api/fan/mains/{mainId}/access-entry` | required | unlock 記録後に short-lived な main entry href を発行する |
| `GET` | `/api/fan/mains/{mainId}/playback` | required | grant 済み main player 開始用 |

## Request Contract

### `GET /api/fan/shorts/{shortId}/unlock`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Response

- `data.short`: `ShortSummary`
- `data.creator`: `CreatorSummary`
- `data.main.id`: `string`
- `data.main.durationSeconds`: `number`
- `data.main.priceJpy`: `number`
- `data.mainAccessEntry.routePath`: `string`
- `data.mainAccessEntry.token`: `string`
- `data.access`: `MainAccessState`
- `data.unlockCta`: `UnlockCtaState`
- `data.setup.required`: `boolean`
- `data.setup.requiresAgeConfirmation`: `boolean`
- `data.setup.requiresTermsAcceptance`: `boolean`

#### Interpretation Rules

- `setup.required = true` のときだけ unlock setup dialog を出します。
- `unlockCta.state = "setup_required"` のとき、frontend は `Unlock` CTA から setup dialog を開きます。
- `unlockCta.state = "unlock_available"` のとき、frontend は billing を挟まず `POST /api/fan/mains/{mainId}/access-entry` に進めます。
- `unlockCta.state = "continue_main"` のとき、frontend は viewer がすでに canonical `main` access を持つ前提で main 再開導線を出します。
- `unlockCta.state = "owner_preview"` のとき、purchase 導線ではなく creator owner preview 導線を出します。
- `priceJpy` は reference price 表示用に残しますが、この leaf では決済実行を意味しません。

#### HTTP States

| case | status |
| --- | --- |
| initial setup required | `200` |
| setup complete, unlock available | `200` |
| `unlocked` | `200` |
| `owner` | `200` |
| `locked` | `403` + `main_locked` |
| `not_found` | `404` + `not_found` |

### `POST /api/fan/mains/{mainId}/access-entry`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `fromShortId` | `string` | yes | entry context を作った short |
| `entryToken` | `string` | yes | `GET /api/fan/shorts/{shortId}/unlock` で返した signed token |
| `acceptedAge` | `boolean` | yes | setup で年齢確認が必要な時だけ true 必須 |
| `acceptedTerms` | `boolean` | yes | setup で利用規約同意が必要な時だけ true 必須 |

#### Response

- `data.href`: `string`
- `data.href` は `/mains/{mainId}?fromShortId=...&grant=...` のような app route を返します。
- `meta.page = null`

#### Interpretation Rules

- この endpoint は payment provider 決済を行わず、setup 条件と access entry token を検証したうえで canonical `main` の unlock access を記録し、short-lived な main entry を発行します。
- unlock 記録は `main` 単位で行い、同じ canonical `main` に対する再進入は idempotent に扱います。
- `grant` は current session に閉じた temporary proof として扱いますが、grant 自体は purchase record や billing ledger と同義にしません。
- `fromShortId` が `mainId` に連結されない、または token が無効な場合は access entry を発行しません。

#### HTTP States

| case | status |
| --- | --- |
| entry issued | `200` |
| invalid payload | `400` |
| unauthenticated | `401` + `auth_required` |
| setup incomplete / invalid token | `403` + `main_locked` |
| linked short not found | `404` + `not_found` |

### `GET /api/fan/mains/{mainId}/playback`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Query

| field | type | required | notes |
| --- | --- | --- | --- |
| `fromShortId` | `string` | yes | short 起点で入るときの context 保持用 |
| `grant` | `string` | yes | `POST /api/fan/mains/{mainId}/access-entry` が発行した short-lived grant |

#### Response

- `data.main.id`: `string`
- `data.main.media`: `VideoDisplayAsset`
- `data.main.durationSeconds`: `number`
- `data.creator`: `CreatorSummary`
- `data.access`: `MainAccessState`
- `data.resumePositionSeconds`: `number \| null`
- `data.entryShort`: `ShortSummary \| null`

#### HTTP States

| case | status |
| --- | --- |
| `unlocked` playback | `200` |
| `owner` playback | `200` |
| `locked` | `403` + `main_locked` |
| `not_found` | `404` + `not_found` |

## State Matrix

| endpoint | locked | unlocked | owner | not_found |
| --- | --- | --- | --- | --- |
| `GET /api/fan/shorts/{shortId}/unlock` | yes | yes | yes | yes |
| `POST /api/fan/mains/{mainId}/access-entry` | yes | yes | yes | yes |
| `GET /api/fan/mains/{mainId}/playback` | yes | yes | yes | yes |

## Out-of-scope Guardrails

- 実課金 mutation はこの文書に含めない
- payment provider 固有 payload は返さない
- `subscription` や bundle 由来 access は返さない
- library item read surface や ledger 詳細 payload はこの leaf に含めない
- explicit preview、thumbnail gallery、related content は返さない

## Fixture Reference

- representative fixture は [fan-unlock-main.json](fixtures/fan-unlock-main.json) を参照します。
