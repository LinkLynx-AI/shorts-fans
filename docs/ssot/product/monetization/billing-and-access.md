# short-fans Product SSOT - Billing And Access

## 位置づけ

- `short -> main` 体験に対して、どの課金単位とアクセス制御が自然かを整理する
- 収益の最終最適化ではなく、`MVP として何が一番噛み合うか` を決めるための土台にする

## 安定した事実

- `short` は公開で見られる面であり、課金しない限り explicit な部分は見えない前提である
- `main` は課金後に視聴する本編である
- `main` は `short` の続きとして始まることが価値の中心である
- `fan` の主導線は `feed -> short -> paywall / unlock -> main` を想定している
- `main` は creator profile から直接選ぶのではなく、short から入る構造を基本にしている
- `MVP` では `複数 short` が `1 canonical main` に流れる構造を許可する
- paywall の UI 詳細は `product/ui/fan-surfaces.md` を canonical にする
- 初回購入は `mini paywall` を通し、支払い設定と年齢 / 利用同意を完了させる
- 支払い設定と同意が済んだ再購入は、`[ Unlock ¥X | n分 ]` タップから full paywall を開かずに unlock して `main` に入る

## 課金モデルの候補

### A. subscription centric

- creator 単位で購読し、その creator の main をまとめて見られる
- 強み
  - 継続売上を作りやすい
  - 課金後の friction が低い
- 弱み
  - `この 1 本の続きを見たい` という短い熱量に対して重い
  - 初回課金の意思決定が大きい
  - MVP では `main 単位の転換検証` がぼやけやすい

### B. pay-per-unlock centric

- short が流入口になり、その先の canonical main を個別 unlock する
- 強み
  - `続きを見たい` という欲求に最短で反応できる
  - `short -> main` 導線と最も自然に噛み合う
  - MVP で `どの short がどの程度 unlock を生むか` を直接見やすい
- 弱み
  - 継続課金の土台は subscription より弱い
  - heavy user の課金体験は煩雑になりやすい

### C. hybrid

- 基本は subscription を持ちつつ、特定 main や上位コンテンツは追加 unlock にする
- 強み
  - 収益の取り方が柔軟
  - OnlyFans 型の複合課金に近い
- 弱み
  - MVP には複雑すぎる
  - fan が何を払うと何が見られるかを誤解しやすい

## 現時点の推奨

- `MVP` は `pay-per-unlock centric` を第一候補に置く
- 理由
  - `short` は `main` の公開冒頭なので、fan の動機は `creator を月額で支援したい` より `この続きだけ今見たい` に近い
  - product の核は `short -> main` の転換検証であり、`main 単位` で課金結果を見た方が学習が早い
  - paywall の意味が明確で、`無料部分` と `有料部分` の境界を作りやすい
  - 同じ main に複数 short がぶら下がっても、unlock object が `main` ならロジックは崩れない

## MVP のアクセス制御

### public

- short feed
- creator profile 上の short 一覧
- short player

### locked

- main player
- main の explicit 続き

### purchase state

- `not purchased`
  - short のみ視聴可能
  - main は paywall 表示
- `purchased`
  - 対応する main を視聴可能

## MVP の paywall 原則

- paywall は `この short の続き` を最短で買うための面として扱う
- UI 詳細は `product/ui/fan-surfaces.md` を canonical にする
- 初回だけ、支払い方法選択と年齢 / 利用同意を追加する
- 再購入では full paywall を開かず、`[ Unlock ¥X | n分 ]` からそのまま unlock して `main` に進める
- 購入後は確認面を挟まず、そのまま `main` 再生へ進める

## 継続仮説

- MVP では `creator subscription` より `main unlock` の方がプロダクト理解と一致する
- `pay-per-unlock` で基礎データを取り、その後に `bundle / subscription / membership` を足す方が自然
- `unlock` は `canonical main 単位` の方が最初はわかりやすい
- 課金面では、価格の見せ方より `続きがすぐ始まること` の方が転換に効く可能性が高い
- 初回課金前に見せる情報は増やすほど良いのではなく、`続きであることがわかる最小構成` の方が転換に合う可能性が高い
- 初回購入と再購入で friction を分ける方が、安全性と転換の両立を取りやすい

## 後で広げる候補

- creator subscription
- main bundle
- 期間限定 unlock
- follow 後の会員導線
- subscription + premium unlock の hybrid

## 未解決論点

- MVP の課金単位を `main 単位` に固定するか
- creator ごとの price setting を許すか
- 初回購入後のレコメンドを `同 creator` に寄せるか、`類似 short` に寄せるか
- 将来 subscription を入れるとき、unlock 購入との整合性をどう取るか
- 複数 short が同じ main に流れたとき、課金済み状態を feed 上でどう見せるか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/scope/mvp-boundaries.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/fan-surfaces.md`
