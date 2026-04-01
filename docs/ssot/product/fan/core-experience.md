# short-fans Product SSOT - Core Experience

## 位置づけ

- `fan` 側のコア体験、コンテンツ構造、画面設計の前提をまとめる
- ここでは仕様の芯だけを保持し、UI の細部や演出案は raw メモへ残す

## 安定した事実

- `short-fans` は `short` 起点で視聴が始まるプロダクトとして構想している
- `short` は独立した `sample` コンテンツではなく、`main` へ連続接続する公開面として扱う
- 投稿コンテンツは `short` と `main` を中心に考えており、`写真 / ストーリー / その他形式` は現時点では対象外
- アカウント種別は `creator account` と `fan account` の 2 種のみを想定している
- `creator profile` では `short` を一覧し、`main` は short から遷移して視聴する構造を想定している
- `main` は `完全課金制` かつ `縦型のみ` を前提にしている
- fan 向け UI の canonical は `product/ui/` で管理する
- 初期の主導配布面は `PWA` を基本に置く

## 配布判断の前提

- `App Store` は `overtly sexual or pornographic material` を禁止している
- `UGC / social` アプリでも、主として `pornographic content` に使われるものは `App Store` に載せにくい
- `sexual content` を `paywall` の奥に置くだけでは、`主用途が成人向けか` という審査論点の回避になりにくい
- そのため `short-fans` の本体体験は、少なくとも初期は `App Store` 前提ではなく `PWA` 前提で組む

## 継続仮説

- プロダクトの核は `short -> main` の連続導線
- `short` は casual に見られる non-explicit な surface に寄せることで、閲覧心理ハードルを下げられる
- `main` は short と同じ文脈、同じ画角、同じ流れで始まる方が転換効率が上がる
- `サンプル` を別枠で持たず、`short 自体が公開面と導線` を兼ねる設計がコアである
- `おすすめ / フォロー中` の 2 面構成が、最小の feed 体験として自然
- `Instagram Explore` 的な見せ方の方が、`MyFans` 的なホーム起点より相性がよい可能性が高い

## 未解決論点

- `fan profile` に何を置くか
- `おすすめ` のランキング軸を何にするか
- `Android native` を補助導線としていつ足すか

## 参照

- `/Users/yonedazen/Projects/short-fans/memo/memo_memo/2026-03-31-product-rough-memo.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/index.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/ui/visual-direction.md`
