# ドキュメント索引

このディレクトリには、プロダクト定義、契約、実装ルールを整理する文書を置きます。

プロダクト仕様と事業仕様の canonical は `docs/ssot/` 配下とし、`docs/*.md` はその前提と矛盾しない実装ルール・技術前提・構成方針を扱います。

## 現在ある文書

- [BACKEND_STRUCTURE.md](BACKEND_STRUCTURE.md): backend のディレクトリ構成方針、依存方向、SQL と transport の配置ルール
- [TECH_STACK.md](TECH_STACK.md): Frontend / Backend / Infra / Payment / Analytics を含む技術選定
- [GO.md](GO.md): Go 実装ルール、package 設計、並行処理、テスト、依存管理の運用ルール
- [TYPESCRIPT.md](TYPESCRIPT.md): TypeScript フロントエンド実装ルールと FSD 運用ルール
- [contracts/mvp-core-domain-contract.md](contracts/mvp-core-domain-contract.md): MVP core 永続化タスク向けの実装用ドメイン契約
- [ssot/LOCAL_INDEX.md](ssot/LOCAL_INDEX.md): 外部 repo から取り込んだ SSOT のローカル入口

## 補助ディレクトリ

- `contracts/`: SSOT を置き換えずに、実装者向けの契約や境界責務をまとめる
- `.local/`: Codex の PR 単位一時メモを置く gitignored ディレクトリ。`./.local/codex-memo/` を使い、PR には含めない

## 今後追加する文書領域

- 設計書: 画面構成、主要導線、ユースケース、権限モデル
- 契約: API 契約、ドメイン契約、境界責務
- 実装ルール: コーディングルール、レイヤールール、命名ルール、運用ルール

今後の文書追加は、目的とスコープが明確になってから行います。
