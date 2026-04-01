# short-fans Product SSOT - Consumer State And Profile

## 位置づけ

- `creator が同じ identity で purchase できるか`
- `fan profile` を何として持つか

- この 2 つをまとめて整理する
- これは `wallet`、`purchase state`、`library`、`mode switching`、`consumer retention` に直結する

## 現時点の推奨

- `consumer-side purchase` は `same user identity` に持たせる
- ただし `purchase action` は `fan mode` に寄せる
- `fan profile` は `public profile` ではなく、`private consumer hub` として持つ
- `pin` は fan-side の private state として同じ identity に持たせる

## 推奨理由

### 1. wallet と unlock state を分断しないため

- `purchase / unlock history`
- `payment method`
- `following`
- `pin state`
- 将来の `library / retention state`

- これらを creator と fan で別 identity に分ける理由は薄い

### 2. mode の役割を濁らせないため

- `creator mode` は `upload -> linkage -> publish -> monitor` の workspace である
- ここに通常の paywall 購入行動まで混ぜると、supply-side と consumer-side の導線が濁る
- よって `purchase 権限は同じ identity に持たせるが、purchase 行為は fan mode でさせる` のがよい

### 3. fan profile を public social surface にしないため

- このプロダクトの public surface は `short` と `creator profile`
- `fan` 側まで public profile を強く出すと、`short -> main` のコア導線が薄まる
- 初期は `consumer state の整理面` としての方が価値が高い

## MVP の決め方

### purchase capability

- creator capability を持つ user でも、同じ identity の中で `purchase / unlock` を持てる
- ただし purchase CTA の primary surface は `fan mode`
- creator が public surface から他 creator の main に進むときは、必要なら `fan mode` へ寄せる

### own content

- creator 自身の `main` は `purchase` ではなく `ownership / preview` で扱う
- つまり自分の作品に対しては paywall を踏ませず、`preview / manage` 導線を出す

### fan profile

- `fan profile` は `private` を基本に置く
- MVP の最低要件
  - `continue watching`
  - `following`
  - `pinned shorts`
  - `unlocked / purchased mains`
  - 必要最小限の `consumer settings`

- ここでは `public-facing identity page` は作らない

## 画面の考え方

### fan mode

- `feed`
- `creator profile`
- `paywall / unlock`
- `main player`
- `fan profile` as private hub

### creator mode

- `creator dashboard`
- `upload / import`
- `short-main linkage`
- `review / moderation status`

- consumer-side purchase は内部的には同じ identity に紐づいていても、primary navigation には出しすぎない

## 継続仮説

- creator でも他 creator の作品を見る / 課金する需要はある
- ただしそれを `creator mode` の主導線に置く必要はない
- `fan profile` は、SNS 的な自己表現面より `library + following + settings` の方が MVP 価値が高い
- `pin` は public social signal ではなく、private な再訪導線として持つ方がよい
- public fan profile を急いで作るより、`unlocked main の再訪` と `follow からの再回遊` を強くした方がよい

## 未解決論点

- creator が `creator mode` のまま purchase まで完結できるようにするか、それとも purchase 時に `fan mode` へ明示的に寄せるか
- `fan profile` に `watch history` を full UI として出すか
- follow、pin、purchase、library を 1 画面にまとめるか、分けるか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-profile-and-engagement.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
