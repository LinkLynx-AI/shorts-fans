# short-fans Business SSOT - Data Products And Governance

## 位置づけ

- `short-fans` のデータを、`誰に / どの粒度で / どの形で` 提供するかを定義する
- `data-strategy` が方向性なら、こちらは実務上の線引きを置く

## データの階層

### Tier 0. Raw event data

- impression
- watch time
- session path
- unlock event
- purchase history
- device / country / timestamp
- user-level identifier

扱い:
- `社内限定`
- recommendation、pricing、fraud、ops のために使う
- creator や外部パートナーへそのまま出さない

### Tier 1. Creator-scoped analytics

- 自分の short ごとの impression
- completion rate
- main click rate
- unlock conversion rate
- price ごとの転換差
- repeat purchase
- follow conversion

扱い:
- `当該 creator に返す`
- user-level の fan identity は含めない
- 基本は dashboard と periodic report の形で返す

### Tier 2. Benchmark analytics

- 国別平均 unlock rate
- カテゴリ別平均 completion rate
- 価格帯別の転換中央値
- creator size 別の benchmark
- short pattern 別の傾向

扱い:
- `creator / agency / partner` に提供可能
- ただし十分な集計母数を持つ匿名化済みデータに限定する
- 個別 creator や特定 fan を逆算できる粒度では出さない

### Tier 3. Public insight

- 市場トレンド
- フォーマット傾向
- 国別の高レベルな需要差

扱い:
- public report や LP にも転用可能
- 極めて粗い粒度に限定する

## creator に返すデータ

### 返してよいもの

- short ごとの再生数
- short ごとの完走率
- short -> main クリック率
- main unlock rate
- 価格ごとの転換傾向
- creator 内比較
- 時系列推移

### 返さないもの

- fan 個人の閲覧履歴
- fan 個人の cross-creator 行動
- 生の session replay 的情報
- payment method 詳細
- 再識別可能な細粒度ログ

### 返し方

- 기본は dashboard
- 高額 creator や agency には periodic report
- 将来的には benchmark を含む上位プランを設計可能

## 外部に出してよいデータ

### 出してよい形

- 匿名化された benchmark
- カテゴリ別 / 国別 / price bucket 別の傾向
- creator cohort 単位の aggregated insight
- trend report
- optimization consulting の一部としての集計結果

### 出してはいけない形

- raw event export
- user-level 行動ログ
- creator を特定できる比較表
- 少数サンプルの集計
- 契約目的を超える再利用

## monetization の現実的な順番

1. 社内最適化
2. creator analytics
3. creator / agency benchmark
4. partner report / advisory

補足:
- `生データ販売` を先にやるのではなく、`プロダクト化された insight` を売る方が筋がよい

## 継続仮説

- 一番強い moat は `raw data` を売ることではなく、`raw data から最適化知見を作れること`
- creator が本当に欲しいのは `数字の羅列` ではなく、`何を変えると unlock が上がるか`
- B2B 側で価値が出るのは、`dataset` 単体より `benchmark + interpretation + recommendation`

## ガバナンス原則

- `need-to-know` で扱う
- 目的外利用をしない
- 再識別リスクを常に下げる
- 集計母数が小さいものは外に出さない
- 規約 / consent / disclosure を product 設計と一緒に詰める

## 未解決論点

- benchmark の最小母数をどこに置くか
- creator に high-spender fan 情報をどこまで返すか
- agency 契約で何を標準提供にするか
- data room 的な提供を許すか
- 規約文面をどう設計するか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/business/data/data-strategy.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
