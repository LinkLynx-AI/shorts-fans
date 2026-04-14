# Fan Short Pin API Contract

## 位置づけ

- この文書は `SHO-161 short pin mutation の contract と fixture を定義する` の成果物です。
- `feed` 上の `pin / unpin` mutation contract を固定します。
- public short の read contract は引き続き `docs/contracts/fan-public-surface-api-contract.md` を正とし、この文書は viewer-private relation 更新 boundary だけを扱います。

## Goals

- authenticated fan が public short read surface を auth 必須にせず `pin / unpin` を実行できるようにする。
- mutation success だけで current surface の `viewer.isPinned` を整合更新できるようにする。
- repeat request でも pin relation が壊れない idempotent な扱いを固定する。

## Non-goals

- `feed` 以外の UI wiring
- `GET /api/fan/profile/pinned-shorts` read endpoint 実装
- public count や social signal の追加
- shared fan auth modal を超える login entry の具体的な visual 実装

## Canonical Sources

- `docs/contracts/fan-auth-modal-ui-contract.md`
- `docs/contracts/fan-public-surface-api-contract.md`
- `docs/contracts/fan-profile-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/fan/fan-profile-and-engagement.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/fan/consumer-state-and-profile.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `PUT` | `/api/fan/shorts/{shortId}/pin` | required | pin relation を存在する状態へ収束させる |
| `DELETE` | `/api/fan/shorts/{shortId}/pin` | required | pin relation を存在しない状態へ収束させる |

## Shared Rules

- path の `shortId` は `GET /api/fan/shorts/{shortId}` と同じ public short identifier を使います。
- request body は両 endpoint とも持ちません。
- success response の `meta.page` は常に `null` です。
- success response は action 後の post-condition だけを返します。`short` summary、creator summary、fan profile list、`currentViewer` は返しません。
- success response の `data` は `viewer.isPinned` だけを返します。
- `pin` は viewer-private state であり、public social signal に昇格させません。

## Request / Response Contract

### Success Payload

- success body は次の shape に固定します。

```json
{
  "data": {
    "viewer": {
      "isPinned": true
    }
  },
  "meta": {
    "requestId": "req_short_pin_put_001",
    "page": null
  },
  "error": null
}
```

### `PUT /api/fan/shorts/{shortId}/pin`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Success Rules

- `200`
- `data.viewer.isPinned = true`
- 既に pin 済みの short へ再 `PUT` しても `200` success とし、relation は維持したまま壊しません

### `DELETE /api/fan/shorts/{shortId}/pin`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Success Rules

- `200`
- `data.viewer.isPinned = false`
- 既に unpin 済みの short へ再 `DELETE` しても `200` success とし、relation は壊しません

## HTTP States

| endpoint | success | auth_required | not_found | internal_error |
| --- | --- | --- | --- | --- |
| `PUT /api/fan/shorts/{shortId}/pin` | `200` | `401` | `404` | `500` |
| `DELETE /api/fan/shorts/{shortId}/pin` | `200` | `401` | `404` | `500` |

## Error Contract

| status | code | meaning |
| --- | --- | --- |
| `401` | `auth_required` | short pin mutation に authenticated fan session が必要 |
| `404` | `not_found` | target short が存在しない、または public short surface に出せない |
| `500` | `internal_error` | 想定外の server failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_short_pin_error_001",
    "page": null
  },
  "error": {
    "code": "auth_required",
    "message": "short pin requires authentication"
  }
}
```

## Boundary Guardrails

- public short read contract 自体を auth 必須にしません。
- mutation response に `short` summary や creator summary を重ねて返しません。
- `feed`、`short detail`、`fan profile pinned-shorts` の再読込 contract はこの文書で固定しません。
- `401 auth_required` を受けた frontend は `docs/contracts/fan-auth-modal-ui-contract.md` の shared fan auth modal を primary entry とします。

## Fixture Reference

- representative fixture は [fan-short-pin.json](fixtures/fan-short-pin.json) を参照します。
