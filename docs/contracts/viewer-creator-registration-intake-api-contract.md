# Viewer Creator Registration Intake API Contract

## 位置づけ

- この文書は `SHO-172 creator審査申請 intake と必要証跡収集を実装する` の transport 契約を固定します。
- fan profile から始める creator onboarding のうち、`draft 保存`、`required evidence upload`、`submit 前 completeness` を扱います。
- state machine 自体は引き続き `docs/contracts/viewer-creator-entry-api-contract.md` を正とし、この文書は concrete payload と file boundary だけを定義します。

## Goals

- shared viewer profile を preview しながら creator 固有の `bio` と審査 intake を保存できるようにする。
- `government_id` と `payout_proof` の private evidence upload を 1 kind = 1 active file で扱えるようにする。
- `POST /api/viewer/creator-registration` が completeness を判定できるよう、required fields と validation boundary を固定する。

## Non-goals

- reviewer decision UI
- shared viewer profile の更新
- evidence の public URL 発行
- KYC provider / payout provider への外部送信

## Canonical Sources

- `docs/contracts/viewer-creator-entry-api-contract.md`
- `docs/contracts/viewer-profile-api-contract.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/creator/creator-onboarding-surface.md`
- `docs/ssot/product/moderation/moderation-and-review.md`

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/viewer/creator-registration/intake` | required | current viewer の intake draft と evidence summary を返す |
| `PUT` | `/api/viewer/creator-registration/intake` | required | current viewer の intake draft を保存する |
| `POST` | `/api/viewer/creator-registration/evidence-uploads` | required | private evidence upload target を発行する |
| `POST` | `/api/viewer/creator-registration/evidence-uploads/complete` | required | uploaded evidence を検証して intake に紐づける |

## Shared Rules

- `displayName / handle / avatar` は shared viewer profile preview として返すだけで、この surface から更新しません。
- shared viewer profile を修正したい場合は `docs/contracts/viewer-profile-api-contract.md` の `/api/viewer/profile` を使います。
- `bio` は creator 固有 draft としてこの surface で編集できます。
- intake / evidence の編集は `draft` と eligible な `rejected` だけに許可します。
- `submitted / approved / suspended`、および `rejected` でも `isReadOnly=true` / `isSupportReviewRequired=true` な case は `409 registration_state_conflict` を返します。
- evidence object は private bucket に保存し、この contract では download URL や public asset URL を返しません。
- required evidence kind は `government_id` と `payout_proof` の 2 種類です。
- allowed mime type は `image/jpeg`、`image/png`、`image/webp`、`application/pdf` です。
- file size 上限は `10MB` です。

## `GET /api/viewer/creator-registration/intake`

### Success

- status は `200 OK`
- onboarding case 未開始でも current shared profile preview と empty draft を返します。

```json
{
  "data": {
    "intake": {
      "sharedProfile": {
        "avatar": {
          "id": "asset_viewer_profile_avatar_minarei",
          "kind": "image",
          "url": "https://cdn.example.com/avatars/mina.jpg",
          "posterUrl": null,
          "durationSeconds": null
        },
        "displayName": "Mina Rei",
        "handle": "@minarei"
      },
      "creatorBio": "quiet rooftop の continuation を中心に投稿します。",
      "legalName": "Mina Rei",
      "birthDate": "1999-04-02",
      "payoutRecipientType": "self",
      "payoutRecipientName": "Mina Rei",
      "declaresNoProhibitedCategory": true,
      "acceptsConsentResponsibility": true,
      "registrationState": "draft",
      "isReadOnly": false,
      "canSubmit": true,
      "evidences": [
        {
          "kind": "government_id",
          "fileName": "government-id.png",
          "mimeType": "image/png",
          "fileSizeBytes": 183442,
          "uploadedAt": "2026-04-17T01:15:00Z"
        },
        {
          "kind": "payout_proof",
          "fileName": "bank-proof.pdf",
          "mimeType": "application/pdf",
          "fileSizeBytes": 84512,
          "uploadedAt": "2026-04-17T01:17:00Z"
        }
      ]
    }
  },
  "meta": {
    "requestId": "req_viewer_creator_registration_intake_get_001",
    "page": null
  },
  "error": null
}
```

### Field Rules

| field | type | notes |
| --- | --- | --- |
| `sharedProfile` | `ViewerProfilePreview` | shared viewer profile の preview。編集不可 |
| `creatorBio` | `string` | creator 固有の bio draft |
| `legalName` | `string` | 本人確認に使う氏名 |
| `birthDate` | `string \| null` | `YYYY-MM-DD` |
| `payoutRecipientType` | `"self" \| "business" \| null` | 売上受取名義の種別 |
| `payoutRecipientName` | `string` | 売上受取名義 |
| `declaresNoProhibitedCategory` | `boolean` | prohibited category 非該当確認 |
| `acceptsConsentResponsibility` | `boolean` | consent / ownership responsibility 確認 |
| `registrationState` | `string \| null` | onboarding case 未開始なら `null` |
| `isReadOnly` | `boolean` | `draft` と eligible `rejected` は `false`、それ以外は `true` |
| `canSubmit` | `boolean` | required fields + required evidence が揃っていて editable なときだけ `true` |
| `evidences` | `CreatorRegistrationEvidence[]` | current draft に紐づく uploaded evidence summary |

## `PUT /api/viewer/creator-registration/intake`

### Request

```json
{
  "creatorBio": "quiet rooftop の continuation を中心に投稿します。",
  "legalName": "Mina Rei",
  "birthDate": "1999-04-02",
  "payoutRecipientType": "self",
  "payoutRecipientName": "Mina Rei",
  "declaresNoProhibitedCategory": true,
  "acceptsConsentResponsibility": true
}
```

### Success

- status は `200 OK`
- response body は `GET` と同じ `intake` shape を返します。
- 初回 save で onboarding case が未開始なら server は `draft` case を作成してよいものとします。
- eligible な `rejected` の場合は resubmit 用の修正保存として同じ endpoint を使います。

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_legal_name` | legal name が不正 |
| `400` | `invalid_birth_date` | `birthDate` が `YYYY-MM-DD` でない |
| `400` | `invalid_payout_recipient_type` | `self / business` 以外 |
| `400` | `invalid_payout_recipient_name` | payout recipient name が不正 |
| `401` | `auth_required` | session 不在 |
| `404` | `not_found` | shared viewer profile 不在 |
| `409` | `registration_state_conflict` | editable でない state から save を要求 |
| `500` | `internal_error` | unexpected failure |

## `POST /api/viewer/creator-registration/evidence-uploads`

### Request

```json
{
  "kind": "government_id",
  "fileName": "government-id.png",
  "mimeType": "image/png",
  "fileSizeBytes": 183442
}
```

### Success

- onboarding case 未開始なら、server は upload target 発行前に `draft` case を初期化してよいものとします。
- eligible な `rejected` では resubmit 用の証跡差し替えとして同じ endpoint を使います。

```json
{
  "data": {
    "evidenceKind": "government_id",
    "evidenceUploadToken": "vcevd_b9f28e5ce0c24f0db969e0b9d4c81d4e",
    "expiresAt": "2026-04-17T01:30:00Z",
    "uploadTarget": {
      "fileName": "government-id.png",
      "mimeType": "image/png",
      "upload": {
        "method": "PUT",
        "url": "https://bucket.s3.ap-northeast-1.amazonaws.com/...",
        "headers": {
          "Content-Type": "image/png"
        }
      }
    }
  },
  "meta": {
    "requestId": "req_viewer_creator_registration_evidence_upload_create_001",
    "page": null
  },
  "error": null
}
```

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `invalid_evidence_kind` | unknown evidence kind |
| `400` | `invalid_evidence_mime_type` | unsupported mime type |
| `400` | `invalid_evidence_file_size` | zero / negative size |
| `400` | `evidence_file_too_large` | `10MB` 超過 |
| `401` | `auth_required` | session 不在 |
| `404` | `not_found` | shared viewer profile 不在 |
| `409` | `registration_state_conflict` | editable でない state から evidence 差し替えを要求 |
| `500` | `internal_error` | unexpected failure |

## `POST /api/viewer/creator-registration/evidence-uploads/complete`

### Request

```json
{
  "evidenceUploadToken": "vcevd_b9f28e5ce0c24f0db969e0b9d4c81d4e"
}
```

### Success

```json
{
  "data": {
    "evidenceKind": "government_id",
    "evidenceUploadToken": "vcevd_b9f28e5ce0c24f0db969e0b9d4c81d4e",
    "evidence": {
      "kind": "government_id",
      "fileName": "government-id.png",
      "mimeType": "image/png",
      "fileSizeBytes": 183442,
      "uploadedAt": "2026-04-17T01:15:00Z"
    }
  },
  "meta": {
    "requestId": "req_viewer_creator_registration_evidence_upload_complete_001",
    "page": null
  },
  "error": null
}
```

### Error Contract

| status | code | notes |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `401` | `auth_required` | session 不在 |
| `404` | `evidence_upload_not_found` | token 不在、または owner mismatch |
| `409` | `evidence_upload_incomplete` | object 不足、mime mismatch、size mismatch |
| `409` | `evidence_upload_expired` | token 期限切れ |
| `409` | `registration_state_conflict` | editable でない state から evidence 差し替えを要求 |
| `500` | `internal_error` | unexpected failure |

## Completeness Rule

- `POST /api/viewer/creator-registration` が success する前提は次のとおりです。
- shared viewer profile の `displayName / handle` が存在する
- `creatorBio` が空でない
- `legalName` が空でない
- `birthDate` が存在する
- `payoutRecipientType / payoutRecipientName` が存在する
- `declaresNoProhibitedCategory = true`
- `acceptsConsentResponsibility = true`
- `government_id` と `payout_proof` の両 evidence が存在する
