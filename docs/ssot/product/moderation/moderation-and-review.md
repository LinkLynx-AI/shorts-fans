# short-fans Product SSOT - Moderation And Review

## 位置づけ

- `creator審査 / short審査 / main審査 / 通報後対応` の 4 層で、`short-fans` の審査モデルを定義する
- これは `creator onboarding`、`公開可否`、`決済 / 法務リスク`、`trust & safety` の前提になる

## 安定した事実

- `short` は public surface であり、課金前に explicit に見えてはいけない
- `main` は課金後に視聴する本編であり、`short` の続きとして扱う
- `creator mode` の private surface では、投稿ステータスや審査状態を表示する前提を置いている
- `MVP` では `1 canonical main : 複数 short` を投稿単位にしている
- `platform` は MVP でも最低限の review / moderation operation を持つ必要がある
- `MVP` では review intake を `1 canonical main + 紐づく複数 short` の `submission package` として扱う

## 現時点の推奨

- `MVP` は `4-layer review model + manual-heavy operation` を基本に置く
- review intake は `submission package`
- review decision state は `creator / main / short` で分ける
- `creator rejected` には `再申請可否` と `理由コード` を持たせる
- 理由
  - `short` は public surface なので、explicit 境界の事故がそのままブランド毀損になる
  - `main` は課金面と法務面の両方のリスクを持つ
  - 成人領域では、`creator本人確認 / 出演者同意 / prohibited category` を軽く始めると後で戻せない

## review intake model

### A. `short / main` を完全に別 submission として扱う

- 強み
  - object ごとの処理は単純
- 弱み
  - `short -> main` の連続性を review 時に見づらい
  - creator の投稿体験が分断されやすい

### B. `submission package` として intake し、decision state は分ける

- 強み
  - `main` と紐づく `short` の文脈をまとめて見られる
  - creator 実態の投稿単位と合う
  - continuity や prohibited category を一度の intake で把握しやすい
- 弱み
  - reviewer 側 UI は少し複雑になる

## submission package の構成

- `canonical main`
- 紐づく `short`
- continuity が分かる metadata
- 必要な creator / consent / ownership 情報

## review output の考え方

- reviewer は package 全体を見て判断する
- ただし結果は分けて持つ
  - `creator status`
  - `main status`
  - `short status`

## 4 層モデル

### 1. creator審査

- 対象
  - creator onboarding
  - 投稿権限の付与
- 主目的
  - `誰が投稿するか` を確定する
  - `年齢 / 本人性 / payout / 同意責任` の基点を作る
- MVP で確認するもの
  - 年齢確認
  - 本人確認
  - payout 受取主体
  - 禁止カテゴリ該当有無
  - 必要な同意責任を creator が負うこと
- status 例
  - `draft`
  - `submitted`
  - `approved`
  - `rejected`
  - `suspended`
- rejection metadata
  - `re-submit eligible`
  - `support review required`
  - `reason code`
- publish rule
  - `creator approved` 前は、short / main を public publish できない

### 2. short審査

- 対象
  - public に出る各 `short`
- 主目的
  - `public surface` として安全かを判定する
- MVP で確認するもの
  - explicit になっていないか
  - `short -> main` の文脈が成立しているか
  - 同じ `main` に紐づく `short` として不自然ではないか
  - 禁止カテゴリに触れていないか
  - 身バレや location leakage のリスクが強すぎないか
- status 例
  - `draft`
  - `pending review`
  - `approved for publish`
  - `revision requested`
  - `rejected`
  - `removed`
- publish rule
  - `short` は review 完了前に public feed へ出さない

### 3. main審査

- 対象
  - `canonical main`
- 主目的
  - `課金対象として公開してよいか` を判定する
- MVP で確認するもの
  - 出演者の年齢 / 同意に問題がないか
  - prohibited category に触れていないか
  - hidden camera、revenge porn、non-consensual の疑いがないか
  - short と main のつながりが成立しているか
  - payout 対象の ownership が明確か
- status 例
  - `draft`
  - `pending review`
  - `approved for unlock`
  - `revision requested`
  - `rejected`
  - `suspended`
  - `removed`
- publish rule
  - `main approved` 前は linked short を publish しない
  - 理由
    - short は `main` への導線なので、飛び先も review 済みである必要がある

### 4. 通報後対応

- 対象
  - creator
  - short
  - main
  - rights / consent / safety complaints
- 主目的
  - publish 後の問題を止血し、再発を防ぐ
- 入口
  - fan report
  - creator report
  - rights holder claim
  - internal moderation flag
- MVP の基本アクション
  - `received`
  - `triaged`
  - `temporarily limited`
  - `takedown`
  - `creator hold / suspension`
  - `resolved`
- high-risk の扱い
  - `minor`
  - `non-consensual`
  - `hidden camera`
  - `imminent safety risk`

- これらは通常 queue ではなく `immediate hold` を第一候補に置く

## 審査単位の考え方

- 投稿の creator-side 編集単位は `1 canonical main + 複数 short`
- review intake も、その単位の `submission package` を基本に置く
- ただし review status は分ける
  - `creator status`
  - `main status`
  - `short status`
- つまり
  - creator は approved 済み
  - main は pending
  - short A は approved
  - short B は revision requested

- のような状態を持てる

## MVP の運用方針

### manual-heavy

- `creator審査`
  - manual 必須
- `short審査`
  - manual pre-publish を第一候補
- `main審査`
  - manual pre-publish を第一候補
- `通報後対応`
  - manual triage 必須

### なぜ manual-heavy か

- `short` の public/non-explicit 境界がこのプロダクトのコアだから
- 初期は supply が限定されるので、manual review のオペ負荷がまだ管理可能だから
- 先に運用ルールを学習しないと、後から automation しても品質が安定しないから

### repeat creator の優先審査

- `repeat creator` に対しては、`full skip` ではなく `優先審査` を基本に置く
- MVP の付与条件
  - creator approval が有効
  - 直近 `3` 件の submission package が approved 済み
  - 過去 `60日` に `post-report takedown / suspension / non-consensual suspicion` がない
  - open な high-risk report がない
- 優先審査の効果
  - queue priority を上げる
  - creator review は再利用する
  - main review は `delta / consent / ownership / prohibited category` 中心に見る
  - short review は pre-publish を維持しつつ、`public safety` の確認に絞る
- 優先審査でも外さないもの
  - short の pre-publish review
  - main の pre-unlock review
  - high-risk claim 時の immediate hold
- revoke 条件
  - serious report
  - policy breach
  - `2回連続` の `revision requested`
  - ownership / consent の不整合

### 優先審査の将来緩和

- `MVP` 後に緩める対象は、`入口条件` と `review scope` までに留める
- 緩めないもの
  - short の pre-publish review
  - main の pre-unlock review
  - open high-risk report の排除
  - high-risk claim 時の immediate hold
- 将来の緩和上限
  - `直近3件 approved` は `直近2件 approved` まで緩めてよい
  - `過去60日 clean` は `過去30日 clean` まで緩めてよい
  - ただし `1件 approved` まで落とすのはやりすぎとする
- 理由
  - `2件 approved` あれば、少なくとも `repeat creator` としての再現性と比較対象ができる
  - `1件 approved` だけでは、安定運用 creator か単発通過かを見分けにくい
  - `優先審査` は speed-up であって trust の代替ではない
- したがって、中長期の第一候補は `2件 approved + 30日 clean + no open high-risk` とする
- それより先に緩めるより、先に reviewer tooling や reason-code 学習を強くする方がよい

## creator に見せる審査状態

- `creator審査`
  - `submitted / approved / rejected / suspended`
- `main`
  - `draft / pending review / approved for unlock / revision requested / rejected`
- `short`
  - `draft / pending review / approved for publish / revision requested / rejected`
- `post-report`
  - `under review / temporarily limited / removed / restored`

## creator rejection の扱い

- `MVP` では `rejected` を 2 つの運用に分ける
  - `fixable reject`
    - user が自力で修正し、同じ onboarding flow から `resubmit` できる
  - `blocked reject`
    - support / manual review を通すまで self-serve resubmit を開けない
- `fixable reject` の主な例
  - 書類不足
  - 書類画質不足
  - payout 情報不足
  - 軽微な name / profile mismatch
- `blocked reject` の主な例
  - 年齢要件不一致
  - impersonation / fraud suspicion
  - prohibited category
  - 同意や ownership の重大不整合
  - safety high-risk
- `MVP` の resubmit guardrail
  - fixable reject には一律 cooldown を置かない
  - 同じ onboarding case の self-serve resubmit は最大 `2回`
  - `2回` 連続で reject されたら `support review required` に切り替える
- つまり、`creator status` は `rejected` のままでも、`次に何ができるか` は metadata で分けて扱う

## 継続仮説

- `short` は `main` より review 基準が厳しい
  - 理由
    - public surface だから
- `main` は explicit そのものより、`consent / legality / prohibited category` の方が主要論点になる
- `creator審査` を重くし、`通報後対応` を速くする方が、全投稿完全事前審査より現実的な運営モデルになる可能性が高い
- 一方で `MVP` の最初は、public surface の事故を避けるため `short` はほぼ事前 review に寄せるのがよい
- `repeat creator` の簡略化は、`skip` より `優先審査 + review scope reduction` の方が安全に始めやすい
- `優先審査` は、`3件/60日` から `2件/30日` までは緩めうるが、`1件 approved` まで下げるべきではない

## 未解決論点

- `revision requested` の差し戻し UI をどこまで作るか
- `creator rejected` の self-serve resubmit 上限 `2回` を将来どこまで緩めるか
- internal moderator / operator 権限を product SSOT 上でどう切るか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/scope/mvp-boundaries.md`
