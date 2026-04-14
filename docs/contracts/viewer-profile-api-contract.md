# Viewer Profile API Contract

## 位置づけ

- この文書は `fan` と `creator` が共有する `displayName / handle / avatar` の self profile transport 契約を固定します。
- `fan profile settings` と `creator workspace profile settings` は同じ shared viewer profile を編集し、creator 固有の `bio` だけを別責務として扱います。
- creator registration 自体は profile 入力の場ではなく、既に作成済みの shared viewer profile を読むだけの entry とします。

## Goals

- sign-up 時点で shared viewer profile の初期値を作成できるようにする。
- authenticated viewer が `fan profile settings` から shared profile を更新できるようにする。
- creator mode viewer が `/creator` から shared profile と creator 固有 bio を更新できるようにする。
- shared profile 更新時に creator mirror を同期し、public / workspace 表示の `displayName / handle / avatar` を一致させる。

## Non-goals

- public creator profile の read 契約
- avatar crop / delete / history
- creator registration review workflow
- creator analytics や workspace overview の read 契約

## Canonical Sources

- `docs/contracts/fan-auth-api-contract.md`
- `docs/contracts/fan-profile-api-contract.md`
- `docs/contracts/viewer-creator-entry-api-contract.md`
- `docs/contracts/creator-workspace-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/account/identity-and-mode-model.md`

## Shared Rules

- `displayName / handle / avatar` の canonical source は shared viewer profile です。
- creator profile が存在する場合、shared viewer profile 更新時に `displayName / handle / avatar` を mirror 同期します。
- `bio` は creator 固有情報であり shared viewer profile には含めません。
- `displayName` は trim 後 non-empty である必要があります。
- `handle` は trim 後 non-empty である必要があります。
- request の `handle` は先頭 `@` を付けても構いません。
- 保存時の `handle` は leading `@` を外し lowercase 化します。
- `handle` で許可する文字は `a-z`、`0-9`、`.`、`_` だけです。
- response の `handle` は viewer / creator 表示用に先頭 `@` 付きで返します。
- avatar は optional であり、未設定時は `null` を返します。
- `avatarUploadToken` を送らない更新は現在の avatar を保持します。
- avatar upload token は opaque token として扱い、client は内容を解釈しません。

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/viewer/profile` | required | current viewer の shared profile を返す |
| `POST` | `/api/viewer/profile/avatar-uploads` | required | shared profile 用 avatar direct upload target を返す |
| `POST` | `/api/viewer/profile/avatar-uploads/complete` | required | upload 済み avatar を completed token に変換する |
| `PUT` | `/api/viewer/profile` | required | fan settings などから shared profile を更新する |
| `PUT` | `/api/creator/workspace/profile` | required creator | shared profile と creator 固有 bio を同時更新する |

## Avatar Upload Shared Rules

- avatar upload は authenticated viewer 向けで、fan / creator のどちらでも利用できます。
- upload は `Presigned PUT` を前提にし、multipart / binary / base64 を profile update request に同梱しません。
- upload create / complete は `data / meta / error` envelope を使い、`meta.page = null` とします。
- `avatarUploadToken` は create した viewer 自身の completed upload にだけ紐づきます。
- completed token は successful update で消費されるまで再利用できます。
- file 制約は次です。
  - `mimeType`: `image/jpeg`、`image/png`、`image/webp`
  - `fileSizeBytes`: `> 0` かつ `<= 5_242_880`

## `GET /api/viewer/profile`

### Success

- status は `200 OK`
- `data.profile` は shared viewer profile を返します。

```json
{
  "data": {
    "profile": {
      "avatar": {
        "id": "asset_viewer_profile_avatar_760f5b5cb41c4cc48817d7f90d7ef6d6",
        "kind": "image",
        "url": "https://cdn.example.com/avatars/mina.jpg",
        "posterUrl": null,
        "durationSeconds": null
      },
      "displayName": "Mina Rei",
      "handle": "@minarei"
    }
  },
  "meta": {
    "requestId": "req_viewer_profile_get_001",
    "page": null
  },
  "error": null
}
```

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `401` | `auth_required` | session 不在 |
| `404` | `not_found` | shared viewer profile を解決できない不整合 |
| `500` | `internal_error` | unexpected failure |

## `POST /api/viewer/profile/avatar-uploads`

### Request

```json
{
  "fileName": "mina-avatar.webp",
  "mimeType": "image/webp",
  "fileSizeBytes": 418204
}
```

### Success

- status は `200 OK`
- body は direct upload target と completed 後に update request へ渡す `avatarUploadToken` を返します。

## `POST /api/viewer/profile/avatar-uploads/complete`

### Request

```json
{
  "avatarUploadToken": "vcupl_01hrx8b0k6w1h6h2a8m4e9d2pr"
}
```

### Success

- status は `200 OK`
- upload object の存在確認と image validation に成功した場合だけ completed avatar を返します。

### Upload Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_avatar_mime_type` | 許可外 MIME |
| `400` | `invalid_avatar_file_size` | `fileSizeBytes <= 0` |
| `400` | `avatar_file_too_large` | `5_242_880` bytes 超過 |
| `401` | `auth_required` | session 不在 |
| `404` | `avatar_upload_not_found` | token を解決できない、または他 viewer の token |
| `409` | `avatar_upload_incomplete` | direct upload 未完了 |
| `409` | `avatar_upload_expired` | token 期限切れ |
| `500` | `internal_error` | unexpected failure |

## `PUT /api/viewer/profile`

### Request

```json
{
  "displayName": "Mina Rei",
  "handle": "@minarei",
  "avatarUploadToken": "vcupl_01hrx8b0k6w1h6h2a8m4e9d2pr"
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `displayName` | `string` | yes | trim 後 non-empty |
| `handle` | `string` | yes | trim 後 non-empty。先頭 `@` 任意 |
| `avatarUploadToken` | `string` | no | completed avatar upload token。省略時は現在 avatar を保持 |

### Success

- status は `204 No Content`
- response body は返しません。

### Update Rules

- shared viewer profile の `displayName / handle / avatar` を更新します。
- creator profile が存在する場合、同じ transaction で `displayName / handle / avatar` を mirror 同期します。
- `avatarUploadToken` を送る場合は same viewer の completed upload である必要があります。
- successful update 後に `avatarUploadToken` を消費します。
- canonical current viewer state は `GET /api/viewer/profile` または `GET /api/viewer/bootstrap` の再読で確認します。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_display_name` | trim 後 empty |
| `400` | `invalid_handle` | trim 後 empty、または許可外文字を含む |
| `400` | `invalid_avatar_upload_token` | token が未完了、期限切れ、他 viewer 所有、または解決不可 |
| `401` | `auth_required` | session 不在 |
| `404` | `not_found` | shared viewer profile を解決できない不整合 |
| `409` | `handle_already_taken` | normalized handle が既存 profile と衝突 |
| `500` | `internal_error` | unexpected failure |

## `PUT /api/creator/workspace/profile`

### Request

```json
{
  "displayName": "Mina Rei",
  "handle": "@minarei",
  "avatarUploadToken": "vcupl_01hrx8b0k6w1h6h2a8m4e9d2pr",
  "bio": "quiet rooftop の continuation を中心に投稿します。"
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `displayName` | `string` | yes | shared viewer profile へ反映 |
| `handle` | `string` | yes | shared viewer profile へ反映 |
| `avatarUploadToken` | `string` | no | completed avatar upload token。省略時は現在 avatar を保持 |
| `bio` | `string` | yes | creator 固有の自己紹介文 |

### Success

- status は `204 No Content`
- response body は返しません。

### Update Rules

- shared viewer profile の `displayName / handle / avatar` と creator 固有 `bio` を同時更新します。
- caller は approved creator capability を持つ必要があります。
- shared profile update と creator bio update の canonical result は `/creator` workspace と public creator surface の再読で確認します。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_display_name` | trim 後 empty |
| `400` | `invalid_handle` | trim 後 empty、または許可外文字を含む |
| `400` | `invalid_avatar_upload_token` | token が未完了、期限切れ、他 viewer 所有、または解決不可 |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | shared viewer profile を解決できない不整合 |
| `409` | `handle_already_taken` | normalized handle が既存 profile と衝突 |
| `500` | `internal_error` | unexpected failure |

## Boundary Guardrails

- shared viewer profile transport は creator registration input を兼ねません。
- `bio` は `/api/viewer/profile` では扱いません。
- avatar binary / multipart / base64 payload を profile update request に含めません。
- profile update response に `currentViewer`、`activeMode`、workspace summary を重ねて返しません。
