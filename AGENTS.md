# リポジトリ概要

## サービス概要

このリポジトリは、短尺動画を起点にしたアダルトサービスを作るためのものです。

- 既存サービスとの近さでいうと OnlyFans に近いです。
- ただし、MVP の主導線は TikTok / Instagram Reels に近い `short -> main unlock` の縦型連続視聴体験です。
- OnlyFans を参考にするのは、課金導線や creator monetization の追加機能です。
- `MVP` の `fan mode` の初回 entry と home は `shorts` (`feed`) を基準にします。
- `MVP` の global nav / tab bar の exact composition は未確定です。
- ユーザーはまず public な `shorts` フィード上で動画を見ます。
- 公開されている `short` は、未購入ユーザーでも見られます。
- `creator profile` は public な `short` 一覧の深掘り面であり、`main` を直接一覧させず、`short` から遷移して視聴する構造を前提にします。
- `fan profile` は public profile ではなく、`following`、`pinned shorts`、`library`、`settings` などを持つ private consumer hub とします。
- `MVP` の検索は `creator search only` を前提にし、`カテゴリ / Explore / short 検索` は入れません。
- `short` は public に流れる non-explicit な surface、`main` は課金後に視聴する paid continuation として扱います。
- 課金モデルは、`MVP` では `main` 単位の `pay-per-unlock` を第一候補にし、`creator subscription` は後で判断します。
- `MVP` では `1 canonical main : 複数 short` を基本にします。
- `fan` と `creator` は別ログインではなく、`1 user identity + active mode switch` を第一候補にします。
- `creator mode` の home は `dashboard`、`fan mode` の home は `feed` を前提にします。
- 動画は `short` / `main` ともに縦型を前提にします。
- 広告の扱いは現時点で未確定として扱います。

## ドキュメント目次

- [docs/README.md](docs/README.md): ドキュメント全体の索引
- [docs/BACKEND_STRUCTURE.md](docs/BACKEND_STRUCTURE.md): backend のディレクトリ構成方針、依存方向、SQL と transport の配置ルール
- [docs/TECH_STACK.md](docs/TECH_STACK.md): Frontend / Backend / Infra / Payment / Analytics を含む技術選定
- [docs/GO.md](docs/GO.md): Go 実装ルール、package 設計、並行処理、テスト、依存管理の運用ルール
- [docs/TYPESCRIPT.md](docs/TYPESCRIPT.md): TypeScript フロントエンド実装ルールと FSD 運用ルール
- [docs/ssot/LOCAL_INDEX.md](docs/ssot/LOCAL_INDEX.md): 外部 repo から取り込んだ SSOT のローカル入口

## ドキュメント運用ルール

- 文書に明記されたものだけを確定仕様として扱うこと
- 未確定のものは、推測で埋めず未確定と明記すること
- 文書を追加・更新したら、このファイルと `docs/README.md` の索引も揃えること
