# Media Display Access Contract

## 位置づけ

- この文書は `delivery-ready` 済み asset をどの surface へ、どの access boundary で出すかを固定する実装用契約です。
- 対象は `public short`、`fan main playback`、`creator owner preview` の 3 つに限定します。
- upload acceptance、processing、review queue、billing persistence の契約はこの文書に含めません。

## Goals

- `mvp-media-workflow-contract.md` の `delivery-ready` 後に downstream が期待してよい display/access 境界を固定する。
- public short と private main と owner preview の URL exposure 方針を揃える。
- frontend が URL を opaque asset reference として扱う前提を明記する。

## Non-goals

- raw upload bucket や transcoder job の構成
- signed URL 実装方式や CDN 製品名の固定
- payment 成功時の ledger、unlock 永続化、library 反映
- review status、analytics、売上指標の payload

## Canonical Sources

- `docs/contracts/mvp-media-workflow-contract.md`
- `docs/contracts/fan-mvp-common-transport-contract.md`
- `docs/contracts/fan-unlock-main-api-contract.md`
- `docs/contracts/creator-workspace-owner-preview-api-contract.md`

## Surface Boundary

| surface | caller | access rule | asset shape |
| --- | --- | --- | --- |
| `public short` | anonymous または fan | public に見せてよい short だけを返す | `VideoDisplayAsset` |
| `main playback` | authenticated fan | current session で valid な access entry または owner access が必要 | `VideoDisplayAsset` |
| `owner preview` | approved creator owner | current viewer 自身の short / main だけを preview 可 | `VideoDisplayAsset` |

## Delivery Rules

### Public Short

- public short は fan feed、short detail、creator profile short grid から同じ display boundary で読めることを前提にします。
- short の `url` と `posterUrl` は public asset ref として返し、client は追加署名や path 推測を行いません。

### Main Playback

- main は locked/private delivery を前提にし、current session で検証済みの access entry がある場合だけ `url` と `posterUrl` を返します。
- actual billing と durable purchase 記録の contract は `fan-unlock-main-api-contract.md` が担当し、この文書では purchase 済みまたは owner access 済みの main だけを playback へ出す前提を固定します。
- `access-entry` は durable purchase や owner access を検証した後に temporary grant を発行するだけで、purchase record 自体は作りません。
- main playback grant は library 永続化や purchase ledger と同義にしません。

### Owner Preview

- owner preview は public short でも fan main playback でもなく、creator private workspace の別境界として扱います。
- owner preview で返す short/main は current viewer 自身が owner であることを前提にし、public 公開状態と独立して判定して構いません。

## Asset Rules

- `VideoDisplayAsset.url` と `VideoDisplayAsset.posterUrl` は同じ access boundary から materialize された pair として扱います。
- client は `url`、`posterUrl`、`id` を stable UI input として使いますが、storage key や CDN path として解釈しません。
- `durationSeconds` は delivery-ready 後の確定値を返し、frontend は formatting だけを担当します。
- fan-facing surface は縦型固定のため、layout 計算用寸法は返しません。

## Guardrails

- `raw` bucket の key や internal storage path を transport に出しません。
- URL の lifetime、署名方式、CDN provider は transport contract に含めません。
- `main` の display asset を public short surface に流用しません。
- owner preview を `/api/creator/workspace` overview の中へ混ぜません。
