# Fan Auth API Contract

## 位置づけ

- この文書は `SHO-166 Cognito 前提で fan auth 契約と境界を更新する` の成果物です。
- Cognito を認証基盤にした fan 側の `sign in / sign up / sign up confirm / password reset / logout / fresh re-auth` transport を固定します。
- current viewer の正規 read surface は引き続き `GET /api/viewer/bootstrap` とし、この文書は auth mutation boundary だけを扱います。
- modal 主導線の UI state と recovery 責務は `docs/contracts/fan-auth-modal-ui-contract.md` を正とします。

## Goals

- unauthenticated viewer が custom modal UI から `email + password` の fan auth を完了できるようにする。
- sign-up flow の完了時点で shared viewer profile の `displayName / handle / avatar` を初期化できるようにする。
- sign in と sign up confirm 成功時に `shorts_fans_session` cookie を発行し、bootstrap が current viewer を継続して読めるようにする。
- password reset と fresh re-auth を Hosted UI ではなく app 既存導線上で扱えるようにする。
- Cognito が担当する認証責務と app 側が保持する internal user / mode / purchase / unlock / creator capability の境界を明確にする。
- account enumeration を避けつつ transport error vocabulary を固定する。

## Non-goals

- creator login
- Cognito Hosted UI を primary flow にすること
- social login / MFA / passkey
- frontend modal の visual 実装
- backend SDK 選定、Terraform、mail delivery 設定
- `currentViewer` や relation state を auth mutation response に重ねて返すこと

## Canonical Sources

- `docs/contracts/fan-auth-modal-ui-contract.md`
- `docs/contracts/viewer-profile-api-contract.md`
- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/TECH_STACK.md`
- `docs/ssot/product/account/identity-and-mode-model.md`

## Responsibility Boundary

- Cognito は `email ownership verification`、`password verification`、`sign up confirmation code delivery`、`password reset code delivery` を担当します。
- app backend は Cognito で認証済みの principal から internal user を解決または作成し、`shorts_fans_session`、`activeMode = fan`、creator capability、purchase / unlock / relation state を正として扱います。
- frontend custom UI は backend fan auth endpoint だけを叩き、Cognito token や Cognito user state を canonical viewer state として保持しません。
- auth mutation 成功後の viewer state は `GET /api/viewer/bootstrap` を再読して確定します。

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `POST` | `/api/fan/auth/sign-in` | public | `email + password` で sign in し、app session cookie を発行 |
| `POST` | `/api/fan/auth/sign-up` | public | sign up を開始し、確認コード入力待ち state へ進める |
| `POST` | `/api/fan/auth/sign-up/confirm` | public | 確認コードを消費し、app session cookie を発行 |
| `POST` | `/api/fan/auth/password-reset` | public | password reset code delivery を開始する |
| `POST` | `/api/fan/auth/password-reset/confirm` | public | reset code と新 password を確定する |
| `POST` | `/api/fan/auth/re-auth` | authenticated fan | current session の fresh auth を更新する |
| `DELETE` | `/api/fan/auth/session` | optional | app session revoke と cookie clear。cookie が無くても成功扱い |

## Shared Rules

- この issue では `provider = email` だけを扱います。
- `fan` と `creator` を別 identity に分けません。新規 session の `activeMode` は常に `fan` で開始します。
- custom modal UI が primary entry です。`/login` route を後続実装で残す場合も secondary fallback 扱いとし、別の auth contract を持ち込みません。
- login / logout / re-auth 成功 body に viewer state は返しません。成功後の current viewer は `GET /api/viewer/bootstrap` から読みます。
- sign-up flow では shared viewer profile の `displayName / handle / avatar` を初期化し、creator registration 時には同じ値を再入力しません。
- `displayName` と `handle` は sign-up request で受け、sign-up confirm 成功時に shared viewer profile の初期値として保存します。
- `handle` の collision 判定は `POST /api/fan/auth/sign-up` boundary で行い、confirm boundary は受理済み sign-up draft を消費するだけに留めます。
- sign-up 時の avatar は optional であり、選択された場合は sign-up confirm 後に `docs/contracts/viewer-profile-api-contract.md` の avatar upload / update transport を同じ sign-up flow の一部として直列実行し、既に作成済みの shared viewer profile に反映します。
- sign in は `missing email` と `wrong password` を wire 上で区別しません。
- sign up と password reset の開始 endpoint は、valid request から account existence を推測できないように扱います。
- sign up / password reset の開始 endpoint を再度呼ぶことは resend として扱えます。response shape は変えません。
- password reset confirm 成功は password 更新だけを意味し、自動 sign in や bootstrap success を意味しません。
- re-auth は current authenticated viewer の recent-auth proof を更新するための boundary であり、`activeMode` や creator capability を変更しません。

## Recent Auth Boundary

- high-risk mutation は current session が有効でも別途 `fresh auth` を要求できます。
- その場合、対象 endpoint 側は `403 + error.code = fresh_auth_required` を返し、frontend は `docs/contracts/fan-auth-modal-ui-contract.md` の `re-auth` mode を開きます。
- この issue では `fresh_auth_required` を使う downstream endpoint 自体は固定しません。対象選定は後続 contract に委ねます。

## Request / Response Contract

### `POST /api/fan/auth/sign-in`

#### Request

```json
{
  "email": "fan@example.com",
  "password": "VeryStrongPass123!"
}
```

#### Success

- `204`
- response body は返しません。
- `Set-Cookie` で `shorts_fans_session` を発行します。

### `POST /api/fan/auth/sign-up`

#### Request

```json
{
  "email": "fan@example.com",
  "displayName": "Mina Rei",
  "handle": "@minarei",
  "password": "VeryStrongPass123!"
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `email` | `string` | yes | sign-up 対象 email |
| `displayName` | `string` | yes | shared viewer profile の初期 display name |
| `handle` | `string` | yes | shared viewer profile の初期 handle。先頭 `@` 任意 |
| `password` | `string` | yes | Cognito sign-up password |

#### Success

- `200`
- `data.nextStep = "confirm_sign_up"` 固定です。
- `data.deliveryDestinationHint` は safe に返せる場合だけ masked email を返し、返せない場合は `null` です。
- accepted response は account existence、internal user 作成完了、shared viewer profile 初期化完了を保証しません。
- accepted sign-up draft は confirm でそのまま消費される前提で、`handle` の collision はこの boundary で解決します。
- frontend modal では `confirm_sign_up` を `confirm-sign-up` mode へ対応づけます。対応表は `docs/contracts/fan-auth-modal-ui-contract.md` を正とします。

```json
{
  "data": {
    "deliveryDestinationHint": "f***@example.com",
    "nextStep": "confirm_sign_up"
  },
  "meta": {
    "requestId": "req_fan_auth_sign_up_accepted_001",
    "page": null
  },
  "error": null
}
```

### `POST /api/fan/auth/sign-up/confirm`

#### Request

```json
{
  "confirmationCode": "123456",
  "email": "fan@example.com"
}
```

#### Success

- `204`
- response body は返しません。
- `Set-Cookie` で `shorts_fans_session` を発行します。
- backend は sign-up request で受けた `displayName / handle` を使って internal user と shared viewer profile を作成します。
- sign-up flow はこの後に optional avatar の初期化を完了させます。
  - avatar を選択していた場合は、authenticated になった直後に `docs/contracts/viewer-profile-api-contract.md` の avatar upload / update transport を呼び、modal を閉じる前に shared viewer profile へ反映します。

### `POST /api/fan/auth/password-reset`

#### Request

```json
{
  "email": "fan@example.com"
}
```

#### Success

- `200`
- `data.nextStep = "confirm_password_reset"` 固定です。
- `data.deliveryDestinationHint` は safe に返せる場合だけ masked email を返し、返せない場合は `null` です。
- accepted response は account existence を保証しません。
- frontend modal では `confirm_password_reset` を `confirm-password-reset` mode へ対応づけます。対応表は `docs/contracts/fan-auth-modal-ui-contract.md` を正とします。

```json
{
  "data": {
    "deliveryDestinationHint": "f***@example.com",
    "nextStep": "confirm_password_reset"
  },
  "meta": {
    "requestId": "req_fan_auth_password_reset_accepted_001",
    "page": null
  },
  "error": null
}
```

### `POST /api/fan/auth/password-reset/confirm`

#### Request

```json
{
  "confirmationCode": "123456",
  "email": "fan@example.com",
  "newPassword": "AnotherStrongPass456!"
}
```

#### Success

- `204`
- response body は返しません。
- app session cookie は発行しません。

### `POST /api/fan/auth/re-auth`

#### Request

```json
{
  "password": "VeryStrongPass123!"
}
```

#### Success

- `204`
- response body は返しません。
- backend は same session を継続しても rotate してもよく、client は `shorts_fans_session` を opaque に扱います。

### `DELETE /api/fan/auth/session`

#### Success

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
| `400` | `invalid_email` | email 形式が不正、または email payload が不正 |
| `400` | `invalid_display_name` | display name が空、または request boundary を満たさない |
| `400` | `invalid_handle` | handle が空、または許可外文字を含む |
| `400` | `invalid_password` | password が空、または request boundary を満たさない |
| `400` | `invalid_confirmation_code` | confirmation code が不正 |
| `400` | `confirmation_code_expired` | confirmation code が期限切れ |
| `400` | `password_policy_violation` | password が Cognito policy を満たさない |
| `401` | `invalid_credentials` | sign in または re-auth の email / password 組み合わせが不正 |
| `401` | `auth_required` | re-auth に current authenticated fan session が必要 |
| `403` | `confirmation_required` | sign in 前に email confirmation が必要 |
| `409` | `handle_already_taken` | normalized handle が既存 shared viewer profile と衝突 |
| `429` | `rate_limited` | retry guardrail または provider throttle に到達 |
| `500` | `internal_error` | 想定外の server failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_fan_auth_error_001",
    "page": null
  },
  "error": {
    "code": "invalid_credentials",
    "message": "email or password is invalid"
  }
}
```

- `confirmation_required` を受けた frontend modal は `confirm-sign-up` mode に遷移します。
- sign-in から `confirmation_required` に入った場合、resend に必要な password recovery は `docs/contracts/fan-auth-modal-ui-contract.md` の modal session rule に従います。

## Boundary Guardrails

- `currentViewer`、`activeMode`、`canAccessCreatorMode` は auth mutation response に重ねて返しません。
- Cognito token、provider subject、delivery medium、password policy raw text、internal user mapping id は返しません。
- `fan / creator` を別 login identity に分けません。
- sign up accepted や password reset accepted を、account existence の公開シグナルとして使いません。
- challenge token や provider 固有 state を app contract に持ち込みません。
- `/login` redirect-only を正規 UX として固定しません。primary entry は shared modal です。

## Fixture Reference

- representative fixture は [fan-auth.json](fixtures/fan-auth.json) を参照します。
