# Fan Creator Follow API Contract

## 位置づけ

- この文書は `SHO-113 creator follow mutation の contract と fixture を定義する` の成果物です。
- public な creator profile から実行する `follow / unfollow` mutation contract を固定します。
- creator profile の read contract は引き続き `docs/contracts/fan-public-surface-api-contract.md` を正とし、この文書は relation 更新 boundary だけを扱います。

## Goals

- authenticated fan が public creator profile を auth 必須にせず `follow / unfollow` を実行できるようにする。
- mutation success だけで creator profile header の `viewer.isFollowing` と `stats.fanCount` を整合更新できるようにする。
- repeat request でも follow relation と count が壊れない idempotent な扱いを固定する。

## Non-goals

- creator search result / feed / short detail からの follow CTA
- fan profile following list や following feed の read contract
- notification、recommendation、creator analytics
- login entry への具体的な UI 遷移設計

## Canonical Sources

- `docs/contracts/fan-public-surface-api-contract.md`
- `docs/contracts/fan-profile-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/fan/fan-journey.md`
- `docs/ssot/product/ui/fan-surfaces.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `PUT` | `/api/fan/creators/{creatorId}/follow` | required | follow relation を存在する状態へ収束させる |
| `DELETE` | `/api/fan/creators/{creatorId}/follow` | required | follow relation を存在しない状態へ収束させる |

## Shared Rules

- path の `creatorId` は `GET /api/fan/creators/{creatorId}` と同じ public creator identifier を使います。
- request body は両 endpoint とも持ちません。
- success response の `meta.page` は常に `null` です。
- success response は action 後の post-condition だけを返します。`creator` summary、short grid、following list、`currentViewer` は返しません。
- success response の `data` は `viewer.isFollowing` と `stats.fanCount` だけを返します。
- `data.stats.fanCount` は action 適用後の count です。repeat request でも drift しません。
- `fan` と `creator` は別 identity に分けず、follow relation は fan relation として扱います。

## Request / Response Contract

### Success Payload

- success body は次の shape に固定します。

```json
{
  "data": {
    "viewer": {
      "isFollowing": true
    },
    "stats": {
      "fanCount": 24001
    }
  },
  "meta": {
    "requestId": "req_creator_follow_put_001",
    "page": null
  },
  "error": null
}
```

### `PUT /api/fan/creators/{creatorId}/follow`

#### Path

| field | type | required |
| --- | --- | --- |
| `creatorId` | `string` | yes |

#### Success Rules

- `200`
- `data.viewer.isFollowing = true`
- `data.stats.fanCount` は follow 後の fan count
- 既に follow 済みの creator へ再 `PUT` しても `200` success とし、count は増やしません

### `DELETE /api/fan/creators/{creatorId}/follow`

#### Path

| field | type | required |
| --- | --- | --- |
| `creatorId` | `string` | yes |

#### Success Rules

- `200`
- `data.viewer.isFollowing = false`
- `data.stats.fanCount` は unfollow 後の fan count
- 既に unfollow 済みの creator へ再 `DELETE` しても `200` success とし、count は減らしません

## HTTP States

| endpoint | success | auth_required | not_found | internal_error |
| --- | --- | --- | --- | --- |
| `PUT /api/fan/creators/{creatorId}/follow` | `200` | `401` | `404` | `500` |
| `DELETE /api/fan/creators/{creatorId}/follow` | `200` | `401` | `404` | `500` |

## Error Contract

| status | code | meaning |
| --- | --- | --- |
| `401` | `auth_required` | follow relation update に authenticated fan session が必要 |
| `404` | `not_found` | target creator が存在しない、または public creator profile を持たない |
| `500` | `internal_error` | 想定外の server failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_creator_follow_error_001",
    "page": null
  },
  "error": {
    "code": "auth_required",
    "message": "creator follow requires authentication"
  }
}
```

## Boundary Guardrails

- public creator profile の read contract を auth 必須にしません。
- mutation response に `creator` summary や `stats.shortCount` を重ねて返しません。
- `creator search`、`feed`、`short detail`、`fan profile following` の再読込 contract はこの文書で固定しません。
- login entry への redirect 先や modal 表現は frontend task の責務とし、この文書では `401 auth_required` だけを固定します。

## Fixture Reference

- representative fixture は [fan-creator-follow.json](fixtures/fan-creator-follow.json) を参照します。
