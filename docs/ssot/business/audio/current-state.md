# short-fans Business SSOT - Audio Current State

## 位置づけ

- `音源` に関する現時点の事実と、いま置ける前提をまとめる
- ここでは `実現可能性` と `費用感` を混ぜず、まず現在の制約を切る

## 安定した事実

- `Reels / TikTok` の流行り音源を、そのまま `short-fans` に移植できる前提は置けない
- `short-fans` の `MVP` は `native music library` を持たず、creator 持ち込み音源前提で整理している
- `short-fans` では `short` だけでなく `main` も成立させる必要があるため、creator が作る対象は `single short` より `short-to-main package` に近い
- `ASCAP` の `Websites & Mobile Apps` ライセンス案内でも、social account への投稿だけなら `ASCAP` ライセンスは不要だが、blocked される場合は `synchronization license` が必要になりうると案内している
- `The MLC` は、DSP の mechanical royalty は `interactive streaming` などの配信形態ごとに statutory rate で計算されると案内している
- つまり `short-fans` が `Reels` 級の音源体験をやる場合、単一契約ではなく `public performance`、`mechanical`、`master / publishing` の複数レイヤーを跨ぐ可能性が高い

## いま成立する音源レーン

### 1. ほぼ問題なく使いやすいもの

- `original audio`
- creator-owned audio
- royalty-free / separately licensed audio
- spoken hook / ambient / SFX ベースの short

### 2. いま前提にしないもの

- `Instagram / TikTok` の流行り曲を、そのまま `short-fans` でも安全に使える
- `Reels` 風の音源検索 UI だけ先に作れば供給問題が解決する
- `音源あり` でなければ `short -> main` が成立しない

## 費用感の読み方

### A. creator / small business 向けの既存ライセンス

- `Epidemic Sound` の `Business Plan` は、`apps and games` を含む business use をカバーする一方で、`VOD / Streaming VOD / Pay-per-View` は custom solution 側としている
- `Lickd` は、mainstream track の one-off license が `starts from $8` と明示している
- これは `曲単位` や `小規模 business license` は比較的安く始められることを示す
- ただし、これは `プラットフォーム内の巨大音源ライブラリ` を持つコストではない

### B. `Reels / TikTok` 級の catalog 契約

- 公開されている exact contract value はほぼ出てこない
- ただし `UMG` は `2024-01-30` の open letter で、`TikTok` が `UMG` に払う額は `only about 1% of our total revenue` と述べている
- `UMG` の最新公開 FY2025 revenue は `€12,507 million` なので、この `1%` を単純に当てると `約 €125M / year` 規模になる
- これは `UMG単体` の proxy に過ぎず、`Sony`、`Warner`、publisher 側、territory 差分は含まない
- したがって、`short-fans` が将来 `Reels` 級の recognizable song catalog を広く持とうとするなら、費用感は `数十万ドル` や `数百万ドル` ではなく、`少なくとも high eight figures、場合によっては nine figures annual commitment` を覚悟する論点と見るべき
- ここは公式数値そのものではなく、公開情報からの推定である

## 現時点の判断

- `MVP` では `音源なし` に固定するより、`audio-safe` に寄せる方がよい
- つまり初期に受けるのは
  - `無音`
  - `original audio`
  - `話し声 / 環境音`
  - `royalty-free`
  - `commissioned / house audio`

- のどれかで成立する short である
- 外部面では音源ありの流入用shortが動いていても、`short-fans` 内では別shortに音を載せ替える前提の方が現実的

## 未解決論点

- `audio-safe` の定義を product でどこまで明文化するか
- `music rights attestation` を upload 時に product 化するか
- 将来 `limited licensed catalog` をやるなら、どの territory / use case から切るか

## ソース

- ASCAP Websites & Mobile Apps license form: https://licensing.ascap.com/?type=digital
- The MLC rates overview: https://www.themlc.com/rates
- UMG open letter on TikTok, 2024-01-30: https://www.universalmusic.com/an-open-letter-to-the-artist-and-songwriter-community-why-we-must-call-time-out-on-tiktok/
- UMG FY2025 results, published 2026-03-05: https://www.universalmusic.com/universal-music-group-n-v-reports-financial-results-for-the-fourth-quarter-and-full-year-ended-december-31-2025/
- Epidemic Sound Business plan: https://www.epidemicsound.com/our-plans/business-plan/
- Epidemic Sound Business Plan help: https://help.epidemicsound.com/hc/en-us/articles/33405023521554-Business-Plan
- Lickd pricing: https://lickd.co/pricing
- Lickd premium pricing article: https://help.lickd.co/knowledge/how-much-does-a-premium-chart-music-license-cost
