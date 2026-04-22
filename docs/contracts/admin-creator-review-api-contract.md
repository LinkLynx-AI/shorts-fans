# Admin Creator Review API Contract

## 位置づけ

- この文書は `SHO-175 admin creator review UI` で使う local admin transport 契約を固定します。
- 対象は `creator registration` の review queue / review detail / decision mutation に限定します。
- UI surface は `frontend` の `/admin/creator-reviews` を前提にしますが、この文書自体は backend transport の payload と state rule を定義します。

## Goals

- submitted creator registration を admin queue から取得できるようにする。
- creator が submit した shared profile / intake / evidence を detail で確認できるようにする。
- approve / reject(reason required) / suspend を state machine に沿って反映できるようにする。

## Non-goals

- production admin 認証
- reviewer assignment / audit log / comment thread
- evidence binary proxy や inline preview
- reject metadata の canonical reason vocabulary 固定

## Canonical Sources

- `docs/contracts/viewer-creator-entry-api-contract.md`
- `docs/contracts/viewer-creator-registration-intake-api-contract.md`
- `docs/contracts/viewer-profile-api-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/moderation/moderation-and-review.md`

## Environment Boundary

- endpoint は `backend` が `development` 環境で起動しているときだけ有効です。
- local admin 向けの暫定 surface として、production auth / RBAC は要求しません。
- backend 側では `loopback remote addr` と `X-Shorts-Fans-Admin-Token` (`ADMIN_API_TOKEN`) が一致しない request を `404` で遮断します。
- frontend の local admin UI は `X-Shorts-Fans-Admin-UI-Token` (`ADMIN_UI_ACCESS_TOKEN`) で発行した署名済み httpOnly cookie を持つ request だけを許可します。
- local admin UI の cookie は `POST /admin/access` から発行します。access token を URL query や cookie value へ直接保存してはいけません。
- frontend の local admin UI は `localhost:3001` で動かし、backend API は `NEXT_PUBLIC_API_BASE_URL` が指す先を利用します。

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/admin/creator-reviews?state=submitted` | dev admin token | state ごとの review queue を返す |
| `GET` | `/api/admin/creator-reviews/:userId` | dev admin token | user 単位の review detail を返す |
| `POST` | `/api/admin/creator-reviews/:userId/decision` | dev admin token | review decision を反映して更新後 case を返す |

## Shared Rules

- `state` query の allowed values は `submitted`、`approved`、`rejected`、`suspended` です。
- queue item は `sharedProfile + creatorBio + legalName + review timestamps` だけを返し、evidence detail は detail endpoint に分離します。
- detail endpoint は evidence ごとに signed GET URL を返します。binary 自体は response body に含めません。
- `reasonCode` は opaque string として扱い、この contract では固定語彙を定義しません。
- decision mutation の `reasonCode` は `rejected` のとき必須、それ以外では空である前提です。
- decision transition は以下だけを許可します。
  - `submitted -> approved`
  - `submitted -> rejected`
  - `approved -> suspended`

## `GET /api/admin/creator-reviews`

### Request

- query parameter `state` が未指定なら `submitted` を使います。

### Success

```json
{
  "data": {
    "state": "submitted",
    "items": [
      {
        "userId": "11111111-1111-1111-1111-111111111111",
        "state": "submitted",
        "creatorBio": "quiet rooftop と low light preview を中心に投稿予定です。",
        "legalName": "Mina Rei",
        "review": {
          "submittedAt": "2026-04-17T08:00:00Z",
          "approvedAt": null,
          "rejectedAt": null,
          "suspendedAt": null
        },
        "sharedProfile": {
          "displayName": "Mina Rei",
          "handle": "@minarei_review",
          "avatar": {
            "id": "asset_viewer_profile_avatar_11111111111111111111111111111111",
            "kind": "image",
            "url": "https://cdn.example.com/mock/review/mina-rei-avatar.jpg",
            "posterUrl": null,
            "durationSeconds": null
          }
        }
      }
    ]
  },
  "meta": {
    "requestId": "req_admin_creator_review_queue_get_001",
    "page": null
  },
  "error": null
}
```

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_review_state` | unknown `state` query |
| `500` | `internal_error` | unexpected failure |

## `GET /api/admin/creator-reviews/:userId`

### Success

```json
{
  "data": {
    "case": {
      "userId": "11111111-1111-1111-1111-111111111111",
      "state": "submitted",
      "creatorBio": "quiet rooftop と low light preview を中心に投稿予定です。",
      "sharedProfile": {
        "displayName": "Mina Rei",
        "handle": "@minarei_review",
        "avatar": {
          "id": "asset_viewer_profile_avatar_11111111111111111111111111111111",
          "kind": "image",
          "url": "https://cdn.example.com/mock/review/mina-rei-avatar.jpg",
          "posterUrl": null,
          "durationSeconds": null
        }
      },
      "intake": {
        "legalName": "Mina Rei",
        "birthDate": "1999-04-02",
        "legalAddress": "東京都渋谷区...",
        "identityDocumentType": "driver_license",
        "targetAudienceCategory": "general_adult",
        "hasCoPerformers": false,
        "payoutRecipientType": "self",
        "payoutRecipientName": "Mina Rei",
        "declaresNoProhibitedCategory": true,
        "acceptsConsentResponsibility": true,
        "acceptsAppearanceVerification": true,
        "acceptsCoPerformerConsentResponsibility": false,
        "acceptsAdultBusinessCompliance": true,
        "confirmsInformationMatchesDocuments": true
      },
      "evidences": [
        {
          "kind": "government_id",
          "fileName": "government-id.png",
          "mimeType": "image/png",
          "fileSizeBytes": 183442,
          "uploadedAt": "2026-04-17T07:45:00Z",
          "accessUrl": "https://signed.example.com/mock/government-id"
        },
        {
          "kind": "identity_selfie",
          "fileName": "selfie.png",
          "mimeType": "image/png",
          "fileSizeBytes": 102400,
          "uploadedAt": "2026-04-17T07:46:00Z",
          "accessUrl": "https://signed.example.com/mock/identity-selfie"
        },
        {
          "kind": "address_proof",
          "fileName": "address-proof.pdf",
          "mimeType": "application/pdf",
          "fileSizeBytes": 84512,
          "uploadedAt": "2026-04-17T07:46:30Z",
          "accessUrl": "https://signed.example.com/mock/address-proof"
        },
        {
          "kind": "payout_proof",
          "fileName": "payout-proof.pdf",
          "mimeType": "application/pdf",
          "fileSizeBytes": 84512,
          "uploadedAt": "2026-04-17T07:45:00Z",
          "accessUrl": "https://signed.example.com/mock/payout-proof"
        }
      ],
      "review": {
        "submittedAt": "2026-04-17T08:00:00Z",
        "approvedAt": null,
        "rejectedAt": null,
        "suspendedAt": null
      },
      "rejection": null
    }
  },
  "meta": {
    "requestId": "req_admin_creator_review_case_get_001",
    "page": null
  },
  "error": null
}
```

### Field Rules

| field | type | notes |
| --- | --- | --- |
| `sharedProfile` | `ViewerProfilePreview` | shared viewer profile の current preview |
| `creatorBio` | `string` | creator 固有 bio |
| `intake` | `AdminCreatorReviewIntake` | submit 済み intake snapshot。銀行口座情報は含めない |
| `evidences` | `AdminCreatorReviewEvidence[]` | signed GET URL 付き evidence summary |
| `review` | `ReviewTimeline` | capability timestamps |
| `rejection` | `Rejection \| null` | rejected state の metadata |

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | `userId` が UUID でない |
| `404` | `not_found` | 対象 user の review case が存在しない |
| `500` | `internal_error` | unexpected failure |

## `POST /api/admin/creator-reviews/:userId/decision`

### Request

```json
{
  "decision": "rejected",
  "reasonCode": "documents_blurry",
  "isResubmitEligible": false,
  "isSupportReviewRequired": false
}
```

### Decision Input Rules

| field | type | notes |
| --- | --- | --- |
| `decision` | `"approved" \| "rejected" \| "suspended"` | next state |
| `reasonCode` | `string` | `rejected` のとき必須 |
| `isResubmitEligible` | `boolean` | `rejected` metadata |
| `isSupportReviewRequired` | `boolean` | `rejected` metadata |

- `isResubmitEligible` と `isSupportReviewRequired` を同時に `true` にはできません。
- `approved` / `suspended` では `reasonCode` は空である前提です。

### Success

- status は `200 OK`
- response body は `GET /api/admin/creator-reviews/:userId` と同じ `case` shape を返します。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_review_decision` | unknown decision |
| `400` | `review_reason_required` | `rejected` に reasonCode が不足 |
| `400` | `review_decision_metadata_conflict` | rejected metadata が矛盾 |
| `404` | `not_found` | 対象 user の review case が存在しない |
| `409` | `registration_incomplete` | `approved` に必要な intake / evidence が不足 |
| `409` | `review_state_conflict` | current state ではその decision を適用できない |
| `500` | `internal_error` | unexpected failure |

## Boundary Guardrails

- admin queue は creator registration review 以外の moderation queue を返しません。
- signed evidence URL は short-lived access のみを目的とし、public URL として再利用しません。
- この surface は `creatorregistration` package の state machine を再定義しません。
- auth / RBAC はこの leaf の対象外であり、production admin surface にそのまま拡張しません。
