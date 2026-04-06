# Fan MVP Fixtures And Integration Guide

## 位置づけ

- この文書は `SHO-20 fan MVP surface 契約の representative fixture と接続前提を整理する` の成果物です。
- `SHO-16` から `SHO-19` までの契約文書と fixture を、backend 実装順と frontend 接続順の両方で迷わず参照できるようにします。

## Deliverables

| issue | contract | fixture |
| --- | --- | --- |
| `SHO-53` | `docs/contracts/fan-auth-api-contract.md` | `docs/contracts/fixtures/fan-auth.json` |
| `SHO-39` | `docs/contracts/viewer-bootstrap-api-contract.md` | `docs/contracts/fixtures/viewer-bootstrap.json` |
| `SHO-16` | `docs/contracts/fan-mvp-common-transport-contract.md` | `docs/contracts/fixtures/fan-mvp-common.json` |
| `SHO-17` | `docs/contracts/fan-public-surface-api-contract.md` | `docs/contracts/fixtures/fan-public-surfaces.json` |
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
3. `SHO-18`
   - `GET /api/fan/shorts/{shortId}/unlock`
   - `GET /api/fan/mains/{mainId}/playback`
4. `SHO-19`
   - `GET /api/fan/profile`
   - `GET /api/fan/profile/following`
   - `GET /api/fan/profile/pinned-shorts`
   - `GET /api/fan/profile/library`
   - `GET /api/fan/profile/settings`

## App Bootstrap Connection Order

1. `SHO-53`
   - `POST /api/fan/auth/sign-in/challenges`
   - `POST /api/fan/auth/sign-in/session`
   - `POST /api/fan/auth/sign-up/challenges`
   - `POST /api/fan/auth/sign-up/session`
   - `DELETE /api/fan/auth/session`
   - auth mutation 成功後は cookie のみを更新し、viewer state はまだ返さない
2. `SHO-39`
   - `GET /api/viewer/bootstrap`
   - current viewer の `id / activeMode / canAccessCreatorMode`
   - unauthenticated bootstrap 時は `currentViewer = null`
3. `SHO-17` 以降の public surface
   - viewer 自身の state ではなく resource relation state だけを参照する
4. `SHO-19`
   - private hub は bootstrap 済みの current viewer を前提に接続する

## Frontend Connection Order

| downstream issue | UI area | primary contract | fixture scenarios |
| --- | --- | --- | --- |
| `SHO-39` | `app shell bootstrap` | `viewer-bootstrap-api-contract.md` | `authenticatedFan`, `authenticatedCreator`, `unauthenticated` |
| `SHO-5` | `feed / short detail` | `fan-public-surface-api-contract.md` | `recommended_public`, `recommended_purchased`, `short_detail_public`, `short_detail_purchased`, `short_detail_owner`, `short_detail_not_found` |
| `SHO-6` | `creator search / creator profile` | `fan-public-surface-api-contract.md` | `search_recent`, `search_filtered`, `creator_profile_normal`, `creator_profile_empty`, `creator_profile_not_found` |
| `SHO-8` | `mini paywall / main player` | `fan-unlock-main-api-contract.md` | `setup_required`, `unlock_available`, `purchased`, `owner`, `locked`, `not_found`, `playback_purchased`, `playback_owner` |
| `SHO-7` | `fan profile private hub` | `fan-profile-api-contract.md` | `overview_populated`, `overview_empty`, `following_populated`, `pinned_populated`, `library_populated`, `settings_default` |

- `SHO-7` の初回表示では `GET /api/fan/profile` で counts を取得し、default tab の `GET /api/fan/profile/pinned-shorts` を別で呼びます。
- `GET /api/fan/profile/library` は tab を開いた時点で初回 fetch し、以後は cursor を使って scroll 追加取得します。
- auth viewer の self / session / active mode は app bootstrap 時の global state を正とし、surface payload からは参照しません。

## Scenario Rules

- `empty` は `200` 成功系で表現します。
- `not_found` は `404 + error.code = not_found` で表現します。
- `locked` は `403 + error.code = main_locked` で表現します。
- `owner` は purchase と混ぜず、`MainAccessState.status = owner` で表現します。
- 金額と時間は raw 値で返し、frontend が `¥` 表示と分秒 formatting を担います。

## Fixture Usage Rules

- fixture は UI 専用 mock ではなく transport contract の canonical example です。
- 新しい business rule を fixture 側に追加しません。未知の仕様を補うのではなく、既存 contract の例示に留めます。
- 1 つの scenario が複数 endpoint にまたがる場合でも、各 endpoint fixture を正とします。frontend 側で scenario 名だけに依存しません。
