# Go 実装ルール

## 0. 位置づけ
- この文書は Go 公式資料をベースに、本リポジトリで Go コードを書くときの運用ルールを定めるものです。
- 目的は、可読性・保守性・API の明確さ・並行処理安全性・テスト容易性を継続的に守ることです。
- ここに書いていない細部は Go の標準ツールと公式ドキュメントの流儀を優先します。

## 1. 優先順位
1. `go test` / コンパイラが要求する正しさ
2. `gofmt`
3. `go vet`
4. `govulncheck`
5. この文書

## 2. 非交渉ルール
- すべての Go コードは `gofmt` 済みであること。
- 変更対象のパッケージは少なくとも `go test` と `go vet` を通すこと。
- `error` は必ず確認すること。`_` で捨てないこと。
- 通常のエラー処理に `panic` を使わないこと。
- `context.Context` は第 1 引数 `ctx` で受け取り、struct に保持しないこと。
- export する型・関数・定数・変数には doc comment を付けること。
- interface は実装側ではなく利用側で定義すること。
- goroutine の終了条件と所有者が説明できないコードを書かないこと。

## 3. フォーマット・import・コメント
- フォーマットは `gofmt` を唯一の基準にします。
- `goimports` をローカル補助に使うのは構いませんが、最終的に `gofmt` で整っていることを前提にします。
- import の別名は衝突回避時だけに限定します。
- blank import は `main` パッケージか副作用が必要な test に限定します。
- dot import は循環参照回避が必要な特殊な test を除いて禁止します。
- top-level の export 名には doc comment を付け、コメントは宣言名から始まる完全な文にします。
- package comment が必要なパッケージでは `package` 宣言の直前に置きます。

## 4. パッケージ設計と配置
- package 名は短く、小文字で、1 つの責務を自然に表す名にします。
- `util`、`common`、`misc`、`api`、`types`、`interfaces` のような意味の薄い package 名は禁止します。
- package 名と識別子の重複を避けます。`user.UserService` より `user.Service` を優先します。
- 外部公開の必要がまだない実装は、まず `internal/` に置きます。
- 複数コマンドを持つ場合は `cmd/<command-name>/main.go` に配置します。
- サーバーやアプリの共通ロジックは `internal/` に寄せ、`main` には配線だけを残します。
- package を分ける基準は「責務が分かれるか」「import 循環を防げるか」「テストしやすくなるか」です。早すぎる細分化は避けます。

## 5. 命名と API 設計
- package 名は lower case の単語にします。`under_score` と `mixedCaps` は使いません。
- 頭字語は Go 流儀に合わせて `ID`、`URL`、`HTTP`、`JSON` のように保ちます。`appId` ではなく `appID` を使います。
- receiver 名は 1〜2 文字程度の短い一貫した名前にします。
- constructor や factory は、実際に抽象境界が必要な場合を除き interface ではなく concrete type を返します。
- 「mock しやすくするためだけ」の interface を実装側 package に置きません。
- pointer 引数は、本当に共有・変更・大型構造体回避が必要な場合に限ります。`*string` や `*io.Reader` のような不要な pointer 化は避けます。

## 6. receiver・値コピー・データ構造
- receiver が状態を変更するなら pointer receiver を使います。
- `sync.Mutex` など同期原語を含む struct は pointer receiver にし、値コピーを禁止します。
- 大きい struct や将来拡張されやすい struct も pointer receiver を優先します。
- pointer receiver を持つ型は安易に値コピーしません。
- 空 slice は `[]T{}` より `var s []T` を基本形にします。
- ただし JSON 契約上 `null` ではなく `[]` を返す必要がある場合は、非 nil の空 slice を明示的に使って構いません。

## 7. Context ルール
- `context.Context` は第 1 引数 `ctx context.Context` で受け取ります。
- `Context` を struct field に保持しません。
- `nil` Context は渡しません。未確定なら `context.TODO()` を使います。
- `context.WithCancel` / `WithTimeout` / `WithDeadline` で得た cancel 関数は全経路で必ず呼びます。通常は直後に `defer cancel()` を置きます。
- `context.Value` は request-scope の値伝搬だけに使い、optional 引数の代替には使いません。
- 各 call ごとに deadline・cancel・metadata を変えられる API を優先します。

## 8. エラー処理
- エラーは値として返します。失敗を `-1`、`""`、`nil` などの in-band 値だけで表現しません。
- 「説明不要な存在判定」は `(T, bool)`、それ以外は `(T, error)` を使います。
- 正常系を左に寄せ、異常系は早期 return で処理します。`if err != nil { ... } else { ... }` を常態化させません。
- エラー文字列は proper noun や略語を除いて小文字で始め、末尾に句読点を付けません。
- 追加情報を付けるときは `fmt.Errorf` を使います。
- `%w` で wrap するのは、その下位エラーを caller が `errors.Is` / `errors.As` で判定してよい API 契約にするときだけです。
- 実装詳細を外に漏らしたくない場合は wrap しません。
- `panic` / `recover` は package 内部の境界で閉じる場合を除き、通常フローに持ち込みません。

## 9. 並行処理と共有状態
- デフォルトは同期 API にします。並行実行が必要なら caller 側から goroutine を足せる設計を優先します。
- goroutine を起動したら、いつ止まるか・何で止めるか・誰が待つかをコード上で明確にします。
- channel send/receive で goroutine が取り残されない設計にします。
- map や可変 state への並行アクセスは mutex / channel / `sync/atomic` で保護します。
- 並行処理を含む package を変更したら `go test -race` を必須にします。
- トークン・鍵・秘密値の生成に `math/rand` を使いません。`crypto/rand` を使います。

## 10. テスト
- 非自明なロジックには `_test.go` を追加します。
- 入力と期待値の組が増える処理は table-driven test を基本にします。
- 各ケースを個別に見やすくしたいときは `t.Run` で subtest 化します。
- assert library は原則使いません。素の Go 条件式と `t.Errorf` / `t.Fatalf` を使います。
- failure message では `got` を先、`want` を後に書きます。
- failure message には関数名と主要入力を含めます。
- struct や map を返す処理は、項目ごとの断片比較ではなく全体比較を優先します。
- `t.Helper()` は責務が明確な補助関数にだけ使い、assert mini-language 化しません。
- parser / decoder / validator / 正規化処理 / 外部入力を受ける pure function には fuzz test を積極的に追加します。
- 新規 public package を追加するときは、使い方が伝わる `Example` またはそれに準ずる test を用意します。

## 11. 依存関係・module 管理
- Go コードは module 単位で管理し、`go.mod` と `go.sum` を必ず commit します。
- 依存追加・更新は `go get` など Go の標準ツールで行います。
- 依存変更後は `go mod tidy` を実行し、不要 dependency を残しません。
- 依存更新は「入れる理由」「上げる理由」が説明できるものだけに絞ります。
- 脆弱性確認のため、依存を触った変更や release 前には `govulncheck ./...` を実行します。

## 12. レビュー前チェック
- Go 変更を含む PR 前に最低限実行するもの:
  - `gofmt -w $(find . -name '*.go' -not -path './vendor/*')`
  - `go test ./...`
  - `go vet ./...`
- 並行処理を触ったときに追加で実行するもの:
  - `go test -race ./...`
- 依存を触ったときに追加で実行するもの:
  - `go mod tidy`
  - `govulncheck ./...`
- 入力境界や parser を触ったときに推奨するもの:
  - `go test -fuzz=Fuzz -fuzztime=10s ./...`

## 13. 参考にした公式資料
- Go Wiki: Go Code Review Comments
  - https://go.dev/wiki/CodeReviewComments
- Effective Go
  - https://go.dev/doc/effective_go
- Go Blog: Package names
  - https://go.dev/blog/package-names
- Go Blog: Errors are values
  - https://go.dev/blog/errors-are-values
- Go Blog: Working with Errors in Go 1.13
  - https://go.dev/blog/go1.13-errors
- Go Blog: Contexts and structs
  - https://go.dev/blog/context-and-structs
- `context` package documentation
  - https://pkg.go.dev/context
- Go Wiki: Go Test Comments
  - https://go.dev/wiki/TestComments
- Go Blog: Using Subtests and Sub-benchmarks
  - https://go.dev/blog/subtests
- Organizing a Go module
  - https://go.dev/doc/modules/layout
- Managing dependencies
  - https://go.dev/doc/modules/managing-dependencies
- Go Vulnerability Management
  - https://go.dev/doc/security/vuln/
- Data Race Detector
  - https://go.dev/doc/articles/race_detector
- `cmd/vet` documentation
  - https://pkg.go.dev/cmd/vet
