# Fan Short Like API Contract

## 位置づけ

- この文書は short に対する最小の `like / unlike` mutation contract を固定します。
- public short の read contract は引き続き `docs/contracts/fan-public-surface-api-contract.md` を正とし、この文書は viewer-private like relation 更新 boundary だけを扱います。

## Goals

- authenticated fan が public short read surface を auth 必須にせず `like / unlike` を実行できるようにする。
- mutation success だけで current surface の `viewer.hasLiked` と `engagement.likeCount` を整合更新できるようにする。
- repeat request でも like relation と count が壊れない idempotent な扱いを固定する。

## Non-goals

- `main` への like
- comment / reaction / notification
- fan profile liked-shorts read endpoint
- recommendation ranking への like signal 反映
- creator analytics や revenue surface への social signal 追加

## Canonical Sources

- `docs/contracts/fan-auth-modal-ui-contract.md`
- `docs/contracts/fan-public-surface-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/fan/fan-profile-and-engagement.md`
- `docs/ssot/product/ui/fan-surfaces.md`
- `docs/ssot/product/fan/consumer-state-and-profile.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `PUT` | `/api/fan/shorts/{shortId}/like` | required | like relation を存在する状態へ収束させる |
| `DELETE` | `/api/fan/shorts/{shortId}/like` | required | like relation を存在しない状態へ収束させる |

## Shared Rules

- path の `shortId` は `GET /api/fan/shorts/{shortId}` と同じ public short identifier を使います。
- request body は両 endpoint とも持ちません。
- success response の `meta.page` は常に `null` です。
- success response は action 後の post-condition だけを返します。`short` summary、creator summary、fan profile list、`currentViewer` は返しません。
- success response の `data` は `viewer.hasLiked` と `engagement.likeCount` だけを返します。
- `hasLiked` は viewer-private state、`likeCount` は public short の social signal として扱います。
- like は short だけを対象にし、canonical `main` に like relation を作りません。

## Request / Response Contract

### Success Payload

```json
{
  "data": {
    "viewer": {
      "hasLiked": true
    },
    "engagement": {
      "likeCount": 42
    }
  },
  "meta": {
    "requestId": "req_short_like_put_001",
    "page": null
  },
  "error": null
}
```

### `PUT /api/fan/shorts/{shortId}/like`

#### Success Rules

- `200`
- `data.viewer.hasLiked = true`
- `data.engagement.likeCount` は like 後の total count
- 既に like 済みの short へ再 `PUT` しても `200` success とし、count は増やしません

### `DELETE /api/fan/shorts/{shortId}/like`

#### Success Rules

- `200`
- `data.viewer.hasLiked = false`
- `data.engagement.likeCount` は unlike 後の total count
- 既に unlike 済みの short へ再 `DELETE` しても `200` success とし、count は減らしません

## HTTP States

| endpoint | success | auth_required | not_found | internal_error |
| --- | --- | --- | --- | --- |
| `PUT /api/fan/shorts/{shortId}/like` | `200` | `401` | `404` | `500` |
| `DELETE /api/fan/shorts/{shortId}/like` | `200` | `401` | `404` | `500` |

## Error Contract

| status | code | meaning |
| --- | --- | --- |
| `401` | `auth_required` | short like mutation に authenticated fan session が必要 |
| `404` | `not_found` | target short が存在しない、または public short surface に出せない |
| `500` | `internal_error` | 想定外の server failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_short_like_error_001",
    "page": null
  },
  "error": {
    "code": "auth_required",
    "message": "short like requires authentication"
  }
}
```

## Boundary Guardrails

- public short read contract 自体を auth 必須にしません。
- mutation response に `short` summary や creator summary を重ねて返しません。
- `main` playback surface には like state や like action を出しません。
- `401 auth_required` を受けた frontend は `docs/contracts/fan-auth-modal-ui-contract.md` の shared fan auth modal を primary entry とします。

## Fixture Reference

- representative fixture は [fan-short-like.json](fixtures/fan-short-like.json) を参照します。
