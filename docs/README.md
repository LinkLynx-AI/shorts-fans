# ドキュメント索引

このディレクトリには、プロダクト定義、契約、実装ルールを整理する文書を置きます。

プロダクト仕様と事業仕様の canonical は `docs/ssot/` 配下とし、`docs/*.md` はその前提と矛盾しない実装ルール・技術前提・構成方針を扱います。

## 現在ある文書

- [BACKEND_STRUCTURE.md](BACKEND_STRUCTURE.md): backend のディレクトリ構成方針、依存方向、SQL と transport の配置ルール
- [TECH_STACK.md](TECH_STACK.md): Frontend / Backend / Infra / Payment / Analytics を含む技術選定
- [GO.md](GO.md): Go 実装ルール、package 設計、並行処理、テスト、依存管理の運用ルール
- [TYPESCRIPT.md](TYPESCRIPT.md): TypeScript フロントエンド実装ルールと FSD 運用ルール
- [infra/dev-media-sandbox.md](infra/dev-media-sandbox.md): dev 用 AWS media sandbox の Terraform 構成、guardrail、適用手順、backend 接続用 env 対応
- [infra/dev-media-smoke.md](infra/dev-media-smoke.md): dev media smoke の再現手順、切り分け、manual recovery、quota/cost 注意点
- [contracts/mvp-core-domain-contract.md](contracts/mvp-core-domain-contract.md): MVP core 永続化タスク向けの実装用ドメイン契約
- [contracts/mvp-media-workflow-contract.md](contracts/mvp-media-workflow-contract.md): MVP media workflow の状態遷移、delivery 境界、avatar 境界を整理した実装用契約
- [contracts/media-display-access-contract.md](contracts/media-display-access-contract.md): delivery-ready asset を public short / main playback / owner preview に出す display/access 境界契約
- [contracts/mvp-media-display-fixtures-and-verification-guide.md](contracts/mvp-media-display-fixtures-and-verification-guide.md): upload complete から public short / main playback / owner preview までの fixture と verification 入口ガイド
- [contracts/fan-auth-api-contract.md](contracts/fan-auth-api-contract.md): Cognito 前提の fan sign in / sign up / password reset / logout / re-auth transport 契約
- [contracts/fan-auth-modal-ui-contract.md](contracts/fan-auth-modal-ui-contract.md): fan auth custom modal UI の state / recovery 契約
- [contracts/viewer-bootstrap-api-contract.md](contracts/viewer-bootstrap-api-contract.md): app shell が読む current viewer bootstrap の read 契約
- [contracts/viewer-profile-api-contract.md](contracts/viewer-profile-api-contract.md): sign-up / fan settings / creator workspace で共有する viewer profile の transport 契約
- [contracts/viewer-creator-entry-api-contract.md](contracts/viewer-creator-entry-api-contract.md): fan profile から始める creator registration / active mode switch の transport 契約
- [contracts/creator-workspace-api-contract.md](contracts/creator-workspace-api-contract.md): `/creator` workspace の creator info / overview / revision summary 契約
- [contracts/creator-workspace-owner-preview-api-contract.md](contracts/creator-workspace-owner-preview-api-contract.md): creator owner 向け short/main preview list/detail と short caption 更新契約
- [contracts/creator-workspace-main-price-api-contract.md](contracts/creator-workspace-main-price-api-contract.md): creator owner が本編価格を変更する private mutation 契約
- [contracts/creator-upload-api-contract.md](contracts/creator-upload-api-contract.md): creator-private な new-package upload の initiation / completion 契約
- [contracts/fan-mvp-common-transport-contract.md](contracts/fan-mvp-common-transport-contract.md): fan MVP read surface 全体で共有する DTO、response envelope、state vocabulary
- [contracts/fan-public-surface-api-contract.md](contracts/fan-public-surface-api-contract.md): `feed / short detail / creator search / creator profile` の read 契約
- [contracts/fan-short-pin-api-contract.md](contracts/fan-short-pin-api-contract.md): `feed` からの `pin / unpin` mutation 契約
- [contracts/fan-creator-follow-api-contract.md](contracts/fan-creator-follow-api-contract.md): `creator profile` からの `follow / unfollow` mutation 契約
- [contracts/fan-unlock-main-api-contract.md](contracts/fan-unlock-main-api-contract.md): `unlock / main player` の read 契約
- [contracts/fan-profile-api-contract.md](contracts/fan-profile-api-contract.md): `fan profile private hub` の read 契約
- [contracts/fan-mvp-fixtures-and-integration-guide.md](contracts/fan-mvp-fixtures-and-integration-guide.md): fixture と backend / frontend 接続順の参照ガイド
- [ssot/LOCAL_INDEX.md](ssot/LOCAL_INDEX.md): 外部 repo から取り込んだ SSOT のローカル入口

## 補助ディレクトリ

- `contracts/`: SSOT を置き換えずに、実装者向けの契約や境界責務をまとめる
- `contracts/fixtures/`: fan / creator-private 契約の representative JSON fixture を置く
- `.local/`: Codex の PR 単位一時メモを置く gitignored ディレクトリ。`./.local/codex-memo/` を使い、PR には含めない

## 今後追加する文書領域

- 設計書: 画面構成、主要導線、ユースケース、権限モデル
- 契約: API 契約、ドメイン契約、境界責務
- 実装ルール: コーディングルール、レイヤールール、命名ルール、運用ルール

今後の文書追加は、目的とスコープが明確になってから行います。
