# Fan Short Comment API Contract

## 位置づけ

- この文書は public `short` に紐づく fan comment の read / create transport contract を固定します。
- comment は `short` 専用であり、`main` playback surface には提供しません。
- comment list は動画取得 API と分離し、comment button 押下後に cursor pagination で取得します。

## Canonical Sources

- [fan-mvp-common-transport-contract.md](fan-mvp-common-transport-contract.md)
- [fan-public-surface-api-contract.md](fan-public-surface-api-contract.md)
- [mvp-core-domain-contract.md](mvp-core-domain-contract.md)

## Endpoint Summary

| method | path | auth | notes |
| --- | --- | --- | --- |
| `GET` | `/api/fan/shorts/{shortId}/comments` | optional | public short comments |
| `POST` | `/api/fan/shorts/{shortId}/comments` | required | create short comment |

## Request Contract

### `GET /api/fan/shorts/{shortId}/comments`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Query

| field | type | required | notes |
| --- | --- | --- | --- |
| `cursor` | `string` | no | opaque newest-first keyset continuation scoped to `{shortId}` |

#### Response

- `data.items`: `ShortComment[]`
- `meta.page`: `CursorPageInfo`
- order は `createdAt DESC, id DESC`

### `POST /api/fan/shorts/{shortId}/comments`

#### Path

| field | type | required |
| --- | --- | --- |
| `shortId` | `string` | yes |

#### Body

| field | type | required | notes |
| --- | --- | --- | --- |
| `body` | `string` | yes | trim 後 `1..500` 文字 |

#### Response

- `data.comment`: created `ShortComment`
- `meta.page = null`

## DTOs

### `ShortComment`

| field | type | notes |
| --- | --- | --- |
| `id` | `string` | comment identifier |
| `shortId` | `string` | comment target short |
| `body` | `string` | comment body |
| `author` | `ShortCommentAuthor` | display-only author summary |
| `createdAt` | `string` | RFC3339 timestamp |

### `ShortCommentAuthor`

| field | type | notes |
| --- | --- | --- |
| `displayName` | `string` | current display name |
| `handle` | `string` | current handle with `@` prefix |
| `avatar` | `ImageAsset \| null` | current avatar |

## HTTP States

| case | status |
| --- | --- |
| list success | `200` |
| list empty | `200` |
| create success | `201` |
| create unauthenticated | `401` + `auth_required` |
| invalid cursor / malformed JSON | `400` + `invalid_request` |
| invalid body | `400` + `validation_error` |
| non-public or missing short | `404` + `not_found` |
| unexpected server error | `500` + `internal_error` |

## Guardrails

- Comment payloads are not embedded in `GET /api/fan/feed` or `GET /api/fan/shorts/{shortId}`.
- Comment count is not embedded in feed, short detail, or action rail payloads in v1.
- Replies, likes, deletion, reports, and moderation queue behavior are out of scope.
- `main` comments are out of scope and must not be introduced by this contract.

## Fixture Reference

- representative fixture は [fan-short-comments.json](fixtures/fan-short-comments.json) を参照します。

## Schema Generation Note

- `backend/db/schema.generated.yaml` は migration chain 全体から生成する snapshot です。
- short comment migration `000020_add_short_comments` の snapshot 更新時に、既存 migration `000018_add_submission_review_decisions` / `000019_add_submission_review_notes` が前回 snapshot から漏れていた場合、それらも generator catch-up として同じ生成差分に含まれます。
- 上記は schema snapshot の整合化であり、この contract が submission review の挙動を変更するものではありません。
