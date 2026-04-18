# Submission Package Review Contract

## 位置づけ

- この文書は `SHO-193 submission package review の契約と状態境界を更新する` の成果物です。
- `upload complete`、`submission package ready`、`review submit`、object-level decision state、manual-first review provenance の境界を固定します。
- `docs/ssot/`、[mvp-core-domain-contract.md](mvp-core-domain-contract.md)、[mvp-media-workflow-contract.md](mvp-media-workflow-contract.md) を補助し、後続の submit / decision 実装が追加の product 判断なしで進められる状態にします。

## Goals

- `upload package` と `submission package` を分離し、`upload complete` が review ready や publish / unlock eligibility を意味しないことを固定する。
- `submission package ready` を review intake 可能条件として固定し、`review submit` を package-level action として明示する。
- review decision は package ではなく `main` と各 `short` に保持する前提を固定する。
- MVP の正式経路が manual review であり、将来 auto review を追加しても manual fallback / override を残す前提を固定する。
- reason code、decision source、decision timestamp を保持できる provenance 前提を固定する。

## Non-goals

- `review submit` / `decision` endpoint の request / response shape
- admin review queue / detail / decision UI の payload
- DB schema、table、column、index、migration の具体設計
- reviewer assignment、SLA、notification channel の運用設計
- auto review provider、model、confidence threshold の決定

## Canonical Sources

- `docs/ssot/product/moderation/moderation-and-review.md`
- `docs/ssot/product/creator/creator-workflow.md`
- `docs/contracts/mvp-core-domain-contract.md`
- `docs/contracts/mvp-media-workflow-contract.md`
- `docs/contracts/creator-upload-api-contract.md`
- `docs/contracts/media-display-access-contract.md`

## Boundary Model

### `upload package`

- `upload package` は raw upload を束ねる creator-private grouping です。
- `upload package` は review intake 単位ではありません。
- `upload complete` で作られるのは draft `main` / draft `shorts` と `uploaded` state の `media_asset` であり、review 状態や公開可否ではありません。

### `submission package`

- `submission package` は review intake 単位です。
- v1 の durable anchor は canonical `main` とします。
- 構成要素は次のとおりです。
  - 1 本の canonical `main`
  - 1 本以上の linked `short`
  - continuity metadata
  - review 判断に必要な creator / ownership / consent 情報
- `submission package` は consumer 向けの独立 object ではなく、review intake boundary です。
- `review submit` / `resubmit` の target は、canonical `main` を anchor に submit 時点の linked `short` set と metadata を解決した current package snapshot です。
- `packageToken` は upload completion のための一時 token であり、durable な `submission package` identity と同一視してはいけません。

### `upload complete`

- `upload complete` の成功は、draft content persistence が完了したことだけを意味します。
- `upload complete` は次を意味しません。
  - `submission package ready`
  - `review submit` 済み
  - `pending review`
  - `approved for publish`
  - `approved for unlock`
  - `short public publishable`
  - `main unlockable`

### `submission package ready`

- `submission package ready` は review submit 前提を満たした package readiness です。
- 次をすべて満たしたときだけ成立します。
  - canonical `main` が存在する
  - linked `short` が 1 本以上ある
  - package 内の video asset がすべて `ready`
  - owner creator が同一
  - purchase target が同一
  - 同一作品として読める continuity metadata が揃う
  - review intake に必要な ownership / consent input が欠けていない
- `submission package ready` は readiness predicate であり、review state ではありません。
- `submission package ready` だけでは publish / unlock eligibility に進みません。

### `review submit`

- `review submit` は `submission package` 単位の intake action です。
- caller は authenticated viewer であり、かつ approved creator capability を持つ package owner 自身に限ります。
- creator は `submission package ready` を満たした package だけを submit / resubmit できます。
- successful `review submit` は review cycle の開始を意味しますが、承認を意味しません。
- MVP では `main` だけ、または一部 `short` だけを切り出した独立 submit path は canonical にしません。submit / resubmit action は package 単位で扱います。

## Decision State Contract

### Package-level と object-level の分離

- reviewer は package 全体を見て判断してよいですが、effective decision は object ごとに保持します。
- package-level publish state や package-level approval state は canonical にしません。
- したがって、少なくとも次のような混在状態を表現できる必要があります。
  - main `pending review`
  - short A `approved for publish`
  - short B `revision requested`

### `main`

| state | meaning | review cycle rule |
| --- | --- | --- |
| `draft` | review submit 前 | ready package の `review submit` で `pending review` に入れる |
| `pending review` | main review 待ち | `approved for unlock` / `revision requested` / `rejected` に遷移できる |
| `approved for unlock` | paid continuation として承認済み | `main unlockable` の review 条件を満たす |
| `revision requested` | 修正して再 submit が必要 | ready package の再 submit で再度 `pending review` に戻せる |
| `rejected` | 現在の形では unlock 対象不可 | 許可された revision flow 後のみ再度 `pending review` に戻せる |

### `short`

| state | meaning | review cycle rule |
| --- | --- | --- |
| `draft` | review submit 前 | ready package の `review submit` で `pending review` に入れる |
| `pending review` | short review 待ち | `approved for publish` / `revision requested` / `rejected` に遷移できる |
| `approved for publish` | public surface 用に承認済み | `short public publishable` の review 条件を満たす |
| `revision requested` | 修正して再 submit が必要 | ready package の再 submit で再度 `pending review` に戻せる |
| `rejected` | 現在の形では publish 不可 | 許可された revision flow 後のみ再度 `pending review` に戻せる |

### Re-submit policy

- resubmit action も package 単位です。
- package 単位の resubmit でも、package 内の全 object を機械的に `pending review` へ戻してはいけません。
- `revision requested` は creator に修正と再 submit を要求する state です。
- `rejected` は current form では通せない state であり、自動 reopen しません。
- `pending review` に再投入するのは、少なくとも次のいずれかに当てはまる object だけです。
  - `draft`
  - `revision requested`
  - 許可された revision flow に入った `rejected`
  - 前回の effective decision 後に新規追加、差し替え、または判断根拠が変わる編集が入った object
- 既に `approved for publish` / `approved for unlock` で、かつ判断根拠が変わっていない object は、その approval を維持します。
- approved 済み object を再 review に戻す必要がある場合は、operator / policy 起点の明示的 reopen として扱います。
- resubmit eligibility の細かな operator policy や transport はこの contract では固定しませんが、object-level decision を package-level state に潰してはいけません。

## Gate Relation Contract

### `main unlockable`

- `main unlockable` は review 完了後の access gate です。
- 次をすべて満たしたときだけ成立します。
  - `main.media_asset` が `ready`
  - creator capability が `approved`
  - `main.state = approved_for_unlock`
  - `ownership_confirmed = true`
  - `consent_confirmed = true`
  - blocking な `post_report_state` がない

### `short public publishable`

- `short public publishable` は review 完了後の public gate です。
- 次をすべて満たしたときだけ成立します。
  - `short.media_asset` が `ready`
  - `short.state = approved_for_publish`
  - canonical `main` が `main unlockable` を満たす
  - blocking な `post_report_state` がない
- したがって `short` 単体が approved でも、linked `main` が unlockable でなければ public に出しません。

### Owner preview

- owner preview は `submission package ready`、`review submit`、public publish / unlock gate と独立した private boundary です。
- owner preview は review 承認の代替ではなく、creator owner が自分の delivery-ready asset を preview / QA する経路です。
- owner preview access を purchase や publish approval と混ぜてはいけません。

## Manual-first Review And Provenance

### MVP の正式経路

- MVP の正式運用は manual review です。
- `short` の pre-publish review と `main` の pre-unlock review は manual path を維持します。
- repeat creator や将来 automation を導入しても、manual review 窓口を削除しません。

### Future automation compatibility

- 将来 auto review を追加しても、少なくとも次を残せる必要があります。
  - manual review fallback
  - manual final decision path
  - manual override
- auto review は queue prioritization、assist signal、または automated decision に使えても、manual path を置き換える前提にしません。

### Decision provenance

- effective decision は object ごとに、少なくとも次を保持できる必要があります。
  - `reason code`
  - `decision source`
  - `decision timestamp`
- `decision source` の canonical vocabulary は次とします。
  - `manual`: human reviewer が effective decision を行った、または確定した
  - `auto`: automated system が effective decision を行い、後続 override がない
  - `manual_override`: automated output の後に human reviewer が effective decision を上書きした
- MVP の expected source は `manual` です。
- provider 名、model 名、confidence、reviewer assignment はこの contract では固定しません。

## Downstream Guidance

- 後続の submit / decision transport は、この文書を基準に `upload complete` と `review submit` を分離すること。
- review intake が package 単位でも、decision state と approval gate は object-level に保つこと。
- public short / main unlock / owner preview の access boundary は、この文書単体で再定義せず既存 access contract と整合させること。
