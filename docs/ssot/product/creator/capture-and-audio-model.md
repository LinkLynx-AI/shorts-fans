# short-fans Product SSOT - Capture And Audio Model

## 位置づけ

- `creator` が `short` と `main` をどう作るか、その production model を整理する
- これは `アプリ内撮影`、`音源`、`投稿UI`、`権利処理`、`MVPの切り方` に直結する

## 現時点の推奨

- `MVP` の creator production model は `import-first` を基本に置く
- つまり、`Reels / TikTok` のような `音源付き social camera` を中核にはしない
- creator はまず外部で素材を撮影 / 編集し、`short-fans` では
  - `canonical main`
  - `紐づく short`
  - `short -> main` の handoff
  - submission package

- を整える
- `MVP` の音源方針は `native music library なし` を基本に置く
- `short-fans` 側で扱う音は、creator が持ち込む `embedded audio` または creator 自身が権利を持つ音源に限る

## 推奨理由

### 1. このプロダクトは `short` 単体ではなく `main` まで必要だから

- `Reels / TikTok` 型の camera は、基本的に `1投稿` を作るための UI である
- しかし `short-fans` では、creator は `short` だけでなく `main` も成立させる必要がある
- しかも現在の前提では、`1 canonical main : 複数 short` が自然である
- この時点で必要なのは `single-post camera` より `project/package authoring` に近い

### 2. 重要なのは `一発撮り` より `連続して見えること` だから

- raw メモでも、creator は `複数 take + 確認` を挟む前提が強い
- したがって product が支えるべきなのは、`ノンストップ録画` より
  - どこで `short` が終わるか
  - どこから `main` が始まるか
  - 別の `short` が同じ `main` に自然につながるか

- の方である

### 3. 音源ライブラリは UI より権利処理が本体だから

- Instagram や TikTok は `original audio` と `music library` の両方を前提にしている
- ただし、これは各社が music rights holder と契約を持っている前提の機能である
- `short-fans` がこの部分だけ UI を真似しても、権利処理なしでは成立しない

### 4. 初期は `viewing app` を `creation suite` にしすぎない方がよいから

- creator 側を強くしすぎると、プロダクトの中心が `feed -> short -> main unlock` から外れやすい
- `MVP` では `capture` より `import / linkage / review / analytics` の方が product fit に直結する

## creator production model

### 推奨フロー

1. creator は device camera や外部 editor で素材を撮る
2. creator は `canonical main` を書き出す
3. 同じ session / outfit / angle から `short` を 1 本以上作る
4. `short-fans` に `main` と複数 `short` を upload / import する
5. `short -> main` の continuity が成立するように紐付ける
6. `submission package` として review に出す

### ここで `short-fans` が支えるべきこと

- `canonical main` と紐づく `short` の object 管理
- `1 main : n short` の linkage
- `short` ごとの handoff point / continuity 確認
- review status と revision 理由の返却
- `short -> main` の conversion analytics

### ここで `short-fans` が最初はやらないこと

- `Reels` 風の音源検索付きカメラ
- multi-track timeline editing suite
- licensed music catalog
- 高度な audio mixing / replacement

## 音源方針

### MVP in

- upload された動画に含まれる `embedded audio` の取り込み
- creator が rights を持つ前提の音源利用
- `short` と `main` を別 asset として持ち込む運用

### MVP out

- `Instagram / TikTok` 型の native music library
- アプリ内の licensed music search
- rights-managed な音源 attribution system
- 別 audio track の本格的な mix/edit

### 実務上の意味

- creator が `Instagram / TikTok` 上で platform-licensed audio を使っていたとしても、そのまま `short-fans` に持ち込めるとは限らない
- `short-fans` 用には
  - original audio
  - creator-owned audio
  - royalty-free / separately licensed audio
  - 無音または ambient ベース

- のどれかで成立させる必要がある

## 将来の拡張候補

### 1. utility capture camera

- `social camera` ではなく、raw capture 用の簡易 camera
- 狙い
  - 画角固定
  - guide overlay
  - 同一 session の take 管理
- ただし `MVP` では不要

### 2. light audio tooling

- mute
- basic volume trim
- short/main での簡易 audio separation

### 3. rights-cleared catalog

- 将来的にやるなら、`Instagram/TikTok` の UI 模倣ではなく、rights-cleared な catalog とセットで考える

## 継続仮説

- `short-fans` に必要なのは `record a reel` より `author a short-to-main package`
- creator の実態に近いのは `camera-first` より `import-first`
- `main` まで必要なプロダクトでは、`music camera` より `continuity editor` の方が価値が高い
- 初期音源は `owned/original only` に寄せた方が product と法務がぶれにくい

## 未解決論点

- 将来 `utility capture camera` を足すか
- `short` を `main` と同じ source timeline から切り出す編集モデルを product に持つか
- creator が raw source を保存する `project` 概念まで必要か

## 参照

- `/Users/yonedazen/Projects/short-fans/memo/research/2026-03-31-short-video-creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/scope/mvp-boundaries.md`
