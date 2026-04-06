# MVP Core Domain Contract

## 位置づけ

- この文書は、MVP core 永続化タスク向けの実装用ドメイン契約です。
- `docs/ssot/` は CEO 承認済みの product source of truth として維持し、この文書で置き換えません。
- この文書は、既存 SSOT にある意思決定を `SHO-11` 以降の実装者が追加の product 判断なしで使えるように、一箇所へ束ね直すことだけを目的にします。

## Goals

- `user / creator capability / creator profile / short / main / unlock / submission package / review` の境界と用語を固定する。
- DB 設計に必要な主要状態遷移と `publish / unlock` 前提条件を固定する。
- SSOT から読める内容のうち、実装上の解釈がぶれやすい箇所を明示する。

## Non-goals

- DB schema、table、column、index、FK の具体設計
- SQL query や `sqlc` 契約の設計
- API / handler 契約の設計
- review tooling UI の詳細
- `subscription` や hybrid 課金の導入判断
- analytics event の詳細設計

## Canonical Sources

- `docs/ssot/product/account/identity-and-mode-model.md`
- `docs/ssot/product/account/account-permissions.md`
- `docs/ssot/product/content/content-model.md`
- `docs/ssot/product/content/short-main-linkage.md`
- `docs/ssot/product/monetization/billing-and-access.md`
- `docs/ssot/product/moderation/moderation-and-review.md`
- `docs/ssot/product/creator/creator-onboarding-surface.md`
- `docs/ssot/product/creator/creator-workflow.md`
- `docs/ssot/product/fan/consumer-state-and-profile.md`
- `docs/ssot/product/scope/mvp-boundaries.md`

## Domain Vocabulary

### user

- `user` は root login identity です。
- 認証、支払い設定、年齢関連確認、support 履歴、mode switching の基点はすべて `user` に置きます。
- `user` は consumer-side state のみを持つことも、consumer-side state に加えて creator capability を持つこともあります。

### fan state

- `fan state` は同一 `user` 配下にある consumer-side state です。
- ここには `follow`、`unlock / purchase`、`viewing history`、`pin`、fan private hub の状態を含めます。
- `fan state` は別 login identity ではありません。

### creator capability

- `creator capability` は、`user` が creator private surface を扱うための approval-based capability です。
- これにより creator onboarding 結果、creator workspace、content ownership、submission、review 状態、creator analytics へのアクセス権が付与されます。
- public profile の見え方そのものとは別境界として扱います。

### creator profile

- `creator profile` は fan が閲覧する public creator surface です。
- MVP contract 上、この profile に固定する最小属性は次の 3 つです。
  - `display name`
  - `avatar`
  - `bio`
- `avatar` は public surface 上の profile image slot を指します。creator が custom avatar を未設定でも、この slot 自体は維持され、read transport では `null` を返して client が platform default avatar を描画してよいものとします。
- public handle、URL、その他 discoverability identifier はこの contract の対象外です。

### short

- `short` は public な non-explicit surface です。
- creator-owned な縦型動画であり、`discovery`、`continuation hook`、canonical `main` への流入口を担います。
- MVP では、1 本の `short` は必ず 1 本の canonical `main` にだけ紐づきます。

### main

- `main` は `short` と同じ文脈から始まる paid continuation です。
- MVP では canonical paid object として扱います。
- 形式は縦型動画のみに固定します。

### main unlock

- `main unlock` は 1 本の canonical `main` に対する purchase / access state です。
- `short` にぶら下げません。
- creator subscription entitlement としても扱いません。
- 機能上の状態は `not purchased` または `purchased` のみです。

### submission package

- `submission package` は creator 側の review intake 単位です。
- 構成要素は次のとおりです。
  - 1 本の canonical `main`
  - 1 本以上の linked `short`
  - continuity metadata
  - 必要な creator / consent / ownership 情報
- consumer 向けに公開される独立 object ではなく、あくまで intake boundary です。

### review state

- `review state` は `creator capability`、`main`、各 `short` に分けて保持します。
- reviewer は package 全体を見て判断してよいですが、decision state は object ごとに残します。

### post-report state

- `post-report state` は report や再審査後の制限状態を表します。
- この contract では state 名と access effect のみを固定します。
- report intake channel、queue、operator workflow の詳細は SSOT 側の moderation 文書に残します。

## Relationship And Ownership Contract

### identity と capability

- 1 つの `user` は 1 つの root identity を持ちます。
- `user` は常に `fan state` を持ちえます。
- `user` は追加で `creator capability` を持ちえます。
- `creator capability` は creator approval によって解放されるものであり、public profile の存在と同一視しません。

### public profile 境界

- `creator profile` は `creator capability` を持つ同じ `user` に属します。
- `creator profile` は private draft と public surface の両状態を持ちえます。
- upload、review、analytics を含む creator private workspace は `creator profile` と分離します。
- creator approval 前は private な `creator profile draft` のみを持てます。public creator profile は持ちません。

### content ownership

- `short` は submission した creator-capable `user` が所有します。
- `main` も submission した creator-capable `user` が所有します。
- ownership と consumer-side purchase は別概念です。
- creator が自分の `main` を見るときは purchase ではなく ownership / preview access を使います。

### short-main linkage

- MVP は `1 canonical main : 複数 short` を基本モデルにします。
- 1 本の `short` は 1 本の canonical `main` に紐づきます。
- 1 本の `main` は 1 本以上の `short` を持てます。
- 同じ `main` に紐づく `short` は、別作品の入口ではなく同一作品の複数導入である必要があります。
- この contract で最低限固定する continuity 条件は次のとおりです。
  - 同じ creator owner
  - 同じ canonical main target
  - 同じ purchase target
  - 同一作品として読める scene / outfit / angle / session lineage

### unlock 境界

- `main unlock` は canonical `main` に紐づきます。
- 複数の `short` が同じ `main` に流れていても、purchase state はその canonical `main` に対して 1 つで足ります。
- 将来 `subscription` を併用するときの整合境界は、この contract では固定しません。

### submission と review 境界

- creator-side の編集単位と review intake 単位は `submission package` です。
- ただし decision state は package-level publish state にまとめません。
- 後続実装では、少なくとも次のような組み合わせを表現できる必要があります。
  - creator approved
  - main pending review
  - short A approved for publish
  - short B revision requested

## Access Boundary Contract

### public

- `short feed`
- `short player`
- `creator profile`
- 承認済みで、現在 publish 可能な `short`

### locked

- access 条件を満たしていない `main`
- public `short` の先にある paid continuation

### purchased

- purchase access を持つ `user` は、対応する canonical `main` を視聴できます。
- purchase state は canonical `main` 単位です。

### owned

- creator owner は purchase なしで自分の `main` を preview / QA できます。
- ownership を purchase として表現してはいけません。

### limited

- post-report limitation によって、以前は利用可能だった `short` や `main` が制限されることがあります。
- `temporarily limited` と `removed` は、approval や purchase より後段で access を止めうる状態です。

## State Transition Contract

### Creator capability state

| State | Meaning | Primary transitions |
| --- | --- | --- |
| `draft` | onboarding 未 submit | `draft -> submitted` |
| `submitted` | creator review 待ち | `submitted -> approved`, `submitted -> rejected` |
| `approved` | creator capability granted | `approved -> suspended` |
| `rejected` | creator capability denied | eligible な resubmit に限り `rejected -> submitted` |
| `suspended` | approval 後に強制制限された状態 | 回復経路はこの contract の対象外 |

追加 metadata 契約:

- `rejected` には `re-submit eligible`、`support review required`、`reason code` を持たせます。
- self-serve resubmit は、eligible な fixable reject にだけ許可します。
- 同じ onboarding case の self-serve resubmit は最大 `2` 回までとします。

### Short state

| State | Meaning | Primary transitions |
| --- | --- | --- |
| `draft` | review submit 前 | `draft -> pending review` |
| `pending review` | short review 待ち | `pending review -> approved for publish`, `pending review -> revision requested`, `pending review -> rejected` |
| `approved for publish` | public surface 用に承認済み | 後から post-report action で `removed` になりうる |
| `revision requested` | 修正して再 submit が必要 | `revision requested -> pending review` |
| `rejected` | 現在の形では publish 不可 | 許可された revision flow 後のみ再度 `pending review` に戻れる |
| `removed` | public surface から外された状態 | この contract では終端状態 |

### Main state

| State | Meaning | Primary transitions |
| --- | --- | --- |
| `draft` | review submit 前 | `draft -> pending review` |
| `pending review` | main review 待ち | `pending review -> approved for unlock`, `pending review -> revision requested`, `pending review -> rejected` |
| `approved for unlock` | paid continuation として承認済み | 後から `suspended` または `removed` になりうる |
| `revision requested` | 修正して再 submit が必要 | `revision requested -> pending review` |
| `rejected` | 現在の形では unlock 対象不可 | 許可された revision flow 後のみ再度 `pending review` に戻れる |
| `suspended` | policy / safety で一時利用不可 | 回復経路はこの contract の対象外 |
| `removed` | takedown 済み | この contract では終端状態 |

### Main unlock access state

| State | Meaning | Primary transitions |
| --- | --- | --- |
| `not purchased` | purchase-based access なし | `not purchased -> purchased` |
| `purchased` | canonical main に対する paid access あり | この contract では逆遷移を定義しない |

### Post-report state

| State | Meaning | Access effect |
| --- | --- | --- |
| `under review` | report 後の再評価中 | 表示継続か一時制限かは operator action に依存するが、state としては保持必須 |
| `temporarily limited` | 一時制限中 | public visibility や main playback が止まりうる |
| `removed` | takedown 中 | 通常 consumer surface では利用不可 |
| `restored` | 制限解除済み | report 以外の最新 approval state が許す access に戻る |

## Publish And Unlock Preconditions

### Creator precondition

- `creator approved` は、`short` と `main` の publish / unlock の前提です。
- approval 前は private な `creator profile` draft data のみ持てます。
- approval 前の `user` は public creator profile、publish 用 content upload、active な creator workspace を持てません。

### Short public precondition

`short` を public surface に出してよいのは、次の条件をすべて満たすときだけです。

- creator state が `approved`
- short state が `approved for publish`
- linked `main` state が `approved for unlock`
- その `short` が public non-explicit safety expectation を満たす
- access-limiting な post-report state ではない

linked `main` approval を必須にするのは、`short` がその `main` への直接導線だからです。

### Main unlock precondition

`main` を unlock 対象にしてよいのは、次の条件をすべて満たすときだけです。

- creator state が `approved`
- main state が `approved for unlock`
- 必要な ownership / consent 情報が review を通過している
- access-limiting な post-report state ではない

### Purchased playback precondition

- purchased playback には `main unlock = purchased` が必要です。
- creator owner による playback は ownership / preview access であり、purchase は不要です。
- ただし purchase 済みでも、後から `temporarily limited` や `removed` になれば playback は止まりえます。

## Implementation Notes For SHO-11 And Later

- `user` を root identity として扱い、creator と fan を別 login entity に分けないこと。
- `creator capability` と `creator profile` を別境界として持つこと。
- purchase access と creator ownership access を混ぜないこと。
- review intake は package 単位でも、decision state は `creator`、`main`、`short` ごとに分けて持つこと。
- 後続実装では少なくとも次の問いに答えられる必要があります。
  - どの `short` がどの canonical `main` に紐づくか
  - どの `main` をどの `user` が purchase 済みか
  - どの `main` を creator が owner として保持しているか
  - publish block が creator review、content review、post-report limitation のどこから来ているか

## SSOT Alignment And Delta Notes

- SSOT では `creator account` と `fan account` という表現が出てきますが、実装上は `1 user identity + optional creator capability + active mode switch` と解釈します。別 login identity にはしません。
- SSOT は `creator profile` を public surface、creator workspace を private surface としています。この contract ではその境界を固定し、public profile の最小属性を `display name / avatar / bio` に限定します。
- SSOT は `submission package` を review intake としています。この contract では、`creator`、`main`、`short` の object-level review state を後続永続化で扱えるよう、package-level publish state は導入しません。
- SSOT は `post-report` を運用面まで含めて扱っています。この contract は state 名と access effect にだけ絞ります。

## Deferred Decisions

- DB schema shape、naming、index、FK policy
- API wire contract と DTO 名
- creator profile の public handle / URL 契約
- `subscription` と `main unlock` の整合設計
- report triage と moderation tooling の operator workflow
- analytics attribution window と event 定義
