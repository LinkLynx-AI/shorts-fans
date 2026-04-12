# Creator Workspace Main Price API Contract

## 位置づけ

- この文書は creator workspace 内で owner 自身が `main` の価格を変更する private mutation contract を固定します。
- 対象は `priceJpy` の変更だけで、publish state、unlock state、review state、linked short の編集は含めません。
- owner preview read contract とは分けて、更新責務と error vocabulary を独立に固定します。

## Goals

- creator owner が本編詳細から現在価格を変更できる最小 mutation surface を固定する。
- `main` 所有者本人だけが `priceJpy` を更新できることを明示する。
- frontend が modal から即時反映できる最小 success payload を返す。

## Non-goals

- 新規 `main` の作成
- `currencyCode` の変更
- `ownershipConfirmed` / `consentConfirmed` / `approvedForUnlockAt` の変更
- review / moderation / sales 指標の更新
- fan unlock 価格表示以外の UI contract 追加

## Canonical Sources

- `docs/contracts/creator-workspace-owner-preview-api-contract.md`
- `docs/contracts/creator-upload-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `PUT` | `/api/creator/workspace/mains/{mainId}/price` | required | owner 自身の `main` 価格を更新 |

## Request Contract

### `PUT /api/creator/workspace/mains/{mainId}/price`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `priceJpy` | `number` | yes | `1` 以上の整数 |

#### Response

- `data.main.id`: `string`
- `data.main.priceJpy`: `number`
- `meta.page = null`

## Response Rules

- caller は authenticated viewer である必要があります。
- caller は approved creator capability を持つ owner 自身である必要があります。
- success payload は変更後の `priceJpy` だけを返し、preview detail 全体は返しません。
- mutation は `main` の price だけを更新し、state / review / consent / ownership / unlock readiness を変更しません。
- `currencyCode` は `JPY` 前提とし、mutation では変更しません。

## Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON、unknown field、または request body shape が不正 |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | owner 自身の更新対象 `main` を解決できない |
| `422` | `validation_error` | `priceJpy` が `1` 未満、または整数条件を満たさない |
| `500` | `internal_error` | unexpected failure |

## Guardrails

- `currencyCode` を body で受け取りません。
- linked short、thumbnail、caption、visibility は変更しません。
- response に review reason、processing state、sales metrics を含めません。
- owner preview read contract を mutation response で置き換えません。

## Fixture Reference

- representative fixture は [creator-workspace-main-price.json](fixtures/creator-workspace-main-price.json) を参照します。
