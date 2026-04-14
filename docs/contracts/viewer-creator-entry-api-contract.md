# Viewer Creator Entry API Contract

## 位置づけ

- この文書は `SHO-121 fan profile から creator entry を始める` の transport 契約を固定します。
- fan private hub から creator registration を開始し、success surface を挟んで `activeMode` を `creator` へ切り替える最小導線を対象にします。
- creator registration は shared viewer profile を使って creator capability を付与する entry であり、`displayName / handle / avatar` の入力面ではありません。

## Goals

- authenticated fan が `fan profile` から creator registration を開始できるようにする。
- creator registration 完了後に `canAccessCreatorMode = true` を bootstrap へ反映できるようにする。
- success surface の CTA から `activeMode = creator` を明示的に切り替えられるようにする。
- creator registration 時に shared viewer profile を読み、fan / creator で同じ `displayName / handle / avatar` を使う。

## Non-goals

- creator dashboard / creator home 自体の read 契約
- shared viewer profile の作成 / 編集
- creator registration 時の `displayName / handle / avatar / bio` 入力 UI
- review / pending / reject / resubmit workflow

## Canonical Sources

- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/viewer-profile-api-contract.md`
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
| `POST` | `/api/viewer/creator-registration` | required | fan profile から始める self-serve creator registration |
| `PUT` | `/api/viewer/active-mode` | required | `fan` / `creator` の active mode switch |

## Shared Rules

- creator registration は current viewer 自身だけを対象にします。
- `displayName / handle / avatar` は `docs/contracts/viewer-profile-api-contract.md` の shared viewer profile から読みます。
- current primary flow では sign-up flow または fan settings で shared viewer profile を準備してから creator registration に進みます。
- creator registration 完了時点では `activeMode` を自動切替しません。
- legacy な creator-registration avatar upload route が残っていても、この文書の canonical flow には含めません。

## `POST /api/viewer/creator-registration`

### Request

```json
{}
```

### Success

- status は `204 No Content`
- 登録成功時点では `activeMode` を切り替えません。
- 後続の `GET /api/viewer/bootstrap` では `canAccessCreatorMode = true`、`activeMode = fan` を返します。

### Registration Rules

- request user を viewer 自身の creator として登録します。
- creator capability は即時 `approved` で upsert します。
- creator profile は registration transaction 内で upsert し、shared viewer profile の `displayName / handle / avatar` を mirror します。
- creator 固有の `bio` は registration 時点では空文字で初期化し、後続の `/api/creator/workspace/profile` で更新します。
- creator profile は暫定運用として successful registration 時に即時 public 化します。
- shared viewer profile が sign-up flow で既に作成済み、または fan settings で準備済みである前提で動作し、registration request 自体は profile field を受け取りません。
- `handle` の uniqueness は shared viewer profile 準備時点で解決済みとし、creator registration 側で再裁定しません。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_display_name` | shared viewer profile の display name が不正 |
| `400` | `invalid_handle` | shared viewer profile の handle が不正 |
| `401` | `auth_required` | session 不在 |
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
- current viewer の session active mode だけを切り替えます。

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
- registration endpoint は `displayName / handle / avatar / bio` を request body で受けません。
- creator registration と mode switch は transport 上で分離し、登録直後に自動遷移しません。
- review state や checklist はこの leaf では transport に含めません。

## Fixture Reference

- representative fixture は [viewer-creator-entry.json](fixtures/viewer-creator-entry.json) を参照します。
