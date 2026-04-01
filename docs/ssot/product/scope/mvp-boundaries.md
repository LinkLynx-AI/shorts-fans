# short-fans Product SSOT - MVP Boundaries

## 位置づけ

- 初期プロダクトで `何をやるか / やらないか` を切る
- 後から広げる余地は残しつつ、`short -> main` 体験を成立させる最小構成を定義する

## MVP で成立させるもの

- `fan`
  - short feed を見る
  - creator をフォローする
  - short から main に進む
  - main を unlock / 課金して視聴する
- `creator`
  - short と main をアップロードする
  - short と main を紐付ける
  - profile 上で short を公開する
  - creator private surface で投稿管理と最低限の analytics を行う
- `platform`
  - short を public surface として配信する
  - main を課金制で保護する
  - 最低限の審査 / 管理オペレーションを持つ

## MVP の in-scope

- `おすすめ / フォロー中` feed
- `creator search`
- `creator profile`
- `short player`
- `main unlock flow`
- `main player`
- `fan profile` の最低限の private hub
- unlocked main 向けの `continue watching`
- `動画インポート`
- `short 切り出し`
- `short-main 紐付け`
- `conversion-first` の最低限 creator analytics
- approval 前 creator 向けの `read-only onboarding surface + static mock preview`
- fixable な creator rejection に対する最小限の `self-serve resubmit`
- self-serve resubmit の `最大2回` ガードレール
- 最低限の creator 向け管理画面
- `1 canonical main : 複数 short` の紐付け
- `creator mode / fan mode` の基本的な権限分離
- `creator / short / main / 通報後対応` の最低限の moderation operation
- `優先審査` は `3件/60日` を初期値とし、将来も `2件/30日` より先には緩めない

## MVP の out-of-scope 寄り

- `写真 / ストーリー / live / blog` のような派生フォーマット
- `アプリ内撮影`
- ネイティブの音源ライブラリ
- 高度な creator editing suite
- 複雑な social graph
- `like / comment` のような公開 engagement
- full `watch history` UI
- creator 向けの高度な分析機能
- API / MCP の外部公開
- `short caption / hashtag / full-text search`
- `カテゴリ / Explore`

## 後で判断するもの

- `subscription` を後から足すかどうか
- `DM / chat / custom request` を MVP に含めるか
- `PWA` を主導面に置いた上で `Android native` を補助導線として足すか

## 継続仮説

- MVP で証明したいのは、`short` が public acquisition surface として機能し、そのまま `main unlock` に繋がるかどうか
- 初期は `creator向け制作機能` を広げるより、`fan側導線` を磨く方が重要
- プロダクトを広げすぎるより、`short -> main` の 1 本線を強く作る方が勝ちやすい
- 初期課金は `pay-per-unlock` を第一候補として扱うのが自然
- `canonical main` を中心に複数 short を紐づける方が、creator 実態に合う
- creator production model は `camera-first` より `import-first` の方が自然
- 音源は `native library` より `creator持ち込み` の方が初期スコープとして現実的

## 未解決論点

- `fan` 側の retention 機能をどこまで入れるか
- creator の初回 onboarding をどこまでプロダクト化するか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/core-experience.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-profile-and-engagement.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/moderation/moderation-and-review.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/capture-and-audio-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-analytics-minimum.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
