# Fan Unlock And Main API Contract

## 位置づけ

- この文書は `SHO-18 unlock / main player の API 契約とモックデータを定義する` の成果物です。
- `short -> unlock -> main` の read contract を固定し、初回 setup 要否、direct unlock、purchased、owner の差分を揃えます。
- DTO と response envelope は `docs/contracts/fan-mvp-common-transport-contract.md` を参照します。

## Canonical Sources

- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/ssot/product/monetization/billing-and-access.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/fan/fan-journey.md`
- `docs/contracts/mvp-core-domain-contract.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/v1/fan/shorts/{shortId}/unlock` | required | mini paywall / direct unlock 判定用 |
| `GET` | `/api/v1/fan/mains/{mainId}/playback` | required | main player 開始用 |

## Request Contract

### `GET /api/v1/fan/shorts/{shortId}/unlock`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Response

- `data.short`: `ShortSummary`
- `data.creator`: `CreatorSummary`
- `data.main.id`: `string`
- `data.main.title`: `string`
- `data.main.durationSeconds`: `number`
- `data.main.priceJpy`: `number`
- `data.purchase`: `PurchaseState`
- `data.access`: `MainAccessState`
- `data.unlockCta`: `UnlockCtaState`
- `data.setup.required`: `boolean`
- `data.setup.requiresAgeConfirmation`: `boolean`
- `data.setup.requiresTermsAcceptance`: `boolean`

#### Interpretation Rules

- `setup.required = true` のときだけ mini paywall setup を出します。
- `unlockCta.state = "setup_required"` のとき、frontend は `Unlock` CTA から mini paywall を開きます。
- `unlockCta.state = "unlock_available"` のとき、frontend は full paywall を挟まず unlock mutation に進めます。
- `unlockCta.state = "continue_main"` のとき、frontend は再購入ではなく main 再開導線を出します。
- `unlockCta.state = "owner_preview"` のとき、purchase 導線ではなく creator owner preview 導線を出します。

#### HTTP States

| case | status |
| --- | --- |
| initial setup required | `200` |
| setup complete, unlock available | `200` |
| `purchased` | `200` |
| `owner` | `200` |
| `locked` | `403` + `main_locked` |
| `not_found` | `404` + `not_found` |

### `GET /api/v1/fan/mains/{mainId}/playback`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Query

| field | type | required | notes |
| --- | --- | --- | --- |
| `fromShortId` | `string` | no | short 起点で入るときの context 保持用。未指定時は `entryShort = null` |

#### Response

- `data.main.id`: `string`
- `data.main.title`: `string`
- `data.main.media`: `MediaAsset`
- `data.main.durationSeconds`: `number`
- `data.creator`: `CreatorSummary`
- `data.access`: `MainAccessState`
- `data.resumePositionSeconds`: `number \| null`
- `data.entryShort`: `ShortSummary \| null`

#### HTTP States

| case | status |
| --- | --- |
| `purchased` playback | `200` |
| `owner` playback | `200` |
| `locked` | `403` + `main_locked` |
| `not_found` | `404` + `not_found` |

## State Matrix

| endpoint | locked | purchased | owner | not_found |
| --- | --- | --- | --- | --- |
| `GET /api/v1/fan/shorts/{shortId}/unlock` | yes | yes | yes | yes |
| `GET /api/v1/fan/mains/{mainId}/playback` | yes | yes | yes | yes |

## Out-of-scope Guardrails

- unlock mutation はこの文書に含めない
- payment provider 固有 payload は返さない
- `subscription` や bundle 由来 access は返さない
- explicit preview、thumbnail gallery、related content は返さない

## Fixture Reference

- representative fixture は [fan-unlock-main.json](fixtures/fan-unlock-main.json) を参照します。
