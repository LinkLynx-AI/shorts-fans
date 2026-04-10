# Viewer Creator Entry API Contract

## 位置づけ

- この文書は `SHO-121 fan profile から creator entry を始める` の transport 契約を固定します。
- fan private hub から creator registration を開始し、success surface を挟んで `activeMode` を `creator` へ切り替える最小導線を対象にします。
- actual `/creator` dashboard surface は別 PR の責務とし、この leaf では route entry と self mutation だけを扱います。

## Goals

- authenticated fan が `fan profile` から creator registration を開始できるようにする。
- creator registration 前に任意 avatar を upload し、completed upload を registration から参照できるようにする。
- registration 完了後に `canAccessCreatorMode = true` を bootstrap へ反映できるようにする。
- success surface の CTA から `activeMode = creator` を明示的に切り替えられるようにする。
- creator registration の時点で unique な handle を取得し、既存 creator を含めて handle を必須に保つ。

## Non-goals

- creator dashboard / creator home 自体の read 契約
- review / pending / reject / resubmit workflow
- avatar crop / delete
- public creator search / profile transport shape の新設・変更

## Canonical Sources

- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/creator-workspace-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/contracts/mvp-media-workflow-contract.md`
- `docs/ssot/product/account/identity-and-mode-model.md`
- `docs/ssot/product/account/account-permissions.md`
- `docs/ssot/product/creator/creator-onboarding-surface.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `POST` | `/api/viewer/creator-registration/avatar-uploads` | required | registration 前の avatar direct upload target を返す |
| `POST` | `/api/viewer/creator-registration/avatar-uploads/complete` | required | upload 済み avatar を registration から参照できる completed token にする |
| `POST` | `/api/viewer/creator-registration` | required | fan profile から始める self-serve creator registration |
| `PUT` | `/api/viewer/active-mode` | required | `fan` / `creator` の active mode switch |

## Avatar Upload Shared Rules

- avatar は creator registration の optional field であり、未選択でも registration は成功します。
- avatar upload flow は authenticated viewer 向けで、creator capability の有無は前提にしません。
- `avatarUploadToken` は opaque token として扱い、client は中身を解釈しません。
- `avatarUploadToken` は upload create を行った viewer 自身の completed upload にだけ紐づきます。
- upload create / complete は `data / meta / error` envelope を使い、`meta.page = null` とします。
- `register` は引き続き `204 No Content` を返し、avatar upload metadata は返しません。
- avatar upload は `Presigned PUT` を前提にし、multipart / binary / base64 を `register` request に同梱しません。
- avatar upload で許可する file 制約は次です。
  - `mimeType`: `image/jpeg`、`image/png`、`image/webp`
  - `fileSizeBytes`: `> 0` かつ `<= 5_242_880`
- 同じ completed token は successful registration で消費されるまで再利用できます。`handle_already_taken` のような別理由で registration が失敗した場合は、token が未期限切れなら再送して構いません。

## `POST /api/viewer/creator-registration/avatar-uploads`

### Request

```json
{
  "fileName": "mina-avatar.webp",
  "mimeType": "image/webp",
  "fileSizeBytes": 418204
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `fileName` | `string` | yes | trim 後 non-empty |
| `mimeType` | `string` | yes | `image/jpeg` / `image/png` / `image/webp` のみ |
| `fileSizeBytes` | `number` | yes | `> 0` かつ `<= 5_242_880` |

### Success

- status は `200 OK`
- body は direct upload target と completed 後に registration へ渡す `avatarUploadToken` を返します
- `avatarUploadToken` は create 時点では未完了であり、`complete` 成功前に registration へ渡しても利用できません

```json
{
  "data": {
    "avatarUploadToken": "vcupl_01hrx8b0k6w1h6h2a8m4e9d2pr",
    "expiresAt": "2026-04-09T12:15:00Z",
    "uploadTarget": {
      "fileName": "mina-avatar.webp",
      "mimeType": "image/webp",
      "upload": {
        "headers": {
          "Content-Type": "image/webp"
        },
        "method": "PUT",
        "url": "https://raw-bucket.example.com/presigned/avatar"
      }
    }
  },
  "meta": {
    "requestId": "req_viewer_creator_avatar_upload_create_001",
    "page": null
  },
  "error": null
}
```

### Success Rules

- client は `uploadTarget.upload.url` と `uploadTarget.upload.headers` をそのまま使って direct `PUT` を行います。
- create 成功だけでは registration で使える completed avatar を意味しません。
- bucket / key / storage provider は opaque とし、client が推測しません。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_avatar_mime_type` | 許可外 MIME |
| `400` | `invalid_avatar_file_size` | `fileSizeBytes <= 0` |
| `400` | `avatar_file_too_large` | `5_242_880` bytes 超過 |
| `401` | `auth_required` | session 不在 |
| `500` | `internal_error` | unexpected failure |

## `POST /api/viewer/creator-registration/avatar-uploads/complete`

### Request

```json
{
  "avatarUploadToken": "vcupl_01hrx8b0k6w1h6h2a8m4e9d2pr"
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `avatarUploadToken` | `string` | yes | create response で返した opaque token |

### Success

- status は `200 OK`
- upload object の存在確認と image validation に成功した場合だけ completed avatar を返します
- response の `avatar` は preview / workspace 表示用に `MediaAsset` へ正規化した shape です

```json
{
  "data": {
    "avatarUploadToken": "vcupl_01hrx8b0k6w1h6h2a8m4e9d2pr",
    "avatar": {
      "id": "asset_creator_registration_avatar_01hrx8b0k6w1h6h2a8m4e9d2pr",
      "kind": "image",
      "url": "https://cdn.example.com/creator/registration/avatar-01hrx8b0.webp",
      "posterUrl": null,
      "durationSeconds": null
    }
  },
  "meta": {
    "requestId": "req_viewer_creator_avatar_upload_complete_001",
    "page": null
  },
  "error": null
}
```

### Success Rules

- completed token は create した viewer 自身だけが registration で利用できます。
- `avatar.kind` は常に `image`、`posterUrl` と `durationSeconds` は常に `null` です。
- completed token は successful registration で消費されるまでは、同じ viewer の registration retry に使えます。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `401` | `auth_required` | session 不在 |
| `404` | `avatar_upload_not_found` | token を解決できない、または他 viewer の token |
| `409` | `avatar_upload_incomplete` | direct upload 未完了 |
| `409` | `avatar_upload_expired` | token 期限切れ |
| `500` | `internal_error` | unexpected failure |

## `POST /api/viewer/creator-registration`

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
| `displayName` | `string` | yes | trim 後 non-empty |
| `handle` | `string` | yes | trim 後 non-empty、先頭 `@` 任意、保存時は lowercase + `@` なし |
| `avatarUploadToken` | `string` | no | completed avatar upload を指す opaque token。未選択なら送らない |
| `bio` | `string` | no | omitted または empty string は empty string と同義 |

### Success

- status は `204 No Content`
- 登録成功時点では `activeMode` を切り替えません
- 後続の `GET /api/viewer/bootstrap` では `canAccessCreatorMode = true`、`activeMode = fan` を返します

### Registration Rules

- request user を viewer 自身の creator として登録します。
- creator capability は即時 `approved` で upsert します。
- creator profile は registration transaction 内で upsert し、暫定運用として成功時に即時 public 化します。
- この即時 public 化は temporary behavior であり、creator review workflow 実装時に削除します。
- `avatarUploadToken` を送らない場合:
  - 初回登録は `avatar_url = null` で作成し、その後 registration 成功時に `published_at` を設定します。
  - 既存 profile がある場合は `displayName`、`handle`、`bio` を更新し、既存 `avatar_url` を保持します。
- `avatarUploadToken` を送る場合:
  - token は same viewer が create / complete 済みであり、未期限切れかつ未消費である必要があります。
  - 初回登録・既存 profile 更新のどちらでも、resolved avatar を `avatar_url` へ反映します。
  - successful registration で token を消費します。
- 既存 profile が既に public の場合は、既存 `published_at` を保持します。
- `bio` を省略した場合は empty string として扱います。
- avatar 未選択は正常系であり、avatar upload failure と混同しません。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_display_name` | trim 後 empty |
| `400` | `invalid_handle` | trim 後 empty、または許可外文字を含む |
| `400` | `invalid_avatar_upload_token` | token が未完了、期限切れ、他 viewer 所有、または解決不可 |
| `401` | `auth_required` | session 不在 |
| `409` | `handle_already_taken` | normalized handle が既存 creator と衝突 |
| `500` | `internal_error` | unexpected failure |

## `PUT /api/viewer/active-mode`

### Request

```json
{
  "activeMode": "creator"
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `activeMode` | `"fan" \| "creator"` | yes | target mode |

### Success

- status は `204 No Content`
- current viewer の session active mode だけを切り替えます

### Mode Switch Rules

- `fan -> creator` は approved creator capability がある viewer だけ許可します。
- `creator -> fan` は authenticated viewer なら許可します。
- mode switch 完了後の canonical current viewer state は `GET /api/viewer/bootstrap` で確認します。
- frontend は registration 完了後に success surface を挟み、そこから明示 CTA で `creator` へ切り替える前提です。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_active_mode` | `"fan" \| "creator"` 以外 |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なしで `creator` を要求 |
| `500` | `internal_error` | unexpected failure |

## Boundary Guardrails

- registration endpoint は profile preview や dashboard data を返しません。
- registration endpoint は avatar binary / multipart / base64 payload を受けず、completed token 参照だけを扱います。
- creator registration と mode switch は transport 上で分離し、登録直後に自動遷移しません。
- review state や checklist はこの leaf では transport に含めません。

## Fixture Reference

- representative fixture は [viewer-creator-entry.json](fixtures/viewer-creator-entry.json) を参照します。
