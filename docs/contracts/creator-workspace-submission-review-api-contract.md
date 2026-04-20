# Creator Workspace Submission Review API Contract

## 位置づけ

- この文書は creator workspace 内で owner 自身が canonical `main` anchor の current `submission package` を review submit / resubmit する private mutation contract を固定します。
- `submission package ready` predicate と object-level review state rule は [submission-package-review-contract.md](submission-package-review-contract.md) を参照し、この文書では transport と error vocabulary を固定します。
- creator-facing の readiness / status read surface はこの contract に含めず、後続 issue に分離します。

## Goals

- approved creator owner が current package を `review submit` できる最小 mutation surface を固定する。
- ready ではない package、owner 不一致 package、current review state から submit 不可な package を transport error として区別する。
- success 時に response body を返さない `204 No Content` contract を固定する。

## Non-goals

- `submission package ready` の詳細理由を response body で返すこと
- review decision queue / detail / apply decision transport
- creator-facing readiness / submission status read API
- `rejected` object の reopen transport
- main / short preview detail payload の再取得

## Canonical Sources

- [submission-package-review-contract.md](submission-package-review-contract.md)
- [creator-workspace-owner-preview-api-contract.md](creator-workspace-owner-preview-api-contract.md)
- [mvp-core-domain-contract.md](mvp-core-domain-contract.md)
- [mvp-media-workflow-contract.md](mvp-media-workflow-contract.md)
- [fan-mvp-common-transport-contract.md](fan-mvp-common-transport-contract.md)

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `POST` | `/api/creator/workspace/mains/{mainId}/review-submissions` | required | owner 自身の current package を review submit / resubmit |

## Request Contract

### `POST /api/creator/workspace/mains/{mainId}/review-submissions`

#### Path

| field | type | required |
| --- | --- | --- |
| `mainId` | `string` | yes |

#### Body

- request body は受け取りません。
- caller は submit 対象 package の current snapshot を path の `mainId` だけで指定します。

#### Response

- success は `204 No Content` を返します。
- response body は返しません。

## Response Rules

- caller は authenticated viewer である必要があります。
- caller は approved creator capability を持つ owner 自身である必要があります。
- submit 対象は canonical `main` を anchor に解決した current `submission package` です。
- initial submit は ready な all-draft package に限ります。
- resubmit は `decision_applied` 済み intake を持ち、かつ package 内に `revision_requested` object を含む package に限ります。
- approved 済み object は resubmit で自動的に `pending_review` へ戻しません。
- duplicated submit retry で、すでに pending intake が存在する場合は no-op success として扱います。
- success response では intake id、ready blockers、object state summary を返しません。

## Error Contract

| status | code | notes |
| --- | --- | --- |
| `401` | `auth_required` | session 不在 |
| `403` | `creator_mode_unavailable` | approved creator capability なし |
| `404` | `not_found` | owner 自身の submit 対象 package を解決できない、または `mainId` が無効 |
| `409` | `review_state_conflict` | current review state から submit / resubmit できない |
| `422` | `submission_not_ready` | package ready 条件未成立 |
| `500` | `internal_error` | unexpected failure |

## Guardrails

- request body で short set、price、consent、ownership、review state を直接上書きしません。
- response に review intake snapshot や internal asset ref を含めません。
- owner preview read contract をこの mutation response で兼用しません。
- `rejected` object の self-serve reopen path は含めません。
