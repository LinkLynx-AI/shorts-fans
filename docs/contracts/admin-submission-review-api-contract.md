# Admin Submission Review API Contract

## 位置づけ

- この文書は `SHO-197 submission package admin review UI` で使う local admin transport 契約を固定します。
- 対象は `submission_review_intakes` の queue / detail / decision mutation に限定します。
- UI surface は `frontend` の `/admin/submission-reviews` を前提にしますが、この文書自体は backend transport の payload と state rule を定義します。

## Goals

- `pending_review` intake を local admin queue から確認できるようにする。
- intake detail で creator summary、canonical `main`、linked `shorts`、preview media、review provenance を確認できるようにする。
- reviewer が `main` と `pending short` ごとに `approved / revision_requested / rejected` を反映できるようにする。
- `revision_requested` / `rejected` では `reasonCode` を必須にし、全 decision で `reviewNote` を任意保存できるようにする。

## Non-goals

- production admin auth / RBAC
- queue assignment、bulk action、approval analytics
- creator-facing review status UI
- auto review provider integration
- moderation history browser

## Canonical Sources

- `docs/contracts/submission-package-review-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/contracts/creator-workspace-owner-preview-api-contract.md`
- `docs/ssot/product/moderation/moderation-and-review.md`

## Environment Boundary

- endpoint は `backend` が `development` 環境で起動しているときだけ有効です。
- local admin 向けの暫定 surface として、現時点では auth を要求しません。
- backend 側では `loopback remote addr` 以外からの access を `404` で遮断します。
- preview media は raw storage ref ではなく、owner preview と同じ signed display asset を返します。

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/admin/submission-reviews` | none | `pending_review` intake queue を返す |
| `GET` | `/api/admin/submission-reviews/:intakeId` | none | intake 単位の review detail を返す |
| `POST` | `/api/admin/submission-reviews/:intakeId/decision` | none | object-level decision を反映し、`204 No Content` を返す |

## Shared Rules

- queue は `pending_review` intake だけを返します。
- detail は `pending_review` と `decision_applied` の両方を返せます。history browse の簡易入口として decision log を確認するためです。
- decision input は `main` と `short` を package-level に潰さず、object-level で送ります。
- `approved` は `main -> approved_for_unlock`、`short -> approved_for_publish` に map されます。
- `revision_requested` / `rejected` は `reasonCode` 必須です。
- `reviewNote` は全 decision で任意です。空文字は未設定として扱います。
- UI は `manual_override` を出しません。request の `decisionSource` 未指定時は backend 側で `manual` を使います。
- detail payload の media asset / caption / price は intake snapshot を source of truth にし、current state / provenance だけを current object から読みます。
- `decisionRequired = true` になるのは `intake.status = pending_review` かつ current object state も `pending_review` のときだけです。

## `GET /api/admin/submission-reviews`

### Success

```json
{
  "data": {
    "items": [
      {
        "intakeId": "11111111-1111-1111-1111-111111111111",
        "submitKind": "initial_submit",
        "submittedAt": "2026-04-20T09:00:00Z",
        "mainDecisionRequired": true,
        "pendingShortCount": 2,
        "shortCount": 3,
        "creator": {
          "id": "creator_11111111111111111111111111111111",
          "displayName": "Mina Rei",
          "handle": "@minarei_review",
          "bio": "quiet rooftop と low light preview を中心に投稿予定です。",
          "avatar": null
        }
      }
    ]
  },
  "meta": {
    "requestId": "req_admin_submission_review_queue_get_001",
    "page": null
  },
  "error": null
}
```

### Field Rules

| field | type | notes |
| --- | --- | --- |
| `mainDecisionRequired` | `boolean` | current `main.state = pending_review` か |
| `pendingShortCount` | `number` | current `short.state = pending_review` の本数 |
| `shortCount` | `number` | intake snapshot に含まれる short の総数 |

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `500` | `internal_error` | unexpected failure |

## `GET /api/admin/submission-reviews/:intakeId`

### Success

```json
{
  "data": {
    "case": {
      "creator": {
        "id": "creator_11111111111111111111111111111111",
        "displayName": "Mina Rei",
        "handle": "@minarei_review",
        "bio": "quiet rooftop と low light preview を中心に投稿予定です。",
        "avatar": null
      },
      "intake": {
        "id": "11111111-1111-1111-1111-111111111111",
        "status": "pending_review",
        "submitKind": "initial_submit",
        "submittedAt": "2026-04-20T09:00:00Z",
        "previousIntakeId": null,
        "creatorUserId": "12111111-1111-1111-1111-111111111111",
        "canonicalMainId": "22222222-2222-2222-2222-222222222222",
        "mainMediaAssetId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
        "mainPriceJpy": 1800,
        "ownershipConfirmed": true,
        "consentConfirmed": true
      },
      "main": {
        "id": "22222222-2222-2222-2222-222222222222",
        "state": "pending_review",
        "decisionRequired": true,
        "priceJpy": 1800,
        "currencyCode": "JPY",
        "media": {
          "id": "asset_main_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
          "kind": "video",
          "url": "https://signed.example.com/main-playback",
          "posterUrl": "https://signed.example.com/main-poster",
          "durationSeconds": 540
        },
        "review": {
          "decisionSource": null,
          "decisionedAt": null,
          "reasonCode": null,
          "reviewNote": null
        },
        "intakeDecisionLog": null
      },
      "shorts": [
        {
          "id": "33333333-3333-3333-3333-333333333333",
          "caption": "quiet rooftop cut",
          "state": "pending_review",
          "decisionRequired": true,
          "media": {
            "id": "asset_short_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
            "kind": "video",
            "url": "https://signed.example.com/short-playback",
            "posterUrl": "https://signed.example.com/short-poster",
            "durationSeconds": 18
          },
          "review": {
            "decisionSource": null,
            "decisionedAt": null,
            "reasonCode": null,
            "reviewNote": null
          },
          "intakeDecisionLog": null
        }
      ]
    }
  },
  "meta": {
    "requestId": "req_admin_submission_review_case_get_001",
    "page": null
  },
  "error": null
}
```

### Field Rules

| field | type | notes |
| --- | --- | --- |
| `review` | `ReviewProvenance` | current object state に残っている provenance |
| `intakeDecisionLog` | `IntakeDecisionLog \| null` | 対象 intake で作られた decision log row |
| `media` | `video asset` | intake snapshot media asset から解決した signed playback / poster pair |
| `caption` | `string \| null` | intake snapshot caption |
| `priceJpy` | `number` | intake snapshot main price |

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | `intakeId` が UUID でない |
| `404` | `not_found` | 対象 intake が存在しない |
| `500` | `internal_error` | unexpected failure |

## `POST /api/admin/submission-reviews/:intakeId/decision`

### Request

```json
{
  "mainDecision": {
    "decision": "approved",
    "reasonCode": "",
    "reviewNote": "unlock ready"
  },
  "shortDecisions": [
    {
      "shortId": "33333333-3333-3333-3333-333333333333",
      "decision": "revision_requested",
      "reasonCode": "quality_issue",
      "reviewNote": "reframe intro"
    }
  ]
}
```

### Decision Input Rules

| field | type | notes |
| --- | --- | --- |
| `decisionSource` | `"manual" \| "auto" \| "manual_override"` | optional。未指定なら `manual` |
| `mainDecision` | `DecisionInput \| null` | current `main.state = pending_review` のときだけ送る |
| `shortDecisions[]` | `ShortDecisionInput[]` | current `short.state = pending_review` の short をすべて含める |
| `reasonCode` | `string` | `revision_requested` / `rejected` のとき必須 |
| `reviewNote` | `string` | 任意。空文字は未設定扱い |

- `POST` 成功時は status `204 No Content` を返します。detail refresh は別途 `GET /api/admin/submission-reviews/:intakeId` で行います。
- `approved` は object kind に応じて `approved_for_unlock` / `approved_for_publish` に map されます。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / invalid UUID |
| `400` | `invalid_review_decision` | unknown decision |
| `400` | `invalid_review_decision_source` | unknown decision source |
| `400` | `review_reason_required` | `revision_requested` / `rejected` に reasonCode が不足 |
| `400` | `review_decision_metadata_conflict` | approved に reasonCode を渡した等の矛盾 |
| `404` | `not_found` | pending intake が存在しない |
| `409` | `review_targets_mismatch` | pending target の集合と request が一致しない |
| `409` | `review_state_conflict` | current state では decision を適用できない |
| `500` | `internal_error` | unexpected failure |

## Boundary Guardrails

- queue は submission review intake 以外の moderation queue を返しません。
- detail は raw storage ref を返さず、signed display asset だけを返します。
- package-level intake と object-level decision の境界はこの surface でも崩しません。
- この surface は `submissionreview` package の state machine を再定義しません。
