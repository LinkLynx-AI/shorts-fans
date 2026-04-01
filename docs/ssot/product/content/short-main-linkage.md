# short-fans Product SSOT - Short Main Linkage

## 位置づけ

- `short` と `main` の対応関係を定義する
- これは `課金単位`、`投稿フロー`、`分析単位`、`UI` に直結するため、MVP では先に固定する

## 選択肢

### A. `1 short : 1 main`

- 1 本の public short が、1 本の有料 main にだけつながる

### B. `複数 short : 1 main`

- 複数の short が、同じ main に流入する

### C. `m short : n main`

- short と main の関係を柔らかく持つ

## 現時点の推奨

- `MVP` は `複数 short : 1 canonical main` を基本に置く
- ただし `m short : n main` のような柔らかい関係は持たず、`main` を canonical object として固定する

## 推奨理由

### 1. 文脈の連続性を保ったまま複数導入を持てる

- fan にとっては `今見ている short の続きはこの main` と理解できればよく、複数 short があっても `同じ main ID` に飛ぶなら混乱は必須ではない
- 特に、`同じ服 / 同じ画角 / 同じシチュエーション` の複数 short であれば、複数導入でも文脈連続性は保てる

### 2. 課金単位は main に固定できる

- `pay-per-unlock` を `main 単位` で持てば、複数 short が同じ main に流れても課金ロジックはぶれない
- fan から見ても、`この short の続き` が `この canonical main` にまとまるだけである

### 3. 分析はむしろ強くなる

- `どの short が同じ main unlock をどれだけ生んだか` を比較できる
- つまり、`main 単位収益` と `short 単位クリエイティブ学習` を両立できる

### 4. creator の実際の作り方に寄る

- 同じ服 / 同じ画角 / 同じ流れで、音源やダンスだけ違う short を複数撮る前提なら、`1 main : 複数 short` の方が creator 実態に合う
- 1:1 にすると、同じ本編のために `main` を重複させるか、short を 1 本に絞る必要が出て不自然

## 成立条件

- 複数 short が同じ main に紐づく場合でも、次は揃える
  - 同じ creator
  - 同じ canonical main
  - 同じ服 / 画角 / シーン系列
  - 同じ課金対象
- つまり、`別作品の入口` ではなく `同一作品の複数導入` に限定する

## MVP の決め方

- 課金単位は `main 1 本`
- `main` を canonical object にする
- `short` は canonical main に紐づく `公開面 / 流入口` として扱う
- creator 投稿では `main` を 1 本作り、その main に対して複数 short を紐づけられるようにする

## 将来拡張の考え方

- 将来は `short` ごとに、A/B 的な最適化や国別出し分けを持てる
- ただしそのときも、課金単位は `main 単位` を維持する方がよい

## 継続仮説

- `main` を中心に据えて、そこに複数 short を紐づける方が、creator 実態と product learning の両方に合う
- `1 short : 1 main` より、`複数 short : 1 main` の方が acquisition creative を試しやすい
- 一方で `m short : n main` のような柔らかすぎる関係は、MVP では不要

## 未解決論点

- creator に 1 本の `main` へ何本の `short` を紐づけるか
- `同じ main` と見なす continuity 条件をどこまで厳しくするか
- analytics 上で `short 単位最適化` と `main 単位収益` をどう両立させるか

## 参照

- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/content/content-model.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/monetization/billing-and-access.md`
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/product/creator/creator-workflow.md`
