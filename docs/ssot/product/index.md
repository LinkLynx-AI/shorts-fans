# short-fans Product SSOT

## 位置づけ

- `product` 領域の安定した前提、継続して参照する仮説、未解決論点を置く
- raw な着想メモは `memo/memo_memo/` に置き、固まったものだけここへ昇格する

## 安定した事実

- `short-fans` は、ショート動画を起点にした `fans` 系プラットフォーム構想
- プロダクトの raw メモは `/Users/yonedazen/Projects/short-fans/memo/memo_memo/` に集約する
- 体験設計は `short` 起点で検討を進めている
- 初期の主導配布面は `PWA` を基本に置く
- fan 向け UI は `product/ui/` で管理する
- `product SSOT` は `fan/`、`ui/`、`content/`、`account/`、`monetization/`、`creator/`、`moderation/`、`scope/` に分けて管理する
- `product SSOT` は `core-experience`、`content-model`、`fan-journey`、`billing-and-access`、`short-main-linkage`、`identity-and-mode-model`、`account-permissions`、`consumer-state-and-profile`、`fan-profile-and-engagement`、`moderation-and-review`、`creator-onboarding-surface`、`creator-workflow`、`capture-and-audio-model`、`creator-analytics-minimum`、`mvp-boundaries`、`visual-direction`、`fan-surfaces` を主軸に整理する
- `課金とアクセス制御` は `billing-and-access` で管理する
- `identity / mode` の持ち方は `identity-and-mode-model` で管理する
- `creator / fan` の権限差分は `account-permissions` で管理する
- `consumer-side state` と `fan profile` は `consumer-state-and-profile` で管理する
- `fan profile` の具体UIと `engagement` の範囲は `fan-profile-and-engagement` で管理する
- `審査 / moderation` は `moderation-and-review` で管理する
- `creator onboarding 前後の UI 露出` は `creator-onboarding-surface` で管理する
- `approved 前` の creator preview は `static mock` を基本に置く
- `creator production model` は `capture-and-audio-model` で管理する
- `creator analytics` の最小要件は `creator-analytics-minimum` で管理する

## 継続仮説

- コア体験は `short -> main` の連続導線
- `short` は外から見ても成立する non-explicit な動画体験に寄せる
- `main` は完全課金制で、縦型・連続視聴を前提にする
- creator 投稿フローには `動画インポート -> short 切り出し -> main 紐付け` が必要になる可能性が高い
- creator production model は `import-first` が妥当
- 初期音源は `native music library` ではなく `creator持ち込み` が妥当
- creator 体験では、投稿管理だけでなく `分析 / API / MCP` が差別化要素になりうる

## 未解決論点

- `非離散的動画` をプロダクトルールとしてどこまで強制するか
- main へのジャンプ UX をどう最短化するか
- `PWA` を主導面に置いた上で、`Android native` をどこまで補完的に足すか
- 将来 `utility capture camera` を足すか
- 将来 `rights-cleared audio catalog` を足すか
- creator analytics の deep metrics をどこまで足すか

## 構造

- core experience: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/core-experience.md`
- ui index: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/index.md`
- visual direction: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/visual-direction.md`
- fan surfaces: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/fan-surfaces.md`
- content model: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
- fan journey: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- billing and access: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- short main linkage: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- identity and mode model: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- account permissions: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- consumer state and profile: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/consumer-state-and-profile.md`
- fan profile and engagement: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-profile-and-engagement.md`
- moderation and review: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/moderation/moderation-and-review.md`
- creator onboarding surface: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-onboarding-surface.md`
- creator workflow: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- capture and audio model: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/capture-and-audio-model.md`
- creator analytics minimum: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-analytics-minimum.md`
- mvp boundaries: `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/scope/mvp-boundaries.md`

## 参照

- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-onlyfans-revenue.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/core-experience.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/index.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/visual-direction.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/fan-surfaces.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/consumer-state-and-profile.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-profile-and-engagement.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/moderation/moderation-and-review.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-onboarding-surface.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/capture-and-audio-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-analytics-minimum.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/scope/mvp-boundaries.md`
