# short-fans Business SSOT - Data Strategy

## 位置づけ

- `short-fans` が `short -> main` 導線から得られるデータを、どう価値化するかを整理する
- ここでは `内部活用`、`creator向け活用`、`外部提供` を分けて考える

## 安定した事実

- `short-fans` のコア体験は `short -> main` の連続導線である
- したがって、`閲覧` から `課金` までが 1 本のトラッカブルな流れとして取得できる可能性が高い
- `main 単位 unlock` を前提にすると、`どの short がどの main unlock を生んだか` を直接ひも付けやすい
- adult / creator 領域のデータは高感度であり、`生の個人データ販売` は信頼・法務・規制の観点でリスクが高い

## まず取るべきデータ

- `impression`
- `view start`
- `view completion`
- `rewatch / loop`
- `profile click`
- `main click`
- `unlock conversion`
- `price`
- `repeat purchase`
- `creator follow`
- `session path`

## 価値化のレイヤー

### 1. 内部活用

- recommendation の改善
- `どの short パターンが unlock を生むか` の学習
- 価格最適化
- `AIシーズニング` の効果検証
- 国 / カテゴリ / creator タイプ別の供給最適化

### 2. creator 向け提供

- `どの short が main unlock を生んだか`
- `どの導線で離脱したか`
- `どの価格帯で転換したか`
- `どの構図 / テンポ / 導線が強いか`
- つまり、`view analytics` ではなく `conversion analytics` を売れる可能性がある

### 3. 企業 / パートナー向け提供

- raw user data ではなく、`匿名化・集計済み` の trend insight
- 例
  - 国別の需要傾向
  - カテゴリ別の unlock 傾向
  - short パターン別の転換傾向
  - creator 供給と fan 需要のギャップ
- 想定相手
  - creator agency
  - production team
  - marketing / growth partner
  - 決済 / インフラ検討先

## 継続仮説

- 一番価値が高いのは `動画視聴データ` ではなく `視聴 -> unlock` の因果に近いデータ
- `short-fans` は、`どの冒頭がどの有料本編に繋がるか` というユニークな conversion dataset を持てる可能性がある
- これは単なる広告 CTR データより、creator economy に近い `purchase-intent data` として強い
- B2B monetization があるとしても、最初に売るべきは `データそのもの` より `ダッシュボード / ベンチマーク / レポート / 最適化支援`
- raw data sale は短期収益化の誘惑があっても、長期的な trust と法務コストに対して割が悪い

## やってはいけない方向

- 生の user-level 行動履歴を第三者へ売る
- creator や fan を再識別できる形で外部提供する
- 同意範囲を超えた二次利用を前提にする
- `データが売れる` ことを前提に、取得過剰なプロダクト設計にする

## 外部提供の現実的な形

- creator 向け analytics SaaS
- creator / agency 向け benchmark report
- 国 / カテゴリ別の market insight report
- 特定 partner との `集計済み insight` 契約
- 運営受託型の growth optimization 支援

## 未解決論点

- どの粒度までなら `匿名化・集計済み` と言えるか
- creator にどこまでデータを返すか
- pricing を SaaS 化するか、report 化するか
- データ利用規約と consent をどう設計するか
- 社内で `data moat` として抱える部分と、外に出す部分をどこで分けるか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/business/revenue/revenue-model.md`
