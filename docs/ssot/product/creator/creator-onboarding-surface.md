# short-fans Product SSOT - Creator Onboarding Surface

## 位置づけ

- creator onboarding 前後で、どこまで creator UI を見せるかを整理する
- これは `creator化導線`、`trust & safety`、`期待値管理`、`投稿オペレーション` に直結する

## 現時点の推奨

- `MVP` では `approved 前は read-only onboarding surface` を基本に置く
- ただし、`creator profile draft` だけは private に許可する
- `creator profile draft` の範囲は `display name / avatar / bio` に固定する
- `approved 前` の creator preview は `static mock` を基本に置く
- つまり、creator capability を持たない user や onboarding 審査中の user には、`full creator dashboard / upload / submission package editor` は解放しない

## 推奨理由

### 1. creator approval を本当の gate にするため

- すでに `creator審査 approved 前は short / main を public publish できない` 前提を置いている
- ここで upload や投稿導線まで広く見せると、`まだ投稿できないのに作業だけできる` 中途半端な状態が増える
- `approved` を明確な解放点にした方が product も運用もぶれにくい

### 2. review queue を余計に複雑化しないため

- 未承認 user から大量の draft upload を受けると
  - ストレージ運用
  - 下書きサポート
  - abandon された draft
  - moderation 境界

- が一気に重くなる
- MVP は `creator approved -> content submission` の順にした方がよい

### 3. creator の期待値を揃えるため

- onboarding 前に見せるべきなのは `何を出せるか` より `何を満たす必要があるか` である
- つまり、`full tool access` より `requirements / checklist / status` の方が価値が高い

### 4. preview は static mock の方が誤解が少ないため

- `read-only dashboard preview` は、まだ触れない workspace を先に見せるぶん、`もう少しで使える` という期待を生みやすい
- しかし実際には approval 前に `upload / submission package / review action` は使えない
- ここで半端な dashboard shell を出すと、`押せない UI` や `空の table` が増え、未承認 state の価値より frustration が大きくなりやすい
- `static mock` なら、`approval 後に何が解放されるか` を見せつつ、今は onboarding が主目的だと明確にできる

### 5. creator profile draft だけ先に作れると friction を下げられるため

- `display name / avatar / bio` のような low-risk 情報は、approval 前でも下書きできる
- これにより onboarding 完了後の初期設定負荷を下げられる
- 一方で upload や publish は開けないため、trust & safety の gate は維持できる

## surface の分け方

### 1. no creator capability

- 対象
  - fan-only user
  - creator onboarding 未開始 user
- 見せるもの
  - `Become Creator` CTA
  - creator になる条件の要約
  - 年齢 / 本人確認 / payout / consent 責任の説明
  - creator product の `static mock preview`
- 見せないもの
  - upload
  - short-main linkage editor
  - creator analytics
  - moderation queue

### 2. onboarding draft / submitted

- 対象
  - creator onboarding の入力中
  - submitted 後の審査待ち
- 見せるもの
  - private な `creator profile draft`
  - onboarding checklist
  - required docs / status
  - review status
  - policy / prohibited category guidance
  - `approval 後に解放される workspace` の `static mock preview`
- 見せないもの
  - full creator dashboard
  - content upload
  - submission package 作成 UI

### 3. approved creator

- 対象
  - creator capability approved 済み user
- 見せるもの
  - creator dashboard
  - upload / import
  - canonical main + 紐づく short の submission package 作成
  - review status
  - creator analytics

### 4. rejected / suspended

- 対象
  - creator approval が rejected / suspended の user
- 見せるもの
  - status explanation
  - next action
  - `re-submit` 可否
- 見せないもの
  - upload / publish 導線

## MVP の原則

- creator onboarding は `access unlock flow`
- creator dashboard は `approved 後の workspace`
- approval 前に見せる creator UI は `read-only onboarding surface + private creator profile draft` に限定する

## creator profile draft の範囲

- allow
  - `display name`
  - `avatar`
  - `bio`
- disallow
  - public profile publish
  - handle / URL の確定公開
  - content upload
  - submission package 作成

- つまり、`creator profile の見た目の下書き` は許すが、`creator profile の公開面そのもの` は approval 後まで作らない

## creator preview の持ち方

- `MVP` では `static mock` を採用する
- ここでいう `static mock` は、実データや実操作を持たない `example-only` の preview を指す
- preview に含める候補
  - `upload / import`
  - `submission package`
  - `review status`
  - `analytics`
- ただしこれらは `使える UI` ではなく、`approval 後にこういう workspace が解放される` と説明するための固定表示に留める
- `disabled button` や `empty state table` のような、触れそうに見える疑似 UI は極力避ける
- 例外として、`display name / avatar / bio` は user 自身の `creator profile draft` を反映した `light profile preview` として見せてもよい
- つまり、`creator dashboard preview` は static、`profile basics preview` だけ軽く personalize する

## rejected 時の resubmit flow

- `MVP` では `rejected` を全部 support 対応にせず、`fixable reject` に限って self-serve の `resubmit flow` を持つ
- ただし status 自体は細かく増やさず、`creator status = rejected` のまま `再申請可否` と `理由コード` を付ける
- `MVP` では time-based の `cooldown` は基本入れず、`self-serve resubmit 回数上限` で制御する
- 基本の考え方
  - `fixable reject`
    - user 自身が修正して再申請できる
  - `blocked reject`
    - support / manual review を通さないと再申請できない
  - `suspended`
    - self-serve resubmit は不可

### self-serve resubmit を許すもの

- 不足書類
- 書類の画質や判別性の不足
- payout 情報の不備
- name / profile basics の軽微な不整合
- onboarding checklist の未完了

### support-only にするもの

- 年齢要件の不一致
- fraud / impersonation suspicion
- prohibited category への該当
- 同意責任や ownership の重大不整合
- safety 上の高リスク判断

### rejected user に見せるもの

- rejection summary
- 理由コードごとの修正 checklist
- `再申請できる / support に連絡が必要 / 現在は不可` の明示
- self-serve 可のときだけ `Edit And Resubmit` CTA

### self-serve resubmit で編集できる範囲

- onboarding 入力情報
- required docs の再提出
- `display name / avatar / bio`

### self-serve resubmit でも開けないもの

- upload
- submission package 作成
- creator dashboard
- public creator profile 公開

### 回数上限と cooldown

- `MVP` では eligible な `fixable reject` に対して、`即時の self-serve resubmit` を許可する
- つまり、初回 reject 後に一律の待ち時間は置かない
- 代わりに、同じ onboarding case に対する `self-serve resubmit` は最大 `2回` までにする
- `2回` 連続で self-serve resubmit が reject されたら、その case は `support review required` に切り替える
- この時点で `Edit And Resubmit` CTA は閉じ、support 導線に切り替える
- `approved` になった case の resubmit count はリセットする
- 将来 abuse が見えた場合にだけ、`24時間 cooldown` のような時間ベース制御を追加検討する

## 継続仮説

- MVP では `tool を先に触らせる` より `審査条件を理解させる` 方が離脱や混乱を減らしやすい
- creator 候補 user に対しては、dynamic な dashboard より `static mock + 明確な checklist` の方が十分価値がある
- `creator profile draft` は low-risk な供給 friction 低減策として機能する可能性が高い
- `rejected` を全部 manual support に寄せるより、fixable なものだけ self-serve resubmit にした方が onboarding ops を軽くしやすい
- 初期は `cooldown` より `少数回まで即時にやり直せる` 方が supply friction を下げやすい
- supply を増やしたくなった後で、必要なら `pre-approval draft` を解放してもよい

## 未解決論点

- `2回` という self-serve resubmit 上限を将来どこまで緩めるか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/moderation/moderation-and-review.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
