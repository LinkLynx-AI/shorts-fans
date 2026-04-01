# short-fans Product SSOT - Fan Journey

## 位置づけ

- `fan` 側の主要導線を、画面遷移と目的ベースで整理する
- ここでは MVP の主導線を定義し、枝葉の機能は後ろに置く

## 安定した事実

- `fan` の入口は `short feed` である
- `MVP` では `fan mode` を初回 entry の default に置く
- feed の基準は `TikTok / Instagram Reels` 系の縦スクロール体験
- 最小構成として `おすすめ / フォロー中` の 2 面を想定している
- `MVP` では `creator search` を入れるが、`カテゴリ / Explore / short 検索` は入れない
- `main` は creator profile の一覧から直接選ぶのではなく、short から遷移して視聴する構造を想定している
- `feed / short detail / paywall` の UI 詳細は `product/ui/fan-surfaces.md` を canonical にする
- 初回購入は `mini paywall` を通し、支払い設定と年齢 / 利用同意を完了させる
- 支払い設定と同意が済んだ fan の再購入は、`[ Unlock ¥X | n分 ]` タップからそのまま unlock して `main` に入る
- 1 セッションの primary loop は `feed -> short -> unlock -> main -> feed 復帰` に置く
- `creator profile` は primary loop ではなく、気になった creator を深掘りする二次導線として扱う

## 主導線

1. fan が app を開く
2. `おすすめ` または `フォロー中` で short を見る
3. short の文脈の中で `main` の存在を認識する
4. 初回は paywall、再購入は direct unlock を行う
5. 同じ文脈の続きとして `main` を視聴する
6. creator をフォローするか、次の short へ戻る

## 導線の原則

- `short -> main` の間で文脈が切れないこと
- `main` に行くまでのタップ数と迷いを最小化すること
- creator profile を経由しても、最終的には short から main へ入る構造を崩さないこと
- fan が `何が無料で、何が有料か` を誤解しないこと

## 画面セット

### 1. feed

- `おすすめ / フォロー中`
- short の縦スクロール視聴
- `creator search` への入口を持つ
- UI 詳細は `product/ui/fan-surfaces.md` を参照する

### 2. short player / detail

- short を単体で見る
- UI 詳細は `product/ui/fan-surfaces.md` を参照する
- profile への入口を持つ

### 3. paywall / unlock

- `main` 視聴前の課金面
- UI 詳細は `product/ui/fan-surfaces.md` を参照する
- 初回購入時の setup 面として使う
- 課金後にそのまま main へ入る

### 4. main player

- short の続きとして再生が始まる
- ここが `価値の本体`

### 5. creator profile

- short 一覧を並べる
- fan が creator 単位で深掘りするときの面
- 一覧上では direct な `Unlock` CTA を出さず、short を開いた先で `Unlock` に入る

### 6. fan profile

- `following`、`pinned shorts`、`library`、`settings` をまとめた private hub
- public profile ではなく、再訪と consumer-side 管理の面として使う

## 継続仮説

- `feed -> main` が最重要で、その他の回遊は二次導線
- feed での `視聴完了率` より、`main クリック率 / unlock 率` が勝負になる
- `creator profile` は検索結果より `深掘り先` として使われやすい
- `creator profile` は必要な時だけ入る面に留め、1 セッションの主回遊先にはしない方がよい
- `main` 前に強い説明や情報量を入れすぎると、没入と転換が落ちる可能性が高い
- `continue watching` は `feed` の代替ではなく、paid 再訪専用の補助導線として置く方がよい

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/identity-and-mode-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/fan-profile-and-engagement.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/core-experience.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/fan-surfaces.md`
- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
