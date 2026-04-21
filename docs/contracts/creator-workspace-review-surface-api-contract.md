# Creator Workspace Review Surface API Contract

## 位置づけ

- この文書は `/creator` workspace 内で owner 自身が current submission package の review status を読む private read contract を固定します。
- dashboard の要対応通知、main / short tile badge、detail 内の審査状態表示に必要な最小 read surface だけを扱います。
- actual submit / resubmit mutation は [creator-workspace-submission-review-api-contract.md](creator-workspace-submission-review-api-contract.md) を正とし、この文書では creator-facing read state と error vocabulary を固定します。

## Goals

- `/creator` dashboard で current package の要対応状態を notification として表示するための package summary を返す。
- main / short tile に表示する要対応 review badge を dedicated read surface で返す。
- main detail / short detail で target state、package state、blockers を表示できるように返す。
- short detail からでも canonical `main` anchor の package 状態を説明できるよう、target と package を分離して返す。

## Non-goals

- `/creator/review` の新規 route
- review queue や reviewer コメント一覧のような admin surface
- publish / unlock の public eligibility 判定を creator-facing read surface で代替すること
- mutation response に review state を混ぜること
- `rejected` object を self-serve reopen する新しい mutation

## Canonical Sources

- [submission-package-review-contract.md](submission-package-review-contract.md)
- [creator-workspace-submission-review-api-contract.md](creator-workspace-submission-review-api-contract.md)
- [creator-workspace-owner-preview-api-contract.md](creator-workspace-owner-preview-api-contract.md)
- [mvp-core-domain-contract.md](mvp-core-domain-contract.md)
- [mvp-media-workflow-contract.md](mvp-media-workflow-contract.md)

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/creator/workspace/review-surface` | required | dashboard 通知と tile badge 用の package summary / item state |
| `GET` | `/api/creator/workspace/mains/{mainId}/review-surface` | required | main detail 用の review surface |
| `GET` | `/api/creator/workspace/shorts/{shortId}/review-surface` | required | short detail 用の review surface |

## Shared Vocabulary

### `WorkspaceReviewPackageReadiness`

| value | notes |
| --- | --- |
| `ready` | 現在の package は review intake / resubmit の readiness を満たしている |
| `blocked` | readiness blocker が存在し action 不可 |
| `conflict` | blocker はないが current review state から action 不可 |
| `none` | pending intake などにより action を出さない |

### `WorkspaceReviewPackageStatus`

| value | notes |
| --- | --- |
| `draft` | package が未申請または mixed draft 状態 |
| `pending_review` | current package について pending intake が存在する |
| `changes_requested` | package 内に `revision_requested` object を含む |
| `rejected` | package 内に `rejected` object を含む |
| `approved` | package 内 object が `approved_for_publish` / `approved_for_unlock` で揃っている |

### `WorkspaceReviewSubmitAction`

| value | notes |
| --- | --- |
| `submit` | initial submit を出す |
| `resubmit` | revision requested package に対する resubmit を出す |
| `none` | CTA を表示しない |

- upload 由来の initial review intake は media processing ready 後に自動投入されるため、detail UI は `submitAction = submit` を primary CTA として表示しません。
- `submitAction` は backend の action eligibility を表す互換 field として残します。

### `WorkspaceReviewTargetKind`

| value | notes |
| --- | --- |
| `main` | main detail surface |
| `short` | short detail surface |

### `WorkspaceReviewBlockerCode`

- `linked_short_missing`
- `main_asset_not_ready`
- `short_asset_not_ready`
- `main_price_missing`
- `ownership_missing`
- `consent_missing`

- blocker 語彙は [creator-workspace-submission-review-api-contract.md](creator-workspace-submission-review-api-contract.md) の readiness predicate と同一です。
- backend は blocker を machine-readable code のみで返し、UI 文言は frontend が責務を持ちます。

## Surface-specific Payloads

### `WorkspaceReviewPackageSummary`

| field | type | notes |
| --- | --- | --- |
| `canonicalMainId` | `string` | package anchor となる canonical `main` |
| `reviewStatus` | `WorkspaceReviewPackageStatus` | package 全体の review state |
| `readiness` | `WorkspaceReviewPackageReadiness` | CTA 可否の基礎状態 |
| `submitAction` | `WorkspaceReviewSubmitAction` | backend が判定した action eligibility。detail UI は初回 submit CTA を表示しない |
| `linkedShortCount` | `number` | package に含まれる linked short 数 |
| `blockers` | `WorkspaceReviewBlockerCode[]` | readiness blocker 一覧 |

### `WorkspaceReviewMainItem`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | main identifier |
| `state` | `string` | object-level `main.state` をそのまま返す |

### `WorkspaceReviewShortItem`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | short identifier |
| `canonicalMainId` | `string` | linked canonical main |
| `state` | `string` | object-level `short.state` をそのまま返す |

### `WorkspaceReviewTarget`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | current detail target の object id |
| `kind` | `WorkspaceReviewTargetKind` | `main` or `short` |
| `canonicalMainId` | `string` | submit mutation の path 解決に使う canonical main |

### `WorkspaceReviewTargetState`

| field | type | notes |
| --- | --- | --- |
| `state` | `string` | target object の current review state |
| `reasonCode` | `string \| null` | current target object の review reason code |

## Request Contract

### `GET /api/creator/workspace/review-surface`

#### Response

- `data.reviewSurface.packages`: `WorkspaceReviewPackageSummary[]`
- `data.reviewSurface.mains`: `WorkspaceReviewMainItem[]`
- `data.reviewSurface.shorts`: `WorkspaceReviewShortItem[]`
- `meta.page = null`

### `GET /api/creator/workspace/mains/{mainId}/review-surface`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Response

- `data.reviewSurface.package`: `WorkspaceReviewPackageSummary`
- `data.reviewSurface.target`: `WorkspaceReviewTarget`
- `data.reviewSurface.review`: `WorkspaceReviewTargetState`
- `meta.page = null`

### `GET /api/creator/workspace/shorts/{shortId}/review-surface`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Response

- `data.reviewSurface.package`: `WorkspaceReviewPackageSummary`
- `data.reviewSurface.target`: `WorkspaceReviewTarget`
- `data.reviewSurface.review`: `WorkspaceReviewTargetState`
- `meta.page = null`

## Response Rules

- caller は authenticated viewer である必要があります。
- caller は approved creator capability を持つ owner 自身である必要があります。
- dashboard read surface は owner 自身の `main` / `short` と、それらから導出される package summary のみを返します。
- detail read surface の `target.id` は current detail object を指します。package-level action を実行する場合の anchor は常に `target.canonicalMainId` です。
- `reviewStatus` は package-level summary であり、public publish / unlock state を直接意味しません。
- `target.review.state` と tile `state` は object-level source of truth であり、`reviewStatus` は UI summary 用の derived state です。
- `reasonCode` は opaque code として返し、message localization は backend が持ちません。
- `pending_review` package では `submitAction = none` かつ `readiness = none` を返します。
- `blocked` package では `submitAction = none` を返し、`blockers` に readiness 未成立理由を返します。
- `rejected` object を含む package でも read surface は返しますが、self-serve reopen action は含めません。
- creator UI は `approved` / `pending_review` / blocker のない normal state を通常表示から抑制し、差し戻し、公開不可、blocker、conflict など creator action が必要な状態だけを review panel / tile badge として表示します。dashboard notification は差し戻し、公開不可など review outcome に限定し、pre-submission blocker 専用通知は表示しません。
- 承認済みを示す copy や badge は、正常状態の再説明になるため creator UI には表示しません。

## Success Example

```json
{
  "data": {
    "reviewSurface": {
      "packages": [
        {
          "canonicalMainId": "main_aoi_blue_balcony",
          "reviewStatus": "changes_requested",
          "readiness": "ready",
          "submitAction": "resubmit",
          "linkedShortCount": 2,
          "blockers": []
        },
        {
          "canonicalMainId": "main_sora_after_rain",
          "reviewStatus": "pending_review",
          "readiness": "none",
          "submitAction": "none",
          "linkedShortCount": 1,
          "blockers": []
        }
      ],
      "mains": [
        {
          "id": "main_aoi_blue_balcony",
          "state": "revision_requested"
        },
        {
          "id": "main_sora_after_rain",
          "state": "pending_review"
        }
      ],
      "shorts": [
        {
          "id": "short_aoi_balcony_cut",
          "canonicalMainId": "main_aoi_blue_balcony",
          "state": "approved_for_publish"
        },
        {
          "id": "short_aoi_balcony_close",
          "canonicalMainId": "main_aoi_blue_balcony",
          "state": "revision_requested"
        }
      ]
    }
  },
  "meta": {
    "requestId": "req_creator_workspace_review_surface_001",
    "page": null
  },
  "error": null
}
```

## Error Contract

| status | code | notes |
| --- | --- | --- |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | owner 自身の review surface target を解決できない |
| `500` | `internal_error` | unexpected failure |

## Guardrails

- overview metrics、analytics、gross revenue、unlock count は返しません。
- owner preview detail payload をこの endpoint に混ぜません。
- submit / resubmit mutation をこの read response で代替しません。
- public creator profile や fan playback への read surface と兼用しません。
- `review reason` の localized message を backend response に持ち込みません。

## Fixture Reference

- representative fixture は [creator-workspace-review-surface.json](fixtures/creator-workspace-review-surface.json) を参照します。
