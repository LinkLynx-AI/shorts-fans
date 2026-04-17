# Fan Unlock Purchase And Main API Contract

## 位置づけ

- この文書は `SHO-206 card 決済前提の unlock purchase contract / fixture を更新する` の成果物です。
- `short -> paywall -> main` の purchase / access boundary を、card-only actual purchase 前提で固定します。
- canonical `main` の durable purchase と current session に閉じた playback grant を分離し、`purchase -> access-entry -> playback` の順で扱います。
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
| `GET` | `/api/fan/shorts/{shortId}/unlock` | required | paywall を開く前の purchase / access state 読み出し |
| `POST` | `/api/fan/mains/{mainId}/purchase` | required | card-only one-time purchase を実行し、成功時だけ durable purchase を記録する |
| `POST` | `/api/fan/mains/{mainId}/access-entry` | required | durable purchase または owner access を検証し、short-lived playback grant を発行する |
| `GET` | `/api/fan/mains/{mainId}/playback` | required | grant 済み main player 開始用 |

## Surface-specific Payloads

### `EntryContext`

| field | type | notes |
| --- | --- | --- |
| `purchasePath` | `string` | purchase mutation path |
| `accessEntryPath` | `string` | playback grant 発行 path |
| `token` | `string` | viewer / short / main に束縛された opaque token |

### `SavedCardSummary`

| field | type | notes |
| --- | --- | --- |
| `paymentMethodId` | `string` | app / provider 境界の opaque saved-card identifier |
| `brand` | `"visa" \| "mastercard" \| "jcb" \| "american_express"` | card brand |
| `last4` | `string` | 下4桁だけを返す |

### `PurchaseSetupState`

| field | type | notes |
| --- | --- | --- |
| `required` | `boolean` | paywall で追加 setup が必要か |
| `requiresCardSetup` | `boolean` | saved card がなく new card setup が必要か |
| `requiresAgeConfirmation` | `boolean` | 年齢確認が必要か |
| `requiresTermsAcceptance` | `boolean` | 利用規約同意が必要か |

### `UnlockPurchaseState`

| field | type | notes |
| --- | --- | --- |
| `state` | `"setup_required" \| "purchase_ready" \| "purchase_pending" \| "already_purchased" \| "owner_preview" \| "unavailable"` | paywall 内の primary state |
| `supportedCardBrands` | `("visa" \| "mastercard" \| "jcb" \| "american_express")[]` | MVP ではこの 4 ブランドだけを返す |
| `savedPaymentMethods` | `SavedCardSummary[]` | saved card 一覧。non-card method は返さない |
| `setup` | `PurchaseSetupState` | purchase 前提条件 |
| `pendingReason` | `"provider_processing" \| null` | `purchase_pending` のときだけ non-null |

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
- `data.entryContext`: `EntryContext`
- `data.access`: `MainAccessState`
- `data.unlockCta`: `UnlockCtaState`
- `data.purchase`: `UnlockPurchaseState`

#### Interpretation Rules

- `unlockCta` は feed / short detail が共通に使う CTA display state として維持し、payment 固有の分岐は `purchase.state` で判断します。
- `unlockCta.state = "setup_required"` のとき、frontend は paywall を開き、`purchase.state = "setup_required"` に従って card setup / consent UI を出します。
- `unlockCta.state = "unlock_available"` のとき、frontend は paywall を開き、`purchase.state = "purchase_ready"` なら purchase を実行できます。
- `unlockCta.state = "continue_main"` のとき、viewer はすでに canonical `main` の durable purchase を持つ前提で、paywall 再購入ではなく main 再開導線を出します。
- `unlockCta.state = "owner_preview"` のとき、purchase 導線ではなく creator owner preview 導線を出します。
- `entryContext.token` は purchase 成功を表しません。`POST /purchase` と `POST /access-entry` が同じ short / main 文脈を検証するための input として使います。
- `supportedCardBrands` は MVP では常に `visa`、`mastercard`、`jcb`、`american_express` の 4 件に固定します。

#### HTTP States

| case | status |
| --- | --- |
| initial setup required | `200` |
| purchase available | `200` |
| purchase pending | `200` |
| already purchased | `200` |
| `owner` | `200` |
| `main_not_unlockable` | `403` + `main_locked` |
| `not_found` | `404` + `not_found` |

### `POST /api/fan/mains/{mainId}/purchase`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `fromShortId` | `string` | yes | entry context を作った short |
| `entryToken` | `string` | yes | `GET /api/fan/shorts/{shortId}/unlock` で返した opaque token |
| `acceptedAge` | `boolean` | yes | `purchase.setup.requiresAgeConfirmation = true` のとき true 必須 |
| `acceptedTerms` | `boolean` | yes | `purchase.setup.requiresTermsAcceptance = true` のとき true 必須 |
| `paymentMethod.mode` | `"saved_card" \| "new_card"` | yes | non-card method は受け付けない |
| `paymentMethod.paymentMethodId` | `string` | no | `saved_card` のとき必須 |
| `paymentMethod.cardSetupToken` | `string` | no | `new_card` のとき必須。provider-neutral な opaque token |

#### Response

- `data.access`: `MainAccessState`
- `data.purchase.status`: `"succeeded" \| "pending" \| "failed" \| "already_purchased" \| "owner_preview"`
- `data.purchase.failureReason`: `"card_brand_unsupported" \| "purchase_declined" \| "authentication_failed" \| null`
- `data.purchase.canRetry`: `boolean`
- `data.entryContext`: `EntryContext \| null`
- `meta.page = null`

#### Interpretation Rules

- purchase success のときだけ canonical `main` の durable purchase を記録します。
- `data.purchase.status = "succeeded"` のとき、frontend は `data.entryContext.accessEntryPath` を使って `POST /access-entry` に進みます。
- `data.purchase.status = "pending"` のとき、unlock はまだ付与しません。frontend は pending UI を出し、unlock surface 再読込または provider 完了待ちへ入ります。
- `data.purchase.status = "failed"` のとき、unlock は付与しません。`failureReason` は public contract で扱う最小語彙だけに留めます。
- `data.purchase.status = "already_purchased"` のとき、追加 billing は行わず、以後は再開導線を使います。
- `data.purchase.status = "owner_preview"` のとき、purchase ではなく owner access を使います。
- provider の request payload、issuer code、3DS provider detail、payment provider purchase ref は response に含めません。

#### HTTP States

| case | status |
| --- | --- |
| purchase succeeded | `200` |
| purchase failed | `200` |
| already purchased | `200` |
| owner preview | `200` |
| purchase pending | `202` |
| invalid payload | `400` |
| unauthenticated | `401` + `auth_required` |
| `main_not_unlockable` / invalid entry context | `403` + `main_locked` |
| linked short or main not found | `404` + `not_found` |

### `POST /api/fan/mains/{mainId}/access-entry`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `fromShortId` | `string` | yes | entry context を作った short |
| `entryToken` | `string` | yes | `GET /api/fan/shorts/{shortId}/unlock` で返した opaque token |

#### Response

- `data.href`: `string`
- `data.href` は `/mains/{mainId}?fromShortId=...&grant=...` のような app route を返します。
- `meta.page = null`

#### Interpretation Rules

- この endpoint は purchase を実行しません。durable purchase または owner access を検証したうえで short-lived な playback grant だけを発行します。
- `grant` は current session に閉じた temporary proof として扱いますが、purchase record や billing ledger と同義にしません。
- `fromShortId` が `mainId` に連結されない、token が無効、または caller が durable purchase / owner access を持たない場合は playback grant を発行しません。
- 同じ canonical `main` への再進入は idempotent に扱って構いませんが、idempotency は purchase 記録とは別境界です。

#### HTTP States

| case | status |
| --- | --- |
| entry issued | `200` |
| invalid payload | `400` |
| unauthenticated | `401` + `auth_required` |
| purchase required / invalid token / main unavailable | `403` + `main_locked` |
| linked short or main not found | `404` + `not_found` |

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
| `purchased` playback | `200` |
| `owner` playback | `200` |
| purchase required / invalid grant | `403` + `main_locked` |
| `not_found` | `404` + `not_found` |

## State Matrix

| endpoint | not purchased | purchase pending | purchased | owner | not_found |
| --- | --- | --- | --- | --- | --- |
| `GET /api/fan/shorts/{shortId}/unlock` | yes | yes | yes | yes | yes |
| `POST /api/fan/mains/{mainId}/purchase` | yes | yes | yes | yes | yes |
| `POST /api/fan/mains/{mainId}/access-entry` | yes | no | yes | yes | yes |
| `GET /api/fan/mains/{mainId}/playback` | yes | no | yes | yes | yes |

## Out-of-scope Guardrails

- payment provider 固有 payload、SDK 要件、issuer detail は返さない
- `BitCash`、`Paidy`、`atone`、`コンビニ / 銀行ATM` など card 以外の支払い方法は含めない
- `subscription` や bundle 由来 access は返さない
- ledger 詳細、refund / chargeback、creator payout 情報はこの文書に含めない
- explicit preview、thumbnail gallery、related content は返さない

## Fixture Reference

- representative fixture は [fan-unlock-main.json](fixtures/fan-unlock-main.json) を参照します。
