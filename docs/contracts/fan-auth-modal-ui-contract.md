# Fan Auth Modal UI Contract

## 位置づけ

- この文書は `SHO-166 Cognito 前提で fan auth 契約と境界を更新する` の成果物です。
- Cognito ベース fan auth を受ける custom modal UI の state、entry、success behavior、recovery responsibility を固定します。
- primary entry は shared modal であり、full-page redirect は正規主導線として扱いません。

## Goals

- existing fan auth modal の雰囲気を維持したまま、`sign in / sign up / sign up confirm / password reset / re-auth` を modal で扱えるようにする。
- `auth_required` と `fresh_auth_required` の両方を shared modal で回収できるようにする。
- auth 成功後の viewer/bootstrap 更新と current context への復帰を現行仮 auth と同じ挙動で固定する。
- public short 視聴体験を壊さず、必要な場面だけ auth を要求する。

## Non-goals

- pixel-perfect design spec
- full-page `/login` fallback page の詳細 UI
- creator login UI
- social login UI
- auth endpoint 自体の transport contract

## Canonical Sources

- `docs/contracts/fan-auth-api-contract.md`
- `docs/contracts/viewer-bootstrap-api-contract.md`
- `docs/contracts/fan-short-pin-api-contract.md`
- `docs/contracts/fan-creator-follow-api-contract.md`
- `docs/ssot/product/fan/fan-journey.md`
- `docs/ssot/product/ui/fan-surfaces.md`

## Entry Rules

- protected fan action が `auth_required` を返したときは shared fan auth modal を開きます。
- route shell が unauthenticated access を検知したときも shared fan auth modal を primary にします。
- explicit login CTA を後続実装で残す場合も、別 UX を作らず同じ modal contract を再利用します。
- full-page redirect は secondary fallback であって primary flow ではありません。

## Shared Modal Rules

- visual tone は既存 `FanAuthDialog` / `FanAuthEntryPanel` を踏襲します。
- overlay、panel、dismiss button、fan access framing は既存 modal と同じ空気感を保ちます。
- submit 中は close button、escape、outside click による dismiss を禁止します。
- mode 切替時は、失って困る field をなるべく保持します。最低でも email は preserve 対象です。
- sign-up confirmation recovery に必要な password は modal session 内の ephemeral state としてだけ保持できます。永続化や analytics 送信はしません。
- error copy は generic に保ち、account existence を露出しません。
- frontend は auth mutation success を current viewer 成功と同一視せず、`GET /api/viewer/bootstrap` で再確定します。

## Transport To UI Mapping

| transport signal | modal mode |
| --- | --- |
| `data.nextStep = "confirm_sign_up"` | `confirm-sign-up` |
| `data.nextStep = "confirm_password_reset"` | `confirm-password-reset` |
| `error.code = "confirmation_required"` | `confirm-sign-up` |

## Modal Modes

### `sign-in`

- fields:
  - `email`
  - `password`
- primary action:
  - `POST /api/fan/auth/sign-in`
- secondary actions:
  - `sign-up` への切替
  - `password-reset-request` への切替
- success:
  - modal を閉じる前に bootstrap を再読します。
  - route shell か dialog opener の current context を維持します。

### `sign-up`

- fields:
  - `email`
  - `password`
- primary action:
  - `POST /api/fan/auth/sign-up`
- secondary actions:
  - `sign-in` への切替
- success:
  - modal を `confirm-sign-up` mode に遷移させます。
  - email は preserve します。
  - auth はまだ完了扱いにしません。

### `confirm-sign-up`

- fields:
  - `email`
  - `confirmationCode`
- optional local-only state:
  - `pendingPassword`
- primary action:
  - `POST /api/fan/auth/sign-up/confirm`
- secondary actions:
  - `pendingPassword` を保持している場合だけ `POST /api/fan/auth/sign-up` の再実行による resend
  - `pendingPassword` が無い場合は `sign-up` mode に戻して password を再入力させてから resend
  - `sign-in` への切替
- success:
  - bootstrap を再読し、current viewer / session state を更新して modal を閉じます。
  - current route や opener context は維持します。

### `password-reset-request`

- fields:
  - `email`
- primary action:
  - `POST /api/fan/auth/password-reset`
- secondary actions:
  - `sign-in` へ戻る
- success:
  - modal を `confirm-password-reset` mode に遷移させます。
  - email は preserve します。
  - app session はまだ発行済み扱いにしません。

### `confirm-password-reset`

- fields:
  - `email`
  - `confirmationCode`
  - `newPassword`
- primary action:
  - `POST /api/fan/auth/password-reset/confirm`
- secondary actions:
  - `POST /api/fan/auth/password-reset` の再実行による resend
  - `sign-in` へ戻る
- success:
  - modal を `sign-in` mode に戻します。
  - email は preserve します。
  - password reset 成功だけでは modal を authenticated success で閉じません。

### `re-auth`

- fields:
  - `password`
- primary action:
  - `POST /api/fan/auth/re-auth`
- secondary actions:
  - caller が許す場合だけ cancel
- success:
  - modal を閉じて blocked action 側へ制御を返します。
  - `activeMode`、route、tab、scroll context を変えません。

## Success Behavior

- auth success 後の挙動は現行仮 auth と同じにします。
- `follow` / `pin` のような mutation は auth success 後に自動 replay しません。
- route-shell 起点の auth_required recovery は current route の bootstrap / data refresh を行い、同じ画面文脈に留めます。
- dialog opener が `afterAuthenticatedHref` のような復帰先を持つ場合でも、目的は文脈復帰であって別 auth UX への遷移ではありません。

## Error Handling Rules

- `invalid_credentials` は current mode に留まり、inline error で扱います。
- `confirmation_required` は `confirm-sign-up` へ進める recoverable state として扱います。
  - `sign-in` 送信で返った場合は `email` を preserve し、送信済み password を `pendingPassword` として modal session 内にだけ保持できます。
- `invalid_confirmation_code` と `confirmation_code_expired` は confirm mode に留まり、resend を許可します。
- `password_policy_violation` は current mode に留まり、password field を修正できるようにします。
- `rate_limited` は modal を閉じずに retry guidance を表示します。
- `auth_required` と `fresh_auth_required` は modal open trigger の責務であり、surface payload 自体に viewer self state を重ねません。

## Boundary Guardrails

- modal は Cognito Hosted UI redirect を内部実装としても primary UX にしません。
- modal success body をもって authenticated viewer を確定しません。
- modal は current route を login 専用 page に切り替えることを前提にしません。
- creator login、creator mode switch、payment setup copy はこの文書に含めません。
