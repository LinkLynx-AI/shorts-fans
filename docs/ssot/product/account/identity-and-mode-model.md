# short-fans Product SSOT - Identity And Mode Model

## 位置づけ

- `1人 = 1 login` をどう持つか、`creator / fan` をどう切り替えるかを整理する
- これは `onboarding`、`navigation`、`purchase state`、`creator化導線`、`support / moderation` の前提になる

## 現時点の推奨

- `MVP` は `1 user identity + optional creator capability + active mode switch` を基本に置く
- つまり、`fan account` と `creator account` を別ログインとして重複させるのではなく、`1つの user` が `fan mode` と `creator mode` を持ちうる構造にする

## 推奨理由

### 1. public surface が共通だから

- `short feed`、`short player`、`creator profile` は public surface であり、fan と creator で完全に別 identity にする理由が弱い
- 特に creator も、自分の short の見え方確認や market 観察のために public surface を見るはずである

### 2. consumer-side 状態を分断しないため

- `follow`
- `purchase / unlock history`
- `payment method`
- `safety / support history`

- これらを `fan login` と `creator login` に分けると、同一人物の行動が二重化して不自然になる

### 3. creator 化の導線が軽くなるため

- 初期は fan として入り、後から creator onboarding に進む流れを作りやすい
- つまり `fan -> creator` の転換が `新規登録し直し` ではなく `capability unlock` になる

### 4. mode ごとの home を分けやすいため

- `fan mode` の home は `feed`
- `creator mode` の home は `dashboard`

- このように `同じ identity` のまま `目的別 workspace` を切り替える方が UX が自然である

## MVP のモデル

### root identity

- `user` は 1 つの login identity を持つ
- 認証、年齢確認、決済、support 履歴の基点はこの `user` に置く

### fan state

- `follow`
- `purchase / unlock`
- `viewing history`
- `fan mode` の app state

### creator capability

- creator onboarding を通過した user にだけ付与される
- `creator profile`
- `main / short` の ownership
- `upload / linkage / analytics`
- `creator mode` の workspace

### active mode

- session ごとに `fan mode` または `creator mode` のどちらかを前面に出す
- mode が変わっても login 自体は変わらない

## UX の決め方

### sign up / first entry

- 初回は `fan mode` をデフォルトに置く
- 理由
  - public acquisition surface が `short feed` だから
  - 新規 user の大半は最初から creator 作業を始めないから

### become creator

1. user が fan として入る
2. `become creator` から onboarding に進む
3. 審査や必要情報の通過後、`creator capability` を付与する
4. mode switcher に `creator mode` が出る

### mode home

- `fan mode`
  - home は `feed`
- `creator mode`
  - home は `dashboard`
  - 投稿、紐付け、審査状態、analytics へ最短で入れるようにする

### mode switcher placement

- `MVP` では `mode switcher` の primary placement は `profile / account menu` に置く
- global nav の常設タブには置かない
- 理由
  - 初期 user の大半は `fan mode` 中心で使う
  - global nav に creator switch を常設すると、fan-only user にとってノイズになる
  - mode 切替は高頻度視聴操作より `account context の切替` に近い
- creator capability を持つ user にだけ、profile menu から `Switch to Creator` / `Switch to Fan` を出す

## account-permissions との関係

- `account-permissions` は `fan mode / creator mode` の権限差分を定義する
- このファイルは、その前提になる `identity` と `mode switching` の持ち方を定義する

## 継続仮説

- 表向きは `creator account / fan account` と呼んでも、内部モデルは `1 user + role/capability` の方がよい
- creator も fan 的な閲覧や competitor check をするため、完全分離ログインより mode switch の方が実態に合う
- `fan mode default` にしておく方が acquisition と onboarding の friction が低い
- creator capability を `approval-based unlock` にすると、safety と growth の両方を扱いやすい
- `mode switcher` は global nav より account menu の方が、public feed の没入を崩しにくい

## 未解決論点

- `approved 前` の creator status 変化を、account menu と onboarding surface のどちらで主に通知するか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/consumer-state-and-profile.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-onboarding-surface.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
