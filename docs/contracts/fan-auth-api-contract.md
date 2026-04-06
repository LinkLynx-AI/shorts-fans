# Fan Auth API Contract

## 位置づけ

- この文書は `SHO-53 fan 認証 transport と session lifecycle service を実装する` の成果物です。
- fan 側の最小 `sign in / sign up / session start / logout` を backend transport として固定します。
- current viewer の正規 read surface は引き続き `GET /api/viewer/bootstrap` とし、この文書は session mutation boundary だけを扱います。

## Goals

- unauthenticated viewer が `sign in` または `sign up` を経て authenticated session に入れるようにする。
- login 成功時に `shorts_fans_session` cookie を発行し、bootstrap が current viewer を継続して読めるようにする。
- logout 時に session revoke と cookie clear を idempotent に扱えるようにする。

## Non-goals

- mail / SMS / external provider への challenge delivery
- MFA / password recovery / social login
- active mode switch
- frontend auth UI

## Canonical Sources

- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/account/identity-and-mode-model.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `POST` | `/api/fan/auth/sign-in/challenges` | public | 既存 email identity 向け challenge 発行 |
| `POST` | `/api/fan/auth/sign-in/session` | public | challenge を消費して session cookie を発行 |
| `POST` | `/api/fan/auth/sign-up/challenges` | public | 未登録 email 向け challenge 発行 |
| `POST` | `/api/fan/auth/sign-up/session` | public | challenge を消費して `user + auth_identity + session` を作成 |
| `DELETE` | `/api/fan/auth/session` | optional | session revoke と cookie clear。cookie が無くても成功扱い |

## Shared Rules

- この issue では `provider = email` だけを扱います。
- `fan` と `creator` を別 identity に分けません。新規 session の `activeMode` は常に `fan` で開始します。
- challenge token は DB に raw 値を保存せず hash だけ保持します。
- challenge delivery 実装は scope 外のため、この issue では challenge 発行 response が raw token を直接返します。
- login / logout 成功 body に viewer state は返しません。成功後の current viewer は `GET /api/viewer/bootstrap` から読みます。

## Request / Response Contract

### Challenge Issue Request

- request body は次の shape に固定します。

```json
{
  "email": "fan@example.com"
}
```

### Challenge Issue Success

- `200`
- `data.challengeToken` は raw challenge token です。
- `data.expiresAt` は RFC3339 文字列です。

```json
{
  "data": {
    "challengeToken": "challenge_token_001",
    "expiresAt": "2026-04-06T19:00:00Z"
  },
  "meta": {
    "requestId": "req_fan_auth_sign_in_challenge_001",
    "page": null
  },
  "error": null
}
```

### Session Start Request

- request body は次の shape に固定します。

```json
{
  "email": "fan@example.com",
  "challengeToken": "challenge_token_001"
}
```

### Session Start Success

- `204`
- response body は返しません。
- `Set-Cookie` で `shorts_fans_session` を発行します。

## Logout Success

- `204`
- response body は返しません。
- cookie の有無や session の有効性に関わらず clear cookie を返します。

## Cookie Rules

- cookie 名は `shorts_fans_session` です。
- `HttpOnly = true`
- `Path = /`
- `SameSite = Lax`
- `Secure = production のみ true`
- domain は固定しません。
- expiry は session 発行時点から 30 日固定です。

## Error Contract

| status | code | meaning |
| --- | --- | --- |
| `400` | `invalid_email` | email 形式が不正、または email request body が不正 |
| `400` | `invalid_challenge` | challenge token が不正、期限切れ、消費済み、または session request body が不正 |
| `404` | `email_not_found` | sign-in 対象の email identity が存在しない |
| `409` | `email_already_registered` | sign-up 対象の email が既に登録済み |
| `500` | `internal_error` | 想定外の server failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_fan_auth_error_001",
    "page": null
  },
  "error": {
    "code": "invalid_challenge",
    "message": "challenge is invalid"
  }
}
```

## Boundary Guardrails

- `currentViewer`、`activeMode`、`canAccessCreatorMode` は session mutation response に重ねて返しません。
- `fan / creator` を別 login identity に分けません。
- provider 固有 metadata や delivery state は返しません。
- `lastSeenAt`、`expiresAt`、`sessionTokenHash` のような session internals は body に返しません。

## Fixture Reference

- representative fixture は [fan-auth.json](fixtures/fan-auth.json) を参照します。
