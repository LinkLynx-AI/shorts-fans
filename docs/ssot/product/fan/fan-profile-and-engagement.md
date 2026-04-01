# short-fans Product SSOT - Fan Profile And Engagement

## 位置づけ

- `fan profile` の具体UIを定義する
- `pin / like / comment` を MVP でどこまで入れるかを切る
- ここでは `consumer retention` を増やしつつ、`short -> main` のコア導線を弱めないことを優先する

## 現時点の推奨

- `fan profile` は `private consumer hub` として設計する
- `continue watching` は MVP に入れる
- `continue watching` の primary placement は `fan profile` に置く
- `follow` は MVP に入れる
- `pin` は MVP に入れる
- `like` は MVP では入れない
- `comment` は MVP では入れない

## 推奨理由

### 1. `follow` は再回遊に直結するため

- fan がまた見たいのは、抽象的な short ではなく `creator` である可能性が高い
- よって `follow` は、feed の再配信と creator 深掘りの両方に効く

### 2. `continue watching` は paid 再訪の friction を下げるため

- `main` は unlock 後に途中離脱する可能性がある
- そのとき、`途中から再開できること` は再訪価値に直結する
- 一方で、これは `watch history 全面開放` まで広げなくても成立する

### 3. `pin` は `library` と別の役割を持つため

- `library` は `unlocked / purchased main` の再訪導線である
- 一方 `pin` は `public short` の再訪導線であり、`library` では代替できない
- `pin` は `続きを買う前にまた見たい short` を持ち帰る用途に合う
- `Pinned Shorts` から後で `unlock` する動きは十分ありうるため、`pin` を軽視しない方がよい

### 4. `like` は強い product signal になりにくいため

- このプロダクトで本当に強い signal は
  - `view`
  - `main click`
  - `unlock`
  - `follow`

- `like` はその補助にはなるが、MVP の核ではない

### 5. `comment` は moderation と social complexity を増やしすぎるため

- コメントを入れると
  - 通報
  - ハラスメント
  - bot / spam
  - 公開面の空気管理

- が一気に重くなる
- 初期は `short -> main` と creator supply を固める方が先である

## MVP の fan profile

### fan profile の役割

- `continue watching` による paid 再開
- `following` の管理
- `pinned shorts` の再訪
- `unlocked / purchased main` の再訪
- `consumer settings` への入口

### fan profile の最低要件

- `Continue Watching`
  - partially watched な unlocked main の再開
- `Following`
  - follow 中 creator の一覧
- `Pinned Shorts`
  - fan 自身が pin した short の一覧
- `Library`
  - unlocked / purchased main の一覧
- `Settings`
  - account / payment / safety の最低限

### fan profile に置かないもの

- public-facing profile page
- follower / following の公開ソーシャルグラフ
- public activity log
- comment history

## feed / player 側の engagement

### in-scope

- `follow`
- `pin`
- `unlock`

### out-of-scope 寄り

- `continue watching` の home surface 常設
- `like`
- `comment`
- `share` のネイティブ拡張

## 継続仮説

- `like` より `follow` の方が creator 単位の関係性を作りやすい
- `comment` を抜くことで、public surface をより clean に保ちやすい
- fan profile は `My Page` より `Library` に近い感覚で作る方が MVP に合う

## 未解決論点

- `watch history` を full UI として出すかどうか
- `pin` を `feed` 上で常設するか、overflow 的な secondary action に留めるか
- `Pinned Shorts` を pin 時系列で出すか、手動 reorder を許すか
- `Library` を creator ごとにまとめるか、unlock 時系列で出すか
- `mode switcher` を profile menu の中でどこまで目立たせるか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/consumer-state-and-profile.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/scope/mvp-boundaries.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
