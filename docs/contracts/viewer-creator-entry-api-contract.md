# Viewer Creator Entry API Contract

## 位置づけ

- この文書は `SHO-121 fan profile から creator entry を始める` の transport 契約を固定します。
- fan private hub から creator registration を開始し、success surface を挟んで `activeMode` を `creator` へ切り替える最小導線を対象にします。
- actual `/creator` dashboard surface は別 PR の責務とし、この leaf では route entry と self mutation だけを扱います。

## Goals

- authenticated fan が `fan profile` から creator registration を開始できるようにする。
- registration 完了後に `canAccessCreatorMode = true` を bootstrap へ反映できるようにする。
- success surface の CTA から `activeMode = creator` を明示的に切り替えられるようにする。
- creator registration の時点で unique な handle を取得し、既存 creator を含めて handle を必須に保つ。

## Non-goals

- creator dashboard / creator home 自体の read 契約
- review / pending / reject / resubmit workflow
- avatar upload
- public creator publish / public search 露出

## Canonical Sources

- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/account/identity-and-mode-model.md`
- `docs/ssot/product/account/account-permissions.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `POST` | `/api/viewer/creator-registration` | required | fan profile から始める self-serve creator registration |
| `PUT` | `/api/viewer/active-mode` | required | `fan` / `creator` の active mode switch |

## `POST /api/viewer/creator-registration`

### Request

```json
{
  "displayName": "Mina Rei",
  "handle": "@minarei",
  "bio": "quiet rooftop の continuation を中心に投稿します。"
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `displayName` | `string` | yes | trim 後 non-empty |
| `handle` | `string` | yes | trim 後 non-empty、先頭 `@` 任意、保存時は lowercase + `@` なし |
| `bio` | `string` | yes | empty string は許容 |

### Success

- status は `204 No Content`
- 登録成功時点では `activeMode` を切り替えません
- 後続の `GET /api/viewer/bootstrap` では `canAccessCreatorMode = true`、`activeMode = fan` を返します

### Registration Rules

- request user を viewer 自身の creator として登録します。
- creator capability は即時 `approved` で upsert します。
- creator profile は private profile として upsert し、初回は normalized handle、`avatar_url = null`、`published_at = null` で保存します。
- 既存 profile がある場合は `displayName`、`handle`、`bio` を更新し、`avatar_url` / `published_at` は保持します。
- registration 完了だけでは public creator profile にはなりません。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_display_name` | trim 後 empty |
| `400` | `invalid_handle` | trim 後 empty、または許可外文字を含む |
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
- creator registration と mode switch は transport 上で分離し、登録直後に自動遷移しません。
- review state や checklist はこの leaf では transport に含めません。

## Fixture Reference

- representative fixture は [viewer-creator-entry.json](fixtures/viewer-creator-entry.json) を参照します。
