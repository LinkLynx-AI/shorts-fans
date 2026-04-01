# short-fans Product SSOT - Content Model

## 位置づけ

- `short-fans` のコンテンツ種別、アカウント種別、公開範囲の前提をまとめる
- UI 表現ではなく、プロダクトの構造を定義する

## 安定した事実

- アカウント種別は `creator account` と `fan account` の 2 種のみを想定している
- `creator / fan` の差は単なる表示差分ではなく、`mode ごとの権限差分` として整理する
- コンテンツ種別は `short` と `main` を中心に考えている
- `写真`、`ストーリー`、その他 `short / main` 以外のコンテンツは現時点では対象外
- `short` は独立した `sample` ではなく、`main` に連続接続する公開面として扱う
- `MVP` では `1 canonical main : 複数 short` を基本に置く
- `main` は `完全課金制` かつ `縦型のみ` を前提にしている
- `creator profile` では `short` を並べ、`main` は short から遷移して視聴する構造を想定している
- `pin` は独立コンテンツ種別ではなく、fan が `short` に付ける private な保存 state として扱う

## コンテンツ構造

### account

- `creator account`
  - short と main を投稿する
  - profile 上では short が公開面になる
- `fan account`
  - short を閲覧する
  - 課金後に main を視聴する

### short

- 公開面に流れる縦動画
- feed と creator profile の基本単位
- 役割は `発見`、`視聴継続`、`main への導線`
- 独立した teaser ではなく、`main の冒頭として意味がつながっている` ことが重要
- `MVP` では 1 本の short は 1 本の canonical main にだけつながる

### main

- 課金後に視聴する本編
- `short` と同じ文脈・同じ流れで接続する前提
- 形式は `縦型動画` に限定する
- `MVP` では 1 本の main は 1 本以上の公開 short を持てる

### profile

- `creator profile`
  - public surface は short 一覧
  - main は直接一覧せず、short から辿る
  - 一覧上で direct な `Unlock` は出さず、short を開いた先で課金導線を見せる
- `fan profile`
  - `private consumer hub` を第一候補に置く
  - `pinned shorts` を持つ

## 継続仮説

- `short-fans` のコンテンツモデルは、`sample + locked content` ではなく `public opening + paid continuation` として理解する方が正しい
- `MVP` でも `1 main : 複数 short` の構成を取る可能性が高い
- `short` は public に流しても違和感がない non-explicit な surface に寄せるべき
- `main` の価値は、単独の本編というより `short の続きをそのまま見られること` にある
- `creator profile` は `作品一覧ページ` ではなく、`short feed の延長` に近い構造の方が自然

## 制約

- `short` 単体で explicit になってはいけない
- `main` への遷移前に、`課金しないと明示的部分は見えない` ことが維持される必要がある
- コンテンツ形式を増やしすぎると、`short -> main` というコアがぼやける

## 未解決論点

- `いいね / コメント / シェア` をどこまで入れるか
- 1 本の `main` に何本の `short` を紐づけるか
- short と main の紐付け単位を、将来的に `シリーズ単位` まで持つか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/account/account-permissions.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/consumer-state-and-profile.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/fan/core-experience.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/short-main-linkage.md`
- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
