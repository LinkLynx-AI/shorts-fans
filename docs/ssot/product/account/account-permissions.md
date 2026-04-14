# short-fans Product SSOT - Account Permissions

## 位置づけ

- `creator account` と `fan account` の権限差分を整理する
- これは `onboarding`、`navigation`、`投稿フロー`、`分析可視化`、`アクセス制御` に直結する

## 安定した事実

- アカウント種別は `creator account` と `fan account` の 2 種のみを想定している
- raw メモ上では、`どちらのモードでログインするかで体験内容が変わる` 前提がある
- `short` は public surface であり、`main` は課金または保有状態によってアクセスが決まる
- `creator profile` は public に見える面だが、投稿管理や分析画面は creator 側の private surface である
- `MVP` の投稿単位は `1 canonical main : 複数 short` を基本に置く
- `MVP` の identity model は `1 user identity + active mode switch` を第一候補に置く

## 基本方針

- `account type` は単なるラベルではなく、`mode ごとの権限束` として扱う
- `fan` と `creator` で同じ画面を見せ分けるのではなく、`home`、`primary CTA`、`管理対象` を明確に分ける
- ただし `short`、`creator profile` のような public surface は、内部権限とは切り分けて考える
- `main` の視聴権限は `account type` だけではなく、`purchase / ownership state` で制御する

## MVP の権限モデル

### creator account

- 自分の `main` を作成、アップロード、更新できる
- 自分の `short` を作成、アップロード、更新できる
- 自分の `canonical main` に対して複数 short を紐付けできる
- 自分の `creator profile` を編集できる
- 自分の投稿ステータス、審査状態、最低限の analytics を見られる
- 自分の `main` は課金なしで preview / QA できる
- 他 creator の内部管理情報や analytics は見られない

### fan account

- `feed`、`short`、`creator profile` を閲覧できる
- creator をフォローできる
- `main` を unlock / 購入できる
- 購入済みの `main` を視聴できる
- 自分の follow や purchase に紐づく consumer-side 状態と private hub を持てる
- 投稿、紐付け、creator analytics には触れられない

## 画面と権限の対応

### public surface

- `short feed`
- `short player`
- `creator profile`

### creator private surface

- `upload / import`
- `short-main linkage editor`
- `creator dashboard`
- `review / moderation status`

### pre-approval creator surface

- `Become Creator`
- shared viewer profile basics の preview
- creator 固有の `bio` draft
- `static mock` の creator preview
- onboarding checklist
- creator approval status
- rejection reason / `re-submit` eligibility
- policy / requirement guidance

### fan private surface

- `purchase state`
- `unlocked main access`
- `following state`

## 現時点の推奨

- `MVP` は `1 user` の中で `fan mode` と `creator mode` を分けて考える方がよい
- 理由
  - `fan` の主導線は `feed -> short -> unlock -> main`
  - `creator` の主導線は `upload -> linkage -> publish -> monitor`
  - 両者は目的が違いすぎるため、同じ home に混ぜるとどちらも弱くなる
- したがって、`public surface は共通`、`private workspace は role ごとに分離` が基本になる

## 継続仮説

- `creator account` は viewing 権限まで完全に切り離すより、同じ identity の中で public surface の確認や market 観察ができた方が自然
- 一方で `creator analytics`、`投稿管理`、`審査状態` のような supply-side 権限は、fan 側と混ぜない方がよい
- `main` の access control は、`creator ownership` と `fan purchase` を分けて持つ方がデータ的にも運用的にもきれい
- 将来 `一人のユーザーが creator と fan の両方を持つ` 可能性はあるが、`MVP` では `どの mode で入っているか` を明示した方が UX がぶれない

## 未解決論点

- admin / moderation operator を、別 account type ではなく内部権限としてどう持つか

## 参照

- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-onboarding-surface.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/consumer-state-and-profile.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/moderation/moderation-and-review.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
