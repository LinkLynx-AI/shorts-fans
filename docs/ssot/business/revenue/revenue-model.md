# short-fans Business SSOT - Revenue Model

## 位置づけ

- `short-fans` の収益モデル、ベンチマーク、見るべき指標をまとめる
- 外部調査の生メモではなく、継続して使う前提だけを置く

## 安定した事実

- `OnlyFans` は `2023年度` と `2024年度` の公開会計で、おおむね `80:20` の take rate を維持している
- `2023年度分` は `2024-09-05` に提出され、gross payments は約 `$6.6B`、net revenue は約 `$1.3B`、pre-tax profit は約 `$658M`
- `2024年度分` は `2025-08-27` に提出され、gross payments は約 `$7.2B`、net revenue は約 `$1.41B`、pre-tax profit は約 `$683.6M`
- `OnlyFans` は `2023年時点` で `non-subscription revenue` が最大と報じられている
- 成人向け領域では、収益性だけでなく `決済 / KYC / コンプライアンス` が事業制約として大きい
- `Stripe` は `adult content / services` を unsupported としており、mainstream PSP を前提にできない
- `Fanvue` は `merchant of record` として税・決済の一部を吸収している
- `myfans` と `Fantia` は card 以外に `Paidy / atone / BitCash / コンビニ / 銀行振込` などのローカル決済を持つ
- `英語圏 adult-first launch` の payment 候補は、`CCBill / Segpay / Verotel` のような adult-capable processor を前提に見るべき
- `rails確実性優先` の場合、`payment`、`creator KYC`、`fan age assurance` を分けて考える方が実務に合う

## 継続仮説

- `short-fans` の収益本体は、`単純な月額課金` より `視聴導線の先で発生する追加課金` に寄る可能性が高い
- `short -> main` の導線が強ければ、`subscription` の有無にかかわらず `ARPPU` と `repeat purchase` を伸ばせる
- `short-fans` の business 設計では、`動画 feed の視聴数` より `有料転換率` と `課金継続率` の方が重要指標になりやすい
- `OnlyFans` 型の高収益性を再現するには、`課金導線` と `creator retention` を同時に強くする必要がある
- 収益モデルの検討では、`price table` より先に `どの payment rail が通るか` を決める必要がある
- `英語圏先行` では、adult-capable MOR/PSP の確保が go-to-market の前提になる
- `CCBill + Veriff + Yoti` のように `payment / creator KYC / fan age assurance` を分けた stack の方が初期事故は減りやすい

## 主要指標

- `gross payment volume`
- `platform take rate`
- `paid conversion rate`
- `ARPPU`
- `repeat purchase rate`
- `creator GMV per active creator`
- `active creators`
- `active paying fans`

## 未解決論点

- `subscription only`、`pay-per-unlock only`、`hybrid` のどれを基本にするか
- 国別で使える決済手段と、成人向けで通る payment rail をどう確保するか
- `merchant of record` を使うか、直接 PSP 群を束ねるか
- `CCBill / Segpay / Verotel` のどれを first-pass 候補に置くか
- `creator KYC` を `Veriff / Sumsub` のどちらで見るか
- `fan age assurance` を `Yoti` 第一候補でよいか
- creator への分配率をどの水準に置くか
- `tips / PPV / DM upsell / custom content` のうち、どこまでプロダクト標準にするか
- 国別価格設計と通貨戦略をどうするか

## 参照

- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-onlyfans-revenue.md`
- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-onlyfans-revenue-explainer.md`
- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-business-landscape.md`
- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-english-market-launch-stack.md`
