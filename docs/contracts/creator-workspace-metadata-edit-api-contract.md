# Creator Workspace Metadata Edit API Contract

## 位置づけ

- この文書は creator owner が owner preview detail から `main price` と `short caption` を更新する mutation contract を固定します。
- owner preview の read surface は `docs/contracts/creator-workspace-owner-preview-api-contract.md` を正とし、この文書は metadata update boundary だけを扱います。
- review status、publish state、delivery asset、analytics はこの leaf に含めません。

## Goals

- creator owner が自分の `main price` を owner preview detail から更新できるようにする。
- creator owner が自分の `short caption` を owner preview detail から更新できるようにする。
- update success を `204 No Content` に固定し、frontend が既存 detail state を局所更新できるようにする。

## Non-goals

- owner preview detail read payload の拡張
- review requested / approved など publish workflow state の更新
- main title、short title、thumbnail、media asset の編集
- analytics、sales、moderation 情報の返却

## Canonical Sources

- `docs/contracts/creator-workspace-owner-preview-api-contract.md`
- `docs/contracts/creator-upload-api-contract.md`
- `docs/contracts/creator-workspace-api-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `PUT` | `/api/creator/workspace/mains/{mainId}/price` | required | owner 自身の main price を更新 |
| `PUT` | `/api/creator/workspace/shorts/{shortId}/caption` | required | owner 自身の short caption を更新 |

## Shared Rules

- caller は authenticated viewer である必要があります。
- caller は approved creator capability を持つ必要があります。
- caller は更新対象の owner 自身である必要があります。
- success response は両 endpoint とも `204 No Content` です。
- success body は返しません。frontend は mutation 前に持っている preview detail state を更新します。
- `main price` は JPY integer で扱います。
- `short caption` は trim 後に空文字なら `null` として保存します。
- review status、publish visibility、delivery-ready 判定は mutation 前後で変えません。

## Request / Response Contract

### `PUT /api/creator/workspace/mains/{mainId}/price`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Request Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `priceJpy` | `number` | yes | `1` 以上の JPY integer |

#### Success Rules

- `204 No Content`
- request body に unknown field を含めません。
- `priceJpy <= 0` は `400 validation_error` とします。

### `PUT /api/creator/workspace/shorts/{shortId}/caption`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Request Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `caption` | `string \| null` | yes | trim 後 blank は `null` に正規化 |

#### Success Rules

- `204 No Content`
- request body に unknown field を含めません。
- `caption = null` または trim 後 blank string は cleared caption として扱います。

## Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | JSON decode 失敗、unknown field、複数 JSON payload |
| `400` | `validation_error` | `priceJpy` が 1 以上の整数でない |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | owner 自身の preview target を解決できない |
| `500` | `internal_error` | unexpected failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_creator_workspace_main_price_invalid_001",
    "page": null
  },
  "error": {
    "code": "validation_error",
    "message": "priceJpy must be positive"
  }
}
```

## Boundary Guardrails

- mutation response に preview detail payload を重ねて返しません。
- creator public profile や fan main playback 契約を兼用しません。
- `main price` 更新で currency code や billing model を拡張しません。
- `short caption` 更新で hashtag / mention parsing の仕様を追加しません。

## Fixture Reference

- representative fixture は [creator-workspace-metadata-edit.json](fixtures/creator-workspace-metadata-edit.json) を参照します。
