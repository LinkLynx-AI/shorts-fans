# Viewer Bootstrap API Contract

## 位置づけ

- この文書は `SHO-39 current viewer bootstrap を実装する` の成果物です。
- app shell が protected flow に入る前に参照する `current viewer bootstrap` の read 契約を固定します。
- fan MVP surface payload に viewer 自身の identity / active mode を混在させず、app bootstrap 側で保持する前提を concrete な wire shape にします。

## Goals

- authenticated viewer の最小 current state を app bootstrap で参照できるようにする。
- unauthenticated viewer でも bootstrap 自体は成功させ、protected state だけを返さないようにする。
- `1 user identity + optional creator capability + active mode switch` を bootstrap payload に反映する。

## Non-goals

- login / logout / session 発行
- protected fan endpoint の `auth_required` 判定
- fan profile / feed / unlock の surface payload
- creator dashboard bootstrap

## Canonical Sources

- `docs/contracts/fan-auth-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/fan-mvp-fixtures-and-integration-guide.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/account/identity-and-mode-model.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/viewer/bootstrap` | optional | app shell global self state |

## Request Boundary

- bootstrap は `shorts_fans_session` cookie から current viewer を解決します。
- `shorts_fans_session` は Cognito-backed fan auth 完了後に app backend が発行する cookie です。bootstrap は Cognito access token / ID token を直接読みません。
- cookie がない、期限切れ、revoked、lookup 不能のいずれでも `unauthenticated success` として扱います。
- bootstrap 自体は `auth_required` を返しません。

## Response Contract

### `CurrentViewer`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | root user identity |
| `activeMode` | `"fan" \| "creator"` | current session で前面に出す mode |
| `canAccessCreatorMode` | `boolean` | approved creator capability の有無 |

### Success Envelope

- `data.currentViewer` は authenticated viewer では object、unauthenticated viewer では `null` です。
- `meta.page = null` 固定です。
- `error = null` 固定です。

```json
{
  "data": {
    "currentViewer": {
      "id": "user_123",
      "activeMode": "fan",
      "canAccessCreatorMode": false
    }
  },
  "meta": {
    "requestId": "req_viewer_bootstrap_001",
    "page": null
  },
  "error": null
}
```

## Response Rules

- unauthenticated viewer は `200` で `data.currentViewer = null` を返します。
- authenticated viewer は `200` で `data.currentViewer` を返します。
- `activeMode = creator` でも approved creator capability がない場合、bootstrap 応答上は `activeMode = fan` に正規化します。
- `canAccessCreatorMode` は approved creator capability がある場合だけ `true` です。
- fan surface の relation state や counts は返しません。

## Error Contract

- 想定外の server failure のみ `500` を返します。
- error code は `internal_error` を使います。

```json
{
  "data": null,
  "meta": {
    "requestId": "req_viewer_bootstrap_500",
    "page": null
  },
  "error": {
    "code": "internal_error",
    "message": "viewer bootstrap could not be loaded"
  }
}
```

## Boundary Guardrails

- `profile`、`settings`、`payment method`、`follow counts` などの viewer-private detail は返しません。
- `lastSeenAt`、`expiresAt`、`sessionTokenHash` のような session internals は返しません。
- `emailVerified`、`recentAuthAt`、Cognito provider state のような auth provider detail は返しません。
- `fan / creator` を別 login identity として扱いません。

## Fixture Reference

- representative fixture は [viewer-bootstrap.json](fixtures/viewer-bootstrap.json) を参照します。
