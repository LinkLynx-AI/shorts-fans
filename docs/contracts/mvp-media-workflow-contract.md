# MVP Media Workflow Contract

## 位置づけ

- この文書は `SHO-23 MVP media workflow の状態遷移とジョブ境界を定義する` の成果物です。
- `docs/ssot/` と [mvp-core-domain-contract.md](mvp-core-domain-contract.md) を補助し、cycle1 の media workflow を `SHO-11` / `SHO-12` 以降の実装者が追加の product 判断なしで参照できる形に固定します。
- 対象は `main` と `short` の video asset、および creator profile の `avatar` をどこまで同じ基盤で扱い、どこから別境界にするかの整理です。

## Goals

- cycle1 の `import -> processing -> linkage -> review readiness -> publish / unlock eligibility` を固定する。
- `media_asset.processing_state` の state vocabulary を固定する。
- `submission package ready`、`main unlockable`、`short public publishable` の条件を固定する。
- `CloudFront` を workflow の必須条件にせず、delivery-ready 抽象で語れるようにする。
- creator `avatar` image を動画 workflow に混ぜない方針を固定する。

## Non-goals

- raw source timeline や project object の導入
- app 内 short cutout editor や source-derived cutout job の設計
- worker、queue、API shape、DB column の詳細設計
- 本番 multi-account、監視、通知、CDN 最適化の完成
- avatar moderation UI や profile editing UX の詳細

## Canonical Sources

- `docs/contracts/mvp-core-domain-contract.md`
- `docs/ssot/product/content/content-model.md`
- `docs/ssot/product/content/short-main-linkage.md`
- `docs/ssot/product/creator/creator-workflow.md`
- `docs/ssot/product/creator/capture-and-audio-model.md`
- `docs/ssot/product/moderation/moderation-and-review.md`
- `docs/ssot/product/monetization/billing-and-access.md`
- `docs/TECH_STACK.md`

## Cycle1 Decisions

### Asset model

- cycle1 は `import-first` を維持し、`main` と `short` は別々の deliverable video asset として import します。
- `media_asset` は raw source や edit timeline ではなく、creator が持ち込む完成済み asset を表します。
- issue 文面の `short cutout` は、cycle1 では platform 内部で short を生成する job ではなく、short asset が review-ready になるまでの handoff boundary として扱います。
- `submission package` は次で構成します。
  - 1 本の canonical `main`
  - 1 本以上の linked `short`
  - continuity metadata
  - creator / consent / ownership 情報

### Asset taxonomy

- `main video asset`
  - `main` が参照する paid continuation 用 asset
- `short video asset`
  - `short` が参照する public surface 用 asset
- `creator avatar profile asset`
  - `creator_profile_drafts.avatar_url` / `creator_profiles.avatar_url` に反映する image asset
- `avatar` は `submission package` に含めず、`short/main` の video workflow には載せません。

## Media Asset State Contract

### `media_asset.processing_state`

| State | Meaning | Required / guaranteed fields | Primary transitions |
| --- | --- | --- | --- |
| `uploaded` | object storage 受理済み。まだ配信用には使わない | `storage_provider` / `storage_bucket` / `storage_key` | `uploaded -> processing`, `uploaded -> failed` |
| `processing` | transcode / normalize / delivery materialization 中 | storage ref は必須。`playback_url` は未保証 | `processing -> ready`, `processing -> failed` |
| `ready` | playback/render に使える delivery ref が揃った状態 | video は `playback_url` 必須。duration はこの状態で確定させる | terminal for cycle1 processing |
| `failed` | 自動処理を止めた状態 | last known storage ref は保持してよい。delivery ref は未保証 | manual requeue のみで `failed -> processing` |

追加ルール:

- `ready` 以外の state は fan surface / creator publishability 判定に使えません。
- `ready` は `review approved` を意味しません。media processing 完了だけを表します。
- 同じ imported file を再処理するときは同じ asset identity で `failed -> processing` を許可します。
- 別ファイルへ差し替えるときは新しい asset として re-import します。

## Handoff Boundary Contract

| Boundary | Input | Output | Downstream が進める条件 |
| --- | --- | --- | --- |
| `upload acceptance` | creator が import した file | `media_asset` in `uploaded` | object storage ref が保存済み |
| `processing` | `uploaded` asset | `media_asset` in `processing` or `failed` | job が asset を占有し、delivery materialization を開始済み |
| `delivery-ready` | `processing` asset | `media_asset` in `ready` | playback/render に必要な delivery ref が利用可能 |
| `linkage validation` | `ready` な main / short asset | canonical main と short linkage が成立した object graph | 同一 creator、同一 purchase target、continuity metadata が揃う |
| `submission package ready` | linkage 済み object graph | review intake 可能な package | `1 main + 1本以上の short` がすべて `ready` |
| `publish / unlock eligibility` | review 済み object | `main unlockable` / `short public publishable` | media readiness と review/access 条件を両方満たす |

## Eligibility Contract

### `submission package ready`

- 次をすべて満たした状態を指します。
  - canonical `main` が存在する
  - linked `short` が 1 本以上ある
  - package 内の video asset がすべて `ready`
  - owner creator が同一
  - purchase target が同一
  - 同一作品として読める continuity metadata が揃う

### `main unlockable`

- 次をすべて満たした状態を指します。
  - `main.media_asset` が `ready`
  - creator capability が `approved`
  - `main.state = approved_for_unlock`
  - `ownership_confirmed = true`
  - `consent_confirmed = true`
  - blocking な `post_report_state` がない

### `short public publishable`

- 次をすべて満たした状態を指します。
  - `short.media_asset` が `ready`
  - `short.state = approved_for_publish`
  - canonical `main` が `main unlockable` を満たす
  - blocking な `post_report_state` がない
- したがって `short` 単体が `ready` でも、canonical `main` が unlockable でなければ public に出しません。

## Delivery Contract

### `delivery-ready`

- cycle1 では CDN 製品名ではなく、現在の access boundary に従って asset を返せることだけを保証します。
- `CloudFront` は infra 上の選択肢ですが、workflow 契約の必須条件にはしません。

### public short delivery

- `short` は public surface に出るため、fan がそのまま再生・表示できる delivery ref を持つ必要があります。
- cycle1 dev では direct S3 public object や軽量 CDN でも構いません。
- 将来 `CloudFront` を使う場合も domain 側は `delivery-ready` のまま扱います。

### paid main delivery

- `main` は locked/private delivery を前提にし、signed URL や access-checked origin を許容します。
- cycle1 dev では direct private S3 + signed URL でも成立可能です。
- `main unlockable` は `CloudFront` 導入の有無ではなく、unlock 後に安全に playback できるかで判定します。

## Creator Avatar Contract

### 境界

- `avatar` は `creator profile` の image asset であり、`short/main` と同じ video workflow に統合しません。
- 永続化先は `creator_profile_drafts.avatar_url` と `creator_profiles.avatar_url` を維持します。
- 同じ S3 / image delivery 基盤を再利用してよいですが、`media_assets` の state machine には載せません。

### cycle1 flow

- cycle1 の avatar flow は次に留めます。
  - image upload
  - basic validation
  - stable image URL の反映
  - draft/profile の `avatar_url` 更新
- `avatar` は `submission package ready`、`main unlockable`、`short public publishable` の条件に含めません。
- `avatar` の差し替えは profile 更新として扱い、動画 asset の retry / recovery と混ぜません。

## Failure, Retry, And Manual Recovery

- 自動 retry は一時的な infra/media-processing failure に限定します。
  - object storage finalization failure
  - transcode / normalize failure
  - delivery materialization failure
- cycle1 の自動 retry 回数は合計 `3` 回までです。
- retry 枯渇時は `failed` に遷移し、以降は manual recovery に切り替えます。
- 次は自動 retry しません。
  - continuity mismatch
  - canonical main linkage mismatch
  - review rejection / revision requested
  - ownership / consent 不足
  - policy / moderation 起因の block
- manual recovery の最小方針は次です。
  - 同じ file を再処理するだけなら same asset identity を requeue
  - file 差し替えが必要なら新しい asset を re-import
  - linkage や metadata の不整合は creator/operator が修正してから再度 review intake する

## Downstream Guidance

- `SHO-11` / `SHO-12` では、少なくとも次を表現できる schema/query を前提にします。
  - `media_asset.processing_state = uploaded | processing | ready | failed`
  - `short` と `main` は別 asset import
  - `avatar` は profile image であり、video workflow の asset state とは別境界
  - `short public publishable` は canonical `main` の unlockability に依存する
