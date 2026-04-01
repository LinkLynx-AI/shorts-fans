# Documentation.md (Status / audit log)

## Current status
- Now: 単一の MVP core domain contract 文書、索引同期、docs-only validation を完了。
- Next: commit、push、PR 作成可否の確認。

## Decisions
- `docs/ssot/` は承認済み設計ソースなので変更しない。
- 実装向け contract は `docs/contracts/mvp-core-domain-contract.md` に置く。
- `post-report` は state と access effect の深さに限定する。
- `creator profile` の最小属性は `display name / avatar / bio` に固定する。
- sandbox 制約で新規 branch ref を作れなかったため、既存の `SHO-10-feat-design_db` を継続利用する。

## How to run / demo
- `docs/contracts/mvp-core-domain-contract.md` を読む。
- その `Canonical Sources` にある SSOT を必要に応じて突き合わせる。
- contract と索引更新だけに絞って `git diff` を確認する。

## Validation results
- `rg -n "## Goals|## Non-goals|## Canonical Sources|## Deferred Decisions" docs/contracts/mvp-core-domain-contract.md`: pass
- `rg -n "## Domain Vocabulary|## Relationship And Ownership Contract|## Access Boundary Contract|## State Transition Contract|## Publish And Unlock Preconditions" docs/contracts/mvp-core-domain-contract.md`: pass
- `git diff --check`: pass
- backend / frontend test suites: skip
- 理由: 今回は docs-only 変更で、実行コードや UI surface を変更していないため
- agent review gate: not run
- 理由: この実行では sub-agent 利用が明示要求されておらず、tool policy 上 `reviewer_simple` / `reviewer_ui_guard` を起動できないため

## Known issues / follow-ups
- sandbox 制約で `.codex/runs/SHO-10/` には書き込めなかった。
- sandbox 制約で `codex/SHO-10-mvp-core-domain-contract` を作れなかったため、既存 issue branch を維持している。
- current branch に upstream がなく、push / PR 作成は別途確認が必要。
