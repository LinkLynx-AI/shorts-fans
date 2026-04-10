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
- [docs/infra/dev-media-sandbox.md](docs/infra/dev-media-sandbox.md): dev 用 AWS media sandbox の Terraform 構成、guardrail、適用手順、backend 接続用 env 対応
- [docs/infra/dev-media-smoke.md](docs/infra/dev-media-smoke.md): dev media smoke の再現手順、切り分け、manual recovery、quota/cost 注意点
- [docs/contracts/mvp-core-domain-contract.md](docs/contracts/mvp-core-domain-contract.md): MVP core 永続化タスク向けの実装用ドメイン契約
- [docs/contracts/mvp-media-workflow-contract.md](docs/contracts/mvp-media-workflow-contract.md): MVP media workflow の状態遷移、delivery 境界、avatar 境界を整理した実装用契約
- [docs/contracts/media-display-access-contract.md](docs/contracts/media-display-access-contract.md): delivery-ready asset を public short / main playback / owner preview に出す display/access 境界契約
- [docs/contracts/fan-auth-api-contract.md](docs/contracts/fan-auth-api-contract.md): fan sign in / sign up / session start / logout の transport 契約
- [docs/contracts/viewer-bootstrap-api-contract.md](docs/contracts/viewer-bootstrap-api-contract.md): app shell が読む current viewer bootstrap の read 契約
- [docs/contracts/viewer-creator-entry-api-contract.md](docs/contracts/viewer-creator-entry-api-contract.md): fan profile から始める creator registration / active mode switch の transport 契約
- [docs/contracts/creator-workspace-api-contract.md](docs/contracts/creator-workspace-api-contract.md): `/creator` workspace の creator info / overview / revision summary 契約
- [docs/contracts/creator-workspace-owner-preview-api-contract.md](docs/contracts/creator-workspace-owner-preview-api-contract.md): creator owner 向け short/main preview list/detail の read 契約
- [docs/contracts/creator-upload-api-contract.md](docs/contracts/creator-upload-api-contract.md): creator-private な new-package upload の initiation / completion 契約
- [docs/contracts/fan-mvp-common-transport-contract.md](docs/contracts/fan-mvp-common-transport-contract.md): fan MVP read surface 全体で共有する DTO、response envelope、state vocabulary
- [docs/contracts/fan-public-surface-api-contract.md](docs/contracts/fan-public-surface-api-contract.md): `feed / short detail / creator search / creator profile` の read 契約
- [docs/contracts/fan-creator-follow-api-contract.md](docs/contracts/fan-creator-follow-api-contract.md): `creator profile` からの `follow / unfollow` mutation 契約
- [docs/contracts/fan-unlock-main-api-contract.md](docs/contracts/fan-unlock-main-api-contract.md): `unlock / main player` の read 契約
- [docs/contracts/fan-profile-api-contract.md](docs/contracts/fan-profile-api-contract.md): `fan profile private hub` の read 契約
- [docs/contracts/fan-mvp-fixtures-and-integration-guide.md](docs/contracts/fan-mvp-fixtures-and-integration-guide.md): fixture と backend / frontend 接続順の参照ガイド
- [docs/ssot/LOCAL_INDEX.md](docs/ssot/LOCAL_INDEX.md): 外部 repo から取り込んだ SSOT のローカル入口

## ドキュメント運用ルール

- 文書に明記されたものだけを確定仕様として扱うこと
- 未確定のものは、推測で埋めず未確定と明記すること
- `docs/ssot/` 配下は upstream 由来の SSOT として扱い、この repo 側の判断で直接編集しないこと
- `docs/ssot/` の内容を補足・調整したい場合は、`docs/contracts/` など repo 固有の文書で扱うこと
- 文書を追加・更新したら、このファイルと `docs/README.md` の索引も揃えること

## 開発運用ルール

- PR に含めない Codex の一時メモは `./.local/codex-memo/` 配下に置くこと
- `./.local/` は gitignore 対象として扱い、`docs/agent_runs/` や tracked な `.codex/` 配下を外部記憶用途に使わないこと

## 実装前レビュー方針

- 実装に入る前に、対象タスクの目的、非目的、変更対象、参照する契約や文書を短く整理してから着手すること
- 仕様が文書で確定していない箇所は、推測で埋めず未確定として扱うこと
- 変更の中心となる layer boundary を先に確認し、frontend は FSD 境界、backend は package 依存方向と transport/domain 境界を崩さないこと
- 既存コードの責務配置で対応できるならそれを優先し、問題の切り分け前に別手法へ飛びつかないこと

## 差分レビューの優先順位

- 1番目に見るものは仕様適合性であり、ユーザーの実装意図、issue の要求、契約文書から外れていないかを確認すること
- 2番目に見るものは回帰リスクであり、既存導線、state 遷移、認証、権限、例外系、空データ系を壊していないかを確認すること
- 3番目に見るものは設計整合性であり、不要な責務追加、過剰 abstraction、重複、FSD 逸脱、backend の依存逆流がないかを確認すること
- 4番目に見るものは検証妥当性であり、変更箇所に対して十分な test や手元確認があるかを確認すること
- 5番目に見るものは可読性であり、命名、分割、コメント、条件分岐が保守しやすいかを確認すること

## Review Agent 選定方針

- reviewer は既定で `reviewer_simple` を使うこと
- `reviewer` を使うのは、複数の専門観点を並列に当てる合理性がある場合だけに絞ること
- `reviewer_simple` と `reviewer` を両方回すことを既定にしないこと
- 明示的に `reviewer` へ昇格する条件は以下とすること
- `auth / session / permission / access control / payment / unlock` の変更を含む
- `DB schema / SQL / index / migration / cache / concurrency / infra / CI` の変更を含む
- frontend と backend をまたぐ、または複数の layer / bounded context をまたぐ
- shared contract、共通基盤、architecture、FSD 境界、package 境界に影響する
- ユーザーが full review を明示的に要求した
- 上記に当てはまらない通常の leaf 実装や局所修正は `reviewer_simple` を優先すること

## 方針転換ルール

- 初手の実装方針で成立しない兆候が出たら、黙って別手法へ切り替えず、何が成立しなかったかを先に説明すること
- ユーザーの意図から外れる追加変更、ついでの refactor、横断 cleanup は許可なしに行わないこと
- 一発で通すことを優先し、思いつきの試行回数を増やすよりも、差分を小さく保って原因を絞り込むこと

## 実装後の自己レビューと検証

- 実装後は必ず diff を自分で読み、不要変更が混ざっていないことを確認してから検証に進むこと
- 検証は touched area に近いものから順に実施し、必要最小限のコマンドで根拠を作ること
- 検証不足が残る場合は、通ったものだけで安全とみなさず、未確認のリスクを明示すること
- 完了報告では、何を変えたかだけでなく、なぜその実装にしたかも説明すること
