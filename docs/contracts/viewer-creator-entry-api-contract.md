# Viewer Creator Entry API Contract

## 位置づけ

- この文書は `SHO-121 fan profile から creator entry を始める` を土台に、`SHO-171 creator registration / review の契約と状態境界を更新する` の transport 契約を固定します。
- fan private hub から始める creator onboarding のうち、`status read`、`initial submit`、`active mode switch` の境界だけを扱います。
- creator registration は shared viewer profile を preview しながら creator capability 審査へ進む entry であり、`displayName / handle / avatar` の再入力面ではありません。

## Goals

- authenticated fan が current viewer 自身の creator registration status を読めるようにする。
- creator registration を `draft / submitted / approved / rejected / suspended` 前提へ置き換え、`approved` まで creator mode を開かないようにする。
- approval 前 surface を `read-only onboarding surface + shared viewer profile basics preview + private bio draft` に限定する。
- rejection metadata を transport と fixture の両方で辿れるようにする。

## Non-goals

- creator onboarding intake payload、required docs、証跡 upload の concrete schema
- shared viewer profile の作成 / 編集
- admin review UI と review decision mutation
- creator dashboard / creator home 自体の read 契約

## Canonical Sources

- `docs/contracts/fan-auth-api-contract.md`
- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/viewer-profile-api-contract.md`
- `docs/contracts/viewer-creator-registration-intake-api-contract.md`
- `docs/contracts/creator-workspace-api-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/account/identity-and-mode-model.md`
- `docs/ssot/product/account/account-permissions.md`
- `docs/ssot/product/creator/creator-onboarding-surface.md`
- `docs/ssot/product/creator/creator-workflow.md`
- `docs/ssot/product/moderation/moderation-and-review.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/viewer/creator-registration` | required | current viewer の creator registration status read |
| `POST` | `/api/viewer/creator-registration` | required | `draft` submit と eligible `rejected` からの self-serve resubmit |
| `PUT` | `/api/viewer/active-mode` | required | `fan` / `creator` の active mode switch |

## Shared Rules

- creator registration は current viewer 自身だけを対象にします。
- auth は `docs/contracts/fan-auth-api-contract.md` の app session cookie を前提にし、Cognito token や provider state を request / response に含めません。
- `displayName / handle / avatar` は `docs/contracts/viewer-profile-api-contract.md` の shared viewer profile から読みます。
- approval 前 state では public creator profile を公開せず、creator workspace / upload / submission package 作成 UI を解放しません。
- creator 固有の `bio` は onboarding draft として private に保持してよいですが、write transport 自体はこの leaf では固定しません。
- registration submit 完了時点では `activeMode` を自動切替しません。
- registration submit は creator capability を即時 `approved` にせず、後続の `GET /api/viewer/bootstrap` でも `canAccessCreatorMode = false` を維持します。
- `reasonCode` は opaque string として扱い、この leaf では enum を固定しません。
- legacy な creator-registration avatar upload route が残っていても、この文書の canonical flow には含めません。

## `GET /api/viewer/creator-registration`

### Success

- status は `200 OK`
- `data.registration` は current viewer の onboarding case が未開始なら `null`、開始済みなら `CreatorRegistrationStatus` を返します。
- `meta.page = null`、`error = null` を維持します。

```json
{
  "data": {
    "registration": {
      "state": "submitted",
      "sharedProfile": {
        "avatar": {
          "id": "asset_viewer_profile_avatar_760f5b5cb41c4cc48817d7f90d7ef6d6",
          "kind": "image",
          "url": "https://cdn.example.com/avatars/mina.jpg",
          "posterUrl": null,
          "durationSeconds": null
        },
        "displayName": "Mina Rei",
        "handle": "@minarei"
      },
      "creatorDraft": {
        "bio": "quiet rooftop の continuation を中心に投稿します。"
      },
      "review": {
        "submittedAt": "2026-04-16T09:30:00Z",
        "approvedAt": null,
        "rejectedAt": null,
        "suspendedAt": null
      },
      "rejection": null,
      "surface": {
        "kind": "read_only_onboarding",
        "workspacePreview": "static_mock"
      },
      "actions": {
        "canSubmit": false,
        "canResubmit": false,
        "canEnterCreatorMode": false
      }
    }
  },
  "meta": {
    "requestId": "req_viewer_creator_registration_status_001",
    "page": null
  },
  "error": null
}
```

### `CreatorRegistrationStatus`

| field | type | notes |
| --- | --- | --- |
| `state` | `"draft" \| "submitted" \| "approved" \| "rejected" \| "suspended"` | creator capability review state |
| `sharedProfile` | `ViewerProfilePreview` | current shared viewer profile の read-only preview |
| `creatorDraft.bio` | `string` | private creator bio draft。未入力なら空文字 |
| `review` | `CreatorRegistrationReviewTimeline` | review timestamps の summary |
| `rejection` | `CreatorRegistrationRejection \| null` | `state = rejected` のときだけ object |
| `surface` | `CreatorRegistrationSurface` | approval 前 guardrail の解釈を固定する |
| `actions` | `CreatorRegistrationActions` | current state から取りうる次 action |

### `ViewerProfilePreview`

- shape は `docs/contracts/viewer-profile-api-contract.md` の `GET /api/viewer/profile` における `profile` と同じです。
- `displayName / handle / avatar` は preview 専用であり、この endpoint から更新しません。
- shared viewer profile を直したい場合は `docs/contracts/viewer-profile-api-contract.md` を使います。

### `CreatorRegistrationReviewTimeline`

| field | type | notes |
| --- | --- | --- |
| `submittedAt` | `string \| null` | ISO-8601。未 submit は `null` |
| `approvedAt` | `string \| null` | ISO-8601 |
| `rejectedAt` | `string \| null` | ISO-8601 |
| `suspendedAt` | `string \| null` | ISO-8601 |

### `CreatorRegistrationRejection`

| field | type | notes |
| --- | --- | --- |
| `reasonCode` | `string` | opaque string。文言化は frontend / copy 側の責務 |
| `isResubmitEligible` | `boolean` | fixable reject に限り `true` |
| `isSupportReviewRequired` | `boolean` | support / manual review を通すまで self-serve resubmit 不可 |
| `selfServeResubmitCount` | `number` | accepted 済み self-serve resubmit 回数 |
| `selfServeResubmitRemaining` | `number` | `0..2`。同じ onboarding case で残る self-serve 回数 |

### `CreatorRegistrationSurface`

| field | type | notes |
| --- | --- | --- |
| `kind` | `"read_only_onboarding" \| "creator_workspace"` | 現在 user に見せてよい surface |
| `workspacePreview` | `"static_mock" \| null` | approval 前は `static_mock`、approved 後は `null` |

### `CreatorRegistrationActions`

| field | type | notes |
| --- | --- | --- |
| `canSubmit` | `boolean` | `draft` のときだけ `true` |
| `canResubmit` | `boolean` | `rejected` かつ self-serve resubmit が許可され、残回数があるときだけ `true` |
| `canEnterCreatorMode` | `boolean` | approved creator capability を持つときだけ `true` |

### Status Rules

- onboarding case 未開始の viewer は `data.registration = null` を返します。
- `draft / submitted / rejected / suspended` はすべて `surface.kind = read_only_onboarding`、`surface.workspacePreview = "static_mock"` を返します。
- `approved` は `surface.kind = creator_workspace`、`surface.workspacePreview = null`、`actions.canEnterCreatorMode = true` を返します。
- `rejection` object は `state = rejected` のときだけ返し、それ以外は `null` です。
- `selfServeResubmitRemaining` は `2 - selfServeResubmitCount` を返し、負値にはしません。
- `actions.canResubmit = true` になるのは、`state = rejected` かつ `isResubmitEligible = true`、`isSupportReviewRequired = false`、`selfServeResubmitRemaining > 0` を同時に満たすときだけです。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `401` | `auth_required` | session 不在 |
| `404` | `not_found` | shared viewer profile 不在 |
| `500` | `internal_error` | unexpected failure |

## `POST /api/viewer/creator-registration`

### Request

```json
{}
```

### Success

- status は `204 No Content`
- canonical current state は `GET /api/viewer/creator-registration` と `GET /api/viewer/bootstrap` の再読で確認します。

### Submit Rules

- target は current viewer 自身の onboarding case だけです。
- onboarding case が `draft` の場合は initial submit として `submitted` へ遷移します。
- onboarding case が eligible な `rejected` の場合は self-serve resubmit として `submitted` へ再遷移します。
- onboarding case 未開始の viewer は、先に `docs/contracts/viewer-creator-registration-intake-api-contract.md` で draft を保存してから submit します。
- submit 成功は creator capability 付与や public creator profile 公開を意味しません。
- submit 後も `GET /api/viewer/bootstrap` では `canAccessCreatorMode = false`、`activeMode = fan` を返します。
- concrete な intake payload、required docs、evidence validation は `docs/contracts/viewer-creator-registration-intake-api-contract.md` を正とします。
- self-serve resubmit 成功時は `selfServeResubmitCount` を `+1` し、rejection metadata を clear します。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_display_name` | shared viewer profile の display name が不正 |
| `400` | `invalid_handle` | shared viewer profile の handle が不正 |
| `401` | `auth_required` | session 不在 |
| `409` | `registration_incomplete` | required intake / evidence が不足している |
| `409` | `registration_state_conflict` | `draft` でも eligible `rejected` でもない state から submit / resubmit を要求 |
| `500` | `internal_error` | unexpected failure |

## Rejected / Resubmit

- `rejected` status と rejection metadata 自体は `GET /api/viewer/creator-registration` で読めます。
- eligible な fixable reject は同じ `POST /api/viewer/creator-registration` を使って self-serve resubmit します。
- support review required の reject と `suspended` は self-serve resubmit 不可です。

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
- `draft / submitted / rejected / suspended` の creator registration status は creator mode 解放条件になりません。
- `creator -> fan` は authenticated viewer なら許可します。
- mode switch 完了後の canonical current viewer state は `GET /api/viewer/bootstrap` で確認します。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_active_mode` | `"fan" \| "creator"` 以外 |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なしで `creator` を要求 |
| `500` | `internal_error` | unexpected failure |

## Boundary Guardrails

- creator registration status は bootstrap に重ねて返しません。app shell の global self state は引き続き `GET /api/viewer/bootstrap`、status detail はこの contract を正とします。
- registration submit endpoint は `displayName / handle / avatar / bio` を request body で受けません。
- approval 前 surface は `read-only onboarding surface + shared viewer profile preview + private bio draft + static mock workspace preview` に限定します。
- registration 成功時に dashboard data、upload target、public creator profile header を返しません。
- public creator profile の公開可否は creator capability approval と別に管理し、registration submit transaction 内で即時 public 化しません。

## Fixture Reference

- representative fixture は [viewer-creator-entry.json](fixtures/viewer-creator-entry.json) を参照します。
