# short-fans Product SSOT - Creator Workflow

## 位置づけ

- `creator` 側の撮影、投稿、分析、運用の前提をまとめる
- feed UX とは分けて、供給面のボトルネックを明確にする

## 安定した事実

- 想定 creator は `個人〜少人数チーム`
- creator は外部プラットフォームにも出せる `非ポルノ short` を入り口に使う想定がある
- creator 投稿フローそのものは、現時点ではまだ十分に把握できておらず、未確定部分が大きい
- `MVP` の投稿単位は `canonical main + 複数 short` として考える
- `MVP` の review intake は `canonical main + 複数 short` の `submission package` を基本に置く
- `repeat creator` には `優先審査` を適用しうるが、`short pre-publish review` は維持する
- `優先審査` は将来的に `直近2件 approved + 過去30日 clean` までは緩めうるが、`1件 approved` まで下げない前提を置く
- production / audio の canonical は `capture-and-audio-model.md` に置く
- analytics の canonical は `creator-analytics-minimum.md` に置く
- `creator onboarding rejected` のうち fixable なものは、同じ onboarding flow から self-serve で再申請できる前提を置く
- 同じ onboarding case の self-serve 再申請は `最大2回` を基本に置き、その後は support review に切り替える
- 現時点で必要機能として、`動画インポート`、`short 切り出し`、`main 紐付け`、`creator 向け分析 / 管理` が挙がっている
- creator 側の差別化候補として、`API / MCP` を強くする案がある
- creator の投稿 / 分析 / 審査操作は `creator mode` の private surface として扱う
- `creator mode` の home は `dashboard` を第一候補に置く
- `MVP` の審査モデルは `creator審査 / short審査 / main審査 / 通報後対応` の 4 層を基本に置く

## 継続仮説

- creator は `一発撮り` より `複数 take + 後確認` で制作することが多い
- 重要なのは `ノーカット` であることより、`連続して見えること` である
- 同じ画角、同じ導線を保った `複数 short -> 1 main` は、MVP でも自然な運用になりうる
- 投稿 UI と審査 UI の両方で、`submission package` 単位の方が creator にとって理解しやすい
- `profile basics draft` を approval 前に持てると、approved 後の立ち上がりは速くなる可能性が高い
- `rejected` を全部 support 送りにせず、fixable なものだけ self-serve resubmit にすると onboarding が詰まりにくい
- `main` まで必要な product では、`music camera` より `continuity / linkage` の方が価値が高い

## 運用上の制約

- 外撮りや場所特定につながる映像は `身バレリスク` がある
- creator 側の撮影支援を強めるほど、プロダクトは `viewing app` ではなく `creation tool` に近づく
- creator の継続率を上げるには、投稿機能だけでなく `分析` と `運用支援` が必要になる可能性が高い

## 未解決論点

- `utility capture camera` を将来足すか
- `main` と `short` を同じ source timeline から切る project モデルを持つか
- creator analytics の `handoff reach` をどう定義するか
- `API / MCP` を誰向けにどこまで開くか
- 1 本の `main` に紐づける `short` の上限や投稿条件をどうするか

## 参照

- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-short-video-creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-onboarding-surface.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/moderation/moderation-and-review.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/capture-and-audio-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-analytics-minimum.md`
