# short-fans Product SSOT - Creator Analytics Minimum

## 位置づけ

- `creator` に最初から何の数字を返すかを整理する
- これは `creator継続率`、`short別学習`、`short -> main` 最適化に直結する

## 現時点の推奨

- `MVP` の creator analytics は `conversion-first` に寄せる
- つまり、`likes / comments / follower vanity` より
  - どの `short` が
  - どの `main` への遷移を生み
  - どれだけ unlock / revenue を生んだか

- を返す

## 推奨理由

### 1. コア体験が `short -> main` だから

- このプロダクトで creator が一番知りたいのは、`バズったか` より `有料につながったか` である
- 特に `1 canonical main : 複数 short` 前提では、`short` 別比較が重要になる

### 2. `MVP` で学習したい単位が明確だから

- `source_short_id -> main_id -> unlock`
- この連鎖が取れること自体が product の強みである
- したがって、analytics もこの object model に揃えるのが自然

### 3. 高度な analytics より先に、意思決定できる最小セットが必要だから

- 初期 creator は `細かい視聴者属性` より
  - どの short を残すか
  - どの main が売れているか
  - どの handoff が強いか

- を知りたい

## MVP で返す数字

### 1. creator overview

- `gross unlock revenue`
- `unlock count`
- `unique purchasers`
- `top performing mains`
- `top performing shorts`

### 2. canonical main 単位

- `paywall views`
- `unlock count`
- `paywall conversion rate`
- `attributed revenue`
- `source shorts`
  - どの `short` から unlock が来たか

### 3. short 単位

- `plays`
- `handoff reach`
  - short の終端近くまで見られた回数 / 率
- `main CTA taps` または `paywall opens`
- `attributed unlocks`
- `short-to-unlock conversion`
- `attributed revenue`

## MVP で返さないもの

- public `like / comment` 前提の engagement analytics
- follower demographics
- viewer active time
- retention cohort
- audience segmentation
- heatmap
- A/B testing suite
- recommended audio / trend discovery

## creator dashboard の最小構成

### overview

- `revenue`
- `unlocks`
- `top mains`
- `top shorts`

### main detail

- 紐づく `short`
- `paywall views`
- `unlocks`
- `conversion`
- `revenue`

### short detail

- `plays`
- `handoff reach`
- `CTA / paywall opens`
- `unlocks`
- `revenue attribution`

## product 上の注意点

- `plays` だけを前面に出しすぎない
- creator に返す主語は `view` ではなく `conversion`
- `review status` や `revision reason` は重要だが、analytics 本体ではなく運用ステータスとして分けて見せる

## 継続仮説

- `creator` は vanity metrics より `売上につながる short` を知りたい
- `handoff reach` は `short-fans` 固有の重要指標になりうる
- `1 本の main に紐づく short 比較` を返せることが、creator 学習速度の差になる

## 未解決論点

- `handoff reach` の具体定義を `最後の0.5秒到達` にするか、別の閾値にするか
- `attributed unlock` の attribution window をどう置くか
- payout / fee / net を creator UI にどこまで出すか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/business/data/data-strategy.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/business/data/data-products-and-governance.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
