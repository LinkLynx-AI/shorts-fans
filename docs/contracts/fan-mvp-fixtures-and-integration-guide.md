# Fan MVP Fixtures And Integration Guide

## 位置づけ

- この文書は `SHO-20 fan MVP surface 契約の representative fixture と接続前提を整理する` の成果物です。
- `SHO-16` から `SHO-19` と `SHO-113` の契約文書と fixture を、backend 実装順と frontend 接続順の両方で迷わず参照できるようにします。

## Deliverables

| issue | contract | fixture |
| --- | --- | --- |
| `SHO-166` | `docs/contracts/fan-auth-api-contract.md` | `docs/contracts/fixtures/fan-auth.json` |
| `SHO-169` | `docs/contracts/fan-auth-modal-ui-contract.md` | `-` |
| `SHO-39` | `docs/contracts/viewer-bootstrap-api-contract.md` | `docs/contracts/fixtures/viewer-bootstrap.json` |
| `SHO-16` | `docs/contracts/fan-mvp-common-transport-contract.md` | `docs/contracts/fixtures/fan-mvp-common.json` |
| `SHO-17` | `docs/contracts/fan-public-surface-api-contract.md` | `docs/contracts/fixtures/fan-public-surfaces.json` |
| `SHO-161` | `docs/contracts/fan-short-pin-api-contract.md` | `docs/contracts/fixtures/fan-short-pin.json` |
| `SHO-113` | `docs/contracts/fan-creator-follow-api-contract.md` | `docs/contracts/fixtures/fan-creator-follow.json` |
| `SHO-18` | `docs/contracts/fan-unlock-main-api-contract.md` | `docs/contracts/fixtures/fan-unlock-main.json` |
| `SHO-19` | `docs/contracts/fan-profile-api-contract.md` | `docs/contracts/fixtures/fan-profile.json` |

## Backend Implementation Order

1. `SHO-16`
   - response envelope
   - common DTOs
   - common state vocabulary
2. `SHO-17`
   - `GET /api/fan/feed`
   - `GET /api/fan/shorts/{shortId}`
   - `GET /api/fan/creators/search`
   - `GET /api/fan/creators/{creatorId}`
   - `GET /api/fan/creators/{creatorId}/shorts`
3. `SHO-161`
   - `PUT /api/fan/shorts/{shortId}/pin`
   - `DELETE /api/fan/shorts/{shortId}/pin`
4. `SHO-113`
   - `PUT /api/fan/creators/{creatorId}/follow`
   - `DELETE /api/fan/creators/{creatorId}/follow`
5. `SHO-18`
   - `GET /api/fan/shorts/{shortId}/unlock`
   - `POST /api/fan/mains/{mainId}/purchase`
   - `POST /api/fan/mains/{mainId}/access-entry`
   - `GET /api/fan/mains/{mainId}/playback`
6. `SHO-19`
   - `GET /api/fan/profile`
   - `GET /api/fan/profile/following`
   - `GET /api/fan/profile/pinned-shorts`
   - `GET /api/fan/profile/library`
   - `GET /api/fan/profile/settings`

## App Bootstrap Connection Order

1. `SHO-166`
   - `POST /api/fan/auth/sign-in`
   - `POST /api/fan/auth/sign-up`
   - `POST /api/fan/auth/sign-up/confirm`
   - `POST /api/fan/auth/password-reset`
   - `POST /api/fan/auth/password-reset/confirm`
   - `POST /api/fan/auth/re-auth`
   - `DELETE /api/fan/auth/session`
   - sign in / sign up confirm / re-auth / logout 成功時も viewer state 自体は返さず、client は cookie と follow-up bootstrap を正とする
   - sign up / password reset の開始 endpoint は modal state を進める accepted response を返す
2. `SHO-39`
   - `GET /api/viewer/bootstrap`
   - current viewer の `id / activeMode / canAccessCreatorMode`
   - unauthenticated bootstrap 時は `currentViewer = null`
3. `SHO-17` 以降の public surface
   - viewer 自身の state ではなく resource relation state だけを参照する
4. `SHO-161`
   - `PUT /api/fan/shorts/{shortId}/pin`
   - `DELETE /api/fan/shorts/{shortId}/pin`
   - bootstrap 済みの authenticated fan session を前提にし、success では `viewer.isPinned` だけを返す
5. `SHO-113`
   - `PUT /api/fan/creators/{creatorId}/follow`
   - `DELETE /api/fan/creators/{creatorId}/follow`
   - bootstrap 済みの authenticated fan session を前提にし、success では relation state と `fanCount` だけを返す
6. `SHO-19`
   - private hub は bootstrap 済みの current viewer を前提に接続する

## Frontend Connection Order

| downstream issue | UI area | primary contract | fixture scenarios |
| --- | --- | --- | --- |
| `SHO-169` | `shared fan auth modal` | `fan-auth-modal-ui-contract.md` + `fan-auth-api-contract.md` | `signInSuccess`, `signInInvalidCredentials`, `signInConfirmationRequired`, `signUpAccepted`, `signUpConfirmSuccess`, `passwordResetAccepted`, `passwordResetConfirmSuccess`, `reAuthSuccess`, `logoutSuccess` |
| `SHO-39` | `app shell bootstrap` | `viewer-bootstrap-api-contract.md` | `authenticatedFan`, `authenticatedCreator`, `unauthenticated` |
| `SHO-5` | `feed / short detail` | `fan-public-surface-api-contract.md` | `recommended_public`, `recommended_unlocked`, `following_ranked`, `following_empty`, `following_auth_required`, `short_detail_public`, `short_detail_unlocked`, `short_detail_owner`, `short_detail_not_found` |
| `SHO-163` | `feed pin CTA` | `fan-short-pin-api-contract.md` | `pin_success`, `pin_auth_required`, `pin_not_found`, `pin_repeat`, `unpin_success`, `unpin_auth_required`, `unpin_not_found`, `unpin_repeat` |
| `SHO-6` | `creator search / creator profile` | `fan-public-surface-api-contract.md` | `search_recent`, `search_filtered`, `creator_profile_header_normal`, `creator_profile_header_not_found`, `creator_profile_shorts_normal`, `creator_profile_shorts_empty`, `creator_profile_shorts_not_found`, `creator_profile_shorts_next_page` |
| `SHO-115` | `creator profile follow CTA` | `fan-creator-follow-api-contract.md` | `follow_success`, `follow_auth_required`, `follow_not_found`, `follow_repeat`, `unfollow_success`, `unfollow_auth_required`, `unfollow_not_found`, `unfollow_repeat` |
| `SHO-8` | `mini setup / main player` | `fan-unlock-main-api-contract.md` | `setup_required`, `unlock_available`, `purchase_pending`, `already_purchased`, `owner`, `main_not_unlockable`, `not_found`, `purchase_succeeded`, `purchase_failed_declined`, `purchase_failed_authentication`, `purchase_failed_card_brand_unsupported`, `entry_issued_after_purchase`, `entry_issued_owner`, `playback_purchased`, `playback_owner` |
| `SHO-7` | `fan profile private hub` | `fan-profile-api-contract.md` | `overview_populated`, `overview_empty`, `following_populated`, `pinned_populated`, `library_populated`, `settings_default` |

- `SHO-6` の creator profile 初回表示では `GET /api/fan/creators/{creatorId}` と `GET /api/fan/creators/{creatorId}/shorts` を並列取得します。
- `SHO-6` の short grid 追加取得では `GET /api/fan/creators/{creatorId}/shorts?cursor=...` だけを再度呼びます。
- `SHO-169` の shared fan auth modal は primary entry を modal に固定し、auth success 後の behavior は current bootstrap refresh を正とします。
- `SHO-163` の feed pin CTA は `PUT / DELETE /api/fan/shorts/{shortId}/pin` を使い、success body の `viewer.isPinned` で current surface state を更新できます。
- `SHO-115` の creator profile follow CTA は `PUT / DELETE /api/fan/creators/{creatorId}/follow` を使い、success body の `viewer.isFollowing` と `stats.fanCount` で header state を更新できます。
- `SHO-8` の `Unlock` CTA は `GET /api/fan/shorts/{shortId}/unlock` で paywall state を読み、未購入なら `POST /api/fan/mains/{mainId}/purchase`、購入済みまたは owner なら `POST /api/fan/mains/{mainId}/access-entry` を経て main route へ遷移します。
- `SHO-7` の初回表示では `GET /api/fan/profile` で counts を取得し、default tab の `GET /api/fan/profile/pinned-shorts` を別で呼びます。
- `GET /api/fan/profile/library` は tab を開いた時点で初回 fetch し、以後は cursor を使って scroll 追加取得します。
- auth viewer の self / session / active mode は app bootstrap 時の global state を正とし、surface payload からは参照しません。

## Scenario Rules

- feed fixture の `items` 順序は recommendation output の representative example であり、`publishedAt DESC` や newest-first を意味しません。
- feed `cursor` は recommendation order を継続取得する opaque token として扱い、sort key や ranking reason を表しません。
- `empty` は `200` 成功系で表現します。
- `not_found` は `404 + error.code = not_found` で表現します。
- `main_not_unlockable` と `purchase_required` は `403 + error.code = main_locked` で表現します。
- `owner` は session unlock と混ぜず、`MainAccessState.status = owner` で表現します。
- fan access は `MainAccessState.reason = purchased` で表現し、temporary playback grant と durable purchase を混ぜません。
- 金額と時間は raw 値で返し、frontend が `¥` 表示と分秒 formatting を担います。

## Fixture Usage Rules

- fixture は UI 専用 mock ではなく transport contract の canonical example です。
- 新しい business rule を fixture 側に追加しません。未知の仕様を補うのではなく、既存 contract の例示に留めます。
- 1 つの scenario が複数 endpoint にまたがる場合でも、各 endpoint fixture を正とします。frontend 側で scenario 名だけに依存しません。
- downstream は feed fixture の item order から raw score や recommendation reason を逆算しません。payload shape と contract に書かれた semantics を正とします。
