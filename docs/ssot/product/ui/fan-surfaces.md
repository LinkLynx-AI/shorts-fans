# short-fans Product SSOT - Fan Surfaces

## 位置づけ

- `fan` が触る主要 surface の情報配置と primary action を整理する
- `journey` の話ではなく、各 surface に何を置くかの canonical を置く

## 安定した事実

- `fan` の入口は `short feed` である
- `feed / short detail / paywall` は `Instagram Reels` をベースに 9 割採用で設計する
- `short` 内の primary CTA は、常時同一 placement の下部固定 `[ Unlock ¥X | n分 ]` を基本に置く
- `creator block` は `Instagram Reels` と同じ考え方にし、`profile 入口` を別導線として分離しない
- CTA は画面下部に置き、`creator block` より上に置く
- 初回課金前の paywall は、`mini paywall` を基本にする
- 支払い設定と同意が済んだ再購入では、CTA タップから full paywall を開かずに unlock して `main` に入る
- 1 セッションの primary loop は `feed` 側に置き、`creator profile` は二次導線にする
- `MVP` の検索は `creator search only` とし、`カテゴリ / Explore / short 検索` は入れない

## Surface 定義

### feed

- `おすすめ / フォロー中` の 2 面を最小構成とする
- short の縦スクロール視聴を基本にする
- `creator search` への入口を持つ
- `like` は持たない
- `comment` は持たない
- `creator block` と caption は `Instagram Reels` と同じ考え方で置く
- その上に下部固定の `[ Unlock ¥X | n分 ]` CTA を置く
- `pin` は primary CTA にはしないが、後で見返して `unlock` につながる前提で secondary action として持つ

### short detail

- short を単体で見る surface とする
- `like` は持たない
- `comment` は持たない
- `creator block` と caption は `Instagram Reels` と同じ考え方で置く
- `[ Unlock ¥X | n分 ]` CTA を置く
- `pin` は secondary action として置ける

### paywall / unlock

- `この short の続き` を最短で買うための surface とする
- `like` は持たない
- `comment` は持たない
- `creator block` と caption は `Instagram Reels` と同じ考え方で置く
- `[ Unlock ¥X | n分 ]` CTA を置く
- `価格` と `main` の長さは CTA 内に含める
- 初回だけ、支払い方法選択と年齢 / 利用同意を追加する
- 再購入では full paywall を開かず、CTA タップからそのまま unlock して `main` に進める
- explicit preview、複数サムネ、cover 画像、他作品レコメンドは置かない
- 購入後は確認面を挟まず、そのまま `main` 再生へ進める

### creator profile

- `Instagram profile` に近い header と short 一覧を基本にする
- 一覧上では `Unlock`、価格、`main` 長さを直接出さない
- `follow` は profile header 側に置く
- short を開いた先の `short detail` で `[ Unlock ¥X | n分 ]` を出す
- つまり `creator profile` は深掘り面だが、`main` への direct list surface や主回遊面にはしない

### search

- `creator search only` を基本にする
- 検索キーは `display name / handle` に絞る
- 検索結果は `creator profile` に落とす
- `short caption / hashtag / full-text search` は持たない

### fan profile

- `private consumer hub` として扱う
- `Following`、`Pinned Shorts`、`Library`、`Settings` を持つ
- `Pinned Shorts` では fan 自身が pin した short を見返せる
- `Pinned Shorts` から `short detail` を開き、そこから `unlock` へ入れる

### main player

- short の続きとして再生が始まる
- `main` の価値は `続きをそのまま見られること` に置く

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-journey.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
