# Creator Upload API Contract

## 位置づけ

- この文書は `SHO-141 creator upload flowのcontractとfixtureを定義する` の成果物です。
- creator-private な `new-package` upload flow に限定して、raw upload initiation と upload completion の transport 契約を固定します。
- mock_ui にある `link-short`、upload page UI、backend endpoint 実装、frontend からの実接続はこの文書の対象外です。

## Goals

- approved creator が `main` 1 本と `short` 1 本以上をまとめて upload 開始できる transport boundary を固定する。
- browser / client が raw bucket へ `Presigned PUT` で直接 upload できる request / response shape を固定する。
- upload 完了後に draft `main` と draft `shorts` を永続化する completion boundary を固定する。

## Non-goals

- `link-short` で既存 `main` に `short` を追加する flow
- upload page UI、route、local state
- media processing worker、retry queue、delivery materialization
- review submit、publish / unlock eligibility、linkage editor
- raw bucket CORS や infra 実装の詳細

## Canonical Sources

- `docs/contracts/mvp-core-domain-contract.md`
- `docs/contracts/mvp-media-workflow-contract.md`
- `docs/contracts/submission-package-review-contract.md`
- `docs/infra/dev-media-sandbox.md`
- `docs/ssot/product/creator/creator-workflow.md`
- `docs/ssot/product/content/short-main-linkage.md`
- `mock_ui/app.js`

## Flow Scope

- 対象は mock_ui の `new-package` flow だけです。
- creator は `main` 1 本と `short` 1 本以上を同時に選択し、揃うまで submit できません。
- raw upload 受理は `publishable`、`unlockable`、`review-ready` を意味しません。
- `main` と `short` は別 asset として upload されますが、completion は package 単位で成功または失敗します。
- completion 成功後の content は draft のままであり、review submit は別 boundary です。

## Vocabulary

### `upload package`

- creator-private な upload grouping 単位です。
- public resource ではなく、review intake 用 `submission package` そのものでもありません。
- `packageToken` は opaque token として扱い、client は中身を解釈しません。

### `upload entry`

- package 内の個別 file target です。
- `main` は常に 1 件、`shorts` は 1 件以上を持ちます。
- `uploadEntryId` は `media_assets.external_upload_ref` に流用できる一意識別子として扱います。

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `POST` | `/api/creator/upload-packages` | required + approved creator | `main` / `shorts` metadata を受けて presigned PUT target を返す |
| `POST` | `/api/creator/upload-packages/complete` | required + approved creator | upload 済み package を draft `main` / `shorts` として確定する |

## Shared Rules

- authenticated viewer であり、かつ `creator capability = approved` の user だけが使えます。
- `activeMode = creator` は必須条件にしません。権限制御は creator capability で行います。
- request / response は backend 既存契約と同じ `data / meta / error` envelope を使い、`meta.page` は常に `null` です。
- JSON request は unknown field を許可しません。
- v1 では `main` と `shorts` の `mimeType` はすべて `video/*` に固定します。
- direct upload は `Presigned PUT` を使い、multipart form は使いません。
- client は `upload-packages` response に含まれる URL と header をそのまま使い、bucket / key を推測しません。
- `link-short`、`main` なし upload、`shorts` 0 件 upload は v1 では扱いません。

## Request / Response Contract

### File Metadata

| field | type | required | notes |
| --- | --- | --- | --- |
| `fileName` | `string` | yes | trim 後 non-empty |
| `mimeType` | `string` | yes | `video/` prefix 必須 |
| `fileSizeBytes` | `number` | yes | `> 0` |

### Upload Target

| field | type | notes |
| --- | --- | --- |
| `uploadEntryId` | `string` | package 内で一意 |
| `role` | `"main" \| "short"` | upload role |
| `fileName` | `string` | request で受けた file name |
| `mimeType` | `string` | request で受けた MIME |
| `upload.method` | `"PUT"` | direct upload method |
| `upload.url` | `string` | presigned PUT URL |
| `upload.headers` | `Record<string, string>` | client が PUT 時にそのまま送る header |

### Created Draft Content

| field | type | notes |
| --- | --- | --- |
| `main.id` | `string` | created draft main ID |
| `main.state` | `"draft"` | upload completion 時点では固定 |
| `main.mediaAsset.id` | `string` | created `media_assets` row |
| `main.mediaAsset.processingState` | `"uploaded"` | raw object 受理直後の state |
| `shorts[].id` | `string` | created draft short ID |
| `shorts[].state` | `"draft"` | upload completion 時点では固定 |
| `shorts[].canonicalMainId` | `string` | created main ID |
| `shorts[].mediaAsset.id` | `string` | created `media_assets` row |
| `shorts[].mediaAsset.processingState` | `"uploaded"` | raw object 受理直後の state |

## `POST /api/creator/upload-packages`

### Request

```json
{
  "main": {
    "fileName": "quiet-rooftop-main.mp4",
    "mimeType": "video/mp4",
    "fileSizeBytes": 184320041
  },
  "shorts": [
    {
      "fileName": "quiet-rooftop-short-a.mp4",
      "mimeType": "video/mp4",
      "fileSizeBytes": 24182019
    },
    {
      "fileName": "quiet-rooftop-short-b.mp4",
      "mimeType": "video/mp4",
      "fileSizeBytes": 21944882
    }
  ]
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `main` | `FileMetadata` | yes | `main` は常に 1 件 |
| `shorts` | `FileMetadata[]` | yes | 1 件以上必須 |

### Success

- status は `200 OK`
- body は package token と upload target 一覧を返します
- `expiresAt` を過ぎた target は再利用できません

```json
{
  "data": {
    "expiresAt": "2026-04-08T12:15:00Z",
    "packageToken": "cupkg_01hrp6wjkq7mh6f3d2f6c5j8rz",
    "uploadTargets": {
      "main": {
        "fileName": "quiet-rooftop-main.mp4",
        "mimeType": "video/mp4",
        "role": "main",
        "upload": {
          "headers": {
            "Content-Type": "video/mp4"
          },
          "method": "PUT",
          "url": "https://raw-bucket.example.com/presigned/main"
        },
        "uploadEntryId": "cu_main_01hrp6wjv9h2k4g1b8s0k5e3pf"
      },
      "shorts": [
        {
          "fileName": "quiet-rooftop-short-a.mp4",
          "mimeType": "video/mp4",
          "role": "short",
          "upload": {
            "headers": {
              "Content-Type": "video/mp4"
            },
            "method": "PUT",
            "url": "https://raw-bucket.example.com/presigned/short-a"
          },
          "uploadEntryId": "cu_short_01hrp6wk4m2q7cxnq0fxsy2f55"
        },
        {
          "fileName": "quiet-rooftop-short-b.mp4",
          "mimeType": "video/mp4",
          "role": "short",
          "upload": {
            "headers": {
              "Content-Type": "video/mp4"
            },
            "method": "PUT",
            "url": "https://raw-bucket.example.com/presigned/short-b"
          },
          "uploadEntryId": "cu_short_01hrp6wkcw3j04s6f5r0f5t2ne"
        }
      ]
    }
  },
  "meta": {
    "requestId": "req_creator_upload_packages_create_001",
    "page": null
  },
  "error": null
}
```

### Success Rules

- response は request と同じ cardinality を返します。
- `uploadTargets.main` は常に 1 件です。
- `uploadTargets.shorts` は request と同じ件数を返します。
- client は各 target に対して file binary をそのまま `PUT` します。
- client は `upload.headers` に含まれる header を変更せず送ります。
- presigned PUT 成功後でも、`complete` を呼ぶまでは content row 作成成功を意味しません。

## `POST /api/creator/upload-packages/complete`

### Request

```json
{
  "packageToken": "cupkg_01hrp6wjkq7mh6f3d2f6c5j8rz",
  "main": {
    "uploadEntryId": "cu_main_01hrp6wjv9h2k4g1b8s0k5e3pf",
    "priceJpy": 1800,
    "ownershipConfirmed": true,
    "consentConfirmed": true
  },
  "shorts": [
    {
      "uploadEntryId": "cu_short_01hrp6wk4m2q7cxnq0fxsy2f55",
      "caption": "quiet rooftop preview。"
    },
    {
      "uploadEntryId": "cu_short_01hrp6wkcw3j04s6f5r0f5t2ne",
      "caption": null
    }
  ]
}
```

| field | type | required | notes |
| --- | --- | --- | --- |
| `packageToken` | `string` | yes | opaque package identifier |
| `main.uploadEntryId` | `string` | yes | initiation response の main target を指す |
| `main.priceJpy` | `number` | yes | `> 0` の JPY integer |
| `main.ownershipConfirmed` | `boolean` | yes | `true` 固定。権利確認が完了していること |
| `main.consentConfirmed` | `boolean` | yes | `true` 固定。同意確認が完了していること |
| `shorts[].uploadEntryId` | `string` | yes | initiation response の short target を指す |
| `shorts[].caption` | `string \| null` | yes | blank caption は client / server で `null` に正規化 |

### Success

- status は `200 OK`
- upload object の検証に成功した場合だけ draft content を返します

```json
{
  "data": {
    "main": {
      "id": "c5bc96d4-5b38-4fcf-a723-b7e2f04fd9e9",
      "mediaAsset": {
        "id": "4f571b4b-2755-49c2-8d77-7dd205899f8d",
        "mimeType": "video/mp4",
        "processingState": "uploaded"
      },
      "state": "draft"
    },
    "shorts": [
      {
        "canonicalMainId": "c5bc96d4-5b38-4fcf-a723-b7e2f04fd9e9",
        "id": "c8d7c617-61b3-48c5-8b43-c1f3807fa19e",
        "mediaAsset": {
          "id": "a890caef-5b4b-4ddd-b08d-866ddf5286d0",
          "mimeType": "video/mp4",
          "processingState": "uploaded"
        },
        "state": "draft"
      },
      {
        "canonicalMainId": "c5bc96d4-5b38-4fcf-a723-b7e2f04fd9e9",
        "id": "8bbf6f50-31ee-4478-9f68-b56046fae1aa",
        "mediaAsset": {
          "id": "3306acb1-b4db-443e-8d48-53f630cfee06",
          "mimeType": "video/mp4",
          "processingState": "uploaded"
        },
        "state": "draft"
      }
    ]
  },
  "meta": {
    "requestId": "req_creator_upload_packages_complete_001",
    "page": null
  },
  "error": null
}
```

### Completion Rules

- completion 成功時に、draft `main` 1 件と draft `shorts` 1 件以上を同一 transaction で作成します。
- `main` / `shorts` / `media_assets` の `creator_user_id` は request user で一致していなければなりません。
- `media_assets` は各 upload entry ごとに 1 件作成し、`processing_state = uploaded`、`storage_provider = s3`、`external_upload_ref = uploadEntryId` を持ちます。
- draft `main` は `state = draft` で作成し、`price_minor = priceJpy`、`currency_code = JPY`、`ownership_confirmed = true`、`consent_confirmed = true` を保存します。
- draft `shorts` も `state = draft` で作成し、`caption` は `null` または trim 後 non-empty string を保存します。
- draft `shorts` はすべて、同じ completion で作った draft `main` の `canonical_main_id` に紐づきます。
- completion 成功時は `media_asset_id` ごとに durable な processing job を 1 件作成します。upload completion 直後の response では `mediaAsset.processingState = uploaded` のままとし、worker 側の claim 以降で processing state を進めます。
- object 検証に失敗した場合は package 全体を失敗扱いにし、`main` / `shorts` / `media_assets` の新規 row は 1 件も作りません。
- completion 成功は `submission package ready` や `review submit` を意味しません。processing / linkage / review は後続 boundary の責務です。
- `review submit` の canonical contract は [submission-package-review-contract.md](submission-package-review-contract.md) を参照します。

## HTTP States

| endpoint | success | invalid_request | validation_error | auth_required | capability_required | upload_expired | upload_failure | storage_failure | internal_error |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /api/creator/upload-packages` | `200` | `400` | `400` | `401` | `403` | no | no | `503` | `500` |
| `POST /api/creator/upload-packages/complete` | `200` | `400` | `400` | `401` | `403` | `409` | `422` | `503` | `500` |

## Error Contract

| status | code | meaning |
| --- | --- | --- |
| `400` | `invalid_request` | malformed JSON / extra payload |
| `400` | `validation_error` | `main` 欠落、`shorts` 0 件、empty file name、non-video MIME、non-positive size など |
| `401` | `auth_required` | creator upload requires authentication |
| `403` | `capability_required` | approved creator capability がない |
| `409` | `upload_expired` | `packageToken` または target の有効期限が切れた |
| `422` | `upload_failure` | expected object が未 upload、欠落、role 不一致、package 不整合 |
| `503` | `storage_failure` | presigned PUT 発行や object 検証で storage dependency が失敗した |
| `500` | `internal_error` | unexpected server failure |

```json
{
  "data": null,
  "meta": {
    "requestId": "req_creator_upload_packages_error_001",
    "page": null
  },
  "error": {
    "code": "capability_required",
    "message": "approved creator capability is required"
  }
}
```

## Boundary Guardrails

- `link-short` は今回の contract に含めません。既存 `main` に `short` を追加する flow は別 issue で扱います。
- response に public playback URL、review state、publishability は含めません。
- upload target は raw bucket 受け口だけを表し、delivery-ready asset を返しません。
- fan/public surface の DTO や fixture に creator-private upload state を混ぜません。
- `main` / `shorts` が揃わない partial package 保存は v1 では行いません。

## Fixture Reference

- representative fixture は [creator-upload.json](fixtures/creator-upload.json) を参照します。
