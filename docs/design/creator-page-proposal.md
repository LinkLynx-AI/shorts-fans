# クリエイターページ構成の調査と提案

## この文書の位置づけ

この文書は確定仕様ではなく、TikTok と OnlyFans を参考にした調査メモと提案です。
このサービスの前提である `下部バーで home / shorts を切り替える構成 + shorts 入口 + クリエイターごとの月額購読 + 限定動画` に合わせて、クリエイターページをどう構成すべきかを整理します。

## 調査サマリ

### TikTok から学ぶべきこと

- TikTok では、プロフィールは視聴フィードから流入した後の受け皿として機能しています。
- TikTok 的なサービスでは、プロフィールは `shorts` 視聴から自然に遷移する画面として成立しています。
- 公式ヘルプでは、プロフィール上に `Liked videos` タブがあり、公開設定に応じてプロフィールから見せられることが確認できます。
- 公式ヘルプでは、`Repost` した動画もプロフィール上の repost タブから見つけられることが確認できます。
- 公式ヘルプでは、クリエイターの `Videos` タブ上部にプレイリストを置けることが確認できます。
- 公式ヘルプでは、Stories は public profiles から視聴できるとされています。
- TikTok の公式発表では、Subscription は月額課金で、限定動画や限定コミュニティを提供する設計です。

ここから言えるのは、TikTok のプロフィールは「フィードから来たユーザーに対して、追加で見たいものをすぐ選ばせる整理されたハブ」であり、情報量はあっても主導線は崩していない、ということです。

### OnlyFans から学ぶべきこと

- OnlyFans はクリエイターごとの購読が中心で、購読するとそのクリエイターの有料コンテンツへアクセスするモデルです。
- 二次情報ベースでは、プロフィール写真やカバー写真は無料で見せつつ、コンテンツ本体は購読後に開放する構成が一般的です。
- 二次情報ベースでは、プロフィールは Instagram に近い見た目を持ちつつ、購読ボタンが価値提案の中心にあります。
- Business Insider では、OnlyFans の弱点として「発見性が弱く、外部流入に頼りやすい」ことが紹介されています。

ここから言えるのは、OnlyFans 的な強さは「誰に課金すれば何が見られるかが明快」な点であり、弱さは「ページが paywall 中心すぎると、発見と回遊が弱くなる」点です。

## 提案の結論

このサービスのクリエイターページは、`UI / UX の骨格はほぼ TikTok`、`限定動画と課金導線の追加部分だけ OnlyFans を参考にする` のが正しいです。

要するに、ページ全体をハイブリッドにするのではなく、土台は TikTok、その上に paywall 機能だけを増築するイメージです。

## 推奨レイアウト

### 1. ヘッダー領域

ファーストビューは TikTok のプロフィール密度に寄せて、次の情報を置きます。

- アイコン
- 表示名
- `@handle`
- 短い紹介文
- 公開 shorts 本数
- 限定動画本数
- 月額価格
- `月額購読` ボタン

ここで重要なのは、見た目を OnlyFans っぽい大きな売り場にしないことです。
TikTok のプロフィールをベースにして、その中に `月額購読` を自然に差し込むのがよいです。

### 2. ヘッダー直下の主要 CTA

ヘッダー直下には主導線を 1 本だけ太く置きます。

- 主 CTA: `このクリエイターを月額購読`
- 補助要素: 月額料金、購読で見られる限定動画数、更新頻度の短い説明

提案としては、まずは TikTok に近い固定されていないプロフィール CTA として始めるのが良いです。
`sticky CTA` は conversion 改善用の後段施策として扱うほうが、今の意図には合います。

### 3. コンテンツ領域のトップレベルタブ

トップレベルタブは 2 つに絞り、見た目も TikTok のタブ切り替えに近づけるのが良いです。

1. `公開shorts`
2. `限定動画`

この分け方を推す理由は、ユーザーが最初に知りたいことが `無料で見られるもの` と `課金すると見られるもの` の差だからです。
ここでタブを増やしすぎると、TikTok 的な軽さが死にます。

### 4. 公開shorts タブ

このタブは TikTok に寄せます。

- デフォルトで最初に開くタブにする
- 3 列グリッドで短尺動画を並べる
- 各カードはサムネイル、再生時間、再生数など最低限の要素だけに絞る
- プレイリストを採用する場合は、このタブ上部に横スクロールで置く

このタブの役割は、`このクリエイターをもっと見る価値があるか` を高速に判断させることです。
一覧は売り場というより、興味を加速させる面として扱うべきです。

### 5. 限定動画 タブ

このタブは見た目は TikTok のまま、意味だけ OnlyFans 的にします。

- 3 列グリッドで、公開shorts と同じ一覧構造を保つ
- 非購読時は全カードをぼかしサムネイルで見せる
- 各カードに `lock`、長さ、短いタイトルか一言を載せる
- タブ上部かタブ直下に `購読するとこの一覧が視聴可能` という短い説明を置く

ここを 2 列など別物の見た目にしすぎると、TikTok ベースの UX から外れます。
まずは同じ 3 列グリッドで統一し、locked 状態の表現だけを追加するほうが筋が良いです。

### 6. 限定動画カード押下時の導線

非購読ユーザーが限定動画カードを押したときは、別ページへ即遷移させるより、まず下から出る導線のほうが良いです。

表示内容の提案は次です。

- クリエイター名
- 月額価格
- 購読で見られる限定動画本数
- 更新頻度や価値を伝える短文
- `月額購読して視聴する` ボタン

これは TikTok のメンバーシップ導線に近い軽さを保ちつつ、OnlyFans 的な支払い理由を明快にするためです。

## 画面構成の提案

モバイルでは、概ね次の順序が良いです。

1. アイコン、表示名、`@handle`
2. 紹介文
3. 本数と価格の要約
4. `月額購読` CTA
5. タブ
6. `公開shorts` または `限定動画` の一覧

## こうしないほうが良いもの

- 公開shorts と限定動画を同じ一覧に混ぜること
- 課金 CTA をページ下部まで隠すこと
- OnlyFans のような時系列投稿一覧や重い売り場 UI をそのまま持ち込むこと
- 限定動画タブでぼかしサムネだけを並べて、価値説明を置かないこと
- TikTok のように見せたいからといって、課金理由を後回しにすること

## この提案の狙い

この構成にすると、`shorts` からの流入は TikTok 型、課金転換は OnlyFans 型、`home` からの探索は発見ハブ型という形にできます。

- フィードから来た人は、TikTok の延長として違和感なくプロフィールを使える
- `home` から来た人も、クリエイターの詳細を見る先として同じプロフィールを使える
- 気になった時点で、同じ画面内に月額購読の理由が見える
- 限定動画タブが `TikTok っぽい見た目のまま、OnlyFans 的な paywall` を担える

## まだ決めていないこと

- フォロー、いいね、コメント、DM をクリエイターページに置くか
- 限定動画カードにタイトルをどこまで出すか
- 月額 CTA を固定バーにするか、ヘッダー常設にするか
- プレイリストを MVP から入れるか
- 更新頻度や投稿本数をどこまで数値で見せるか

## 私の提案

現時点では、次の構成を第一候補にするのが妥当です。

1. ヘッダーに `プロフィール情報 + 月額価格 + 主 CTA`
2. タブは `公開shorts / 限定動画` の 2 つだけ
3. `公開shorts` も `限定動画` も 3 列グリッド
4. 限定動画にはぼかし、lock、短い説明を重ねる
5. 限定動画タップ時は購読モーダルではなく、下から出る軽い導線

この形が、あなたの意図である `UI / UX はほとんど TikTok`、`アダルト機能の追加部分だけ OnlyFans を参考にする` を最も素直に表現できます。

## 参考

- TikTok Help Center: Setting up your profile  
  https://support.tiktok.com/en/getting-started/setting-up-your-profile
- TikTok Help Center: Watch videos in a playlist  
  https://support.tiktok.com/en/using-tiktok/exploring-videos/watch-videos-in-a-playlist?lang=en
- TikTok Help Center: Liking  
  https://support.tiktok.com/en/using-tiktok/exploring-videos/liking
- TikTok Help Center: Repost  
  https://support.tiktok.com/en/article/reposting-videos
- TikTok Help Center: Watching Stories on TikTok  
  https://support.tiktok.com/en/using-tiktok/exploring-videos/watching-stories-on-tiktok
- TikTok Newsroom: Subscription Expansion  
  https://newsroom.tiktok.com/empowering-creators-and-fostering-communities-with-the-expanded-subscription-feature?lang=en
- TikTok Newsroom: Introducing Series  
  https://newsroom.tiktok.com/introducing-a-new-way-for-creators-to-share-premium-content-with-series?lang=en
- DatingScout: OnlyFans Review  
  https://www.datingscout.com/onlyfans/review
- Business Insider: After selling OnlyFans, its founder is launching a rival creator platform  
  https://www.businessinsider.com/onlyfans-cofounder-tim-stokely-launches-rival-creator-platform-subs-2025-5
