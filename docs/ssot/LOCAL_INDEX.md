# SSOT Local Index

このディレクトリは `zenyonedad-glitch/short-fans` の `_project/short-fans/ssot` を、このリポジトリ向けに取り込んだ実体コピーです。

- upstream の取り込み元は [SOURCE.md](SOURCE.md) を参照
- 実データの再同期は `./scripts/sync_ssot.sh [ref]`
- upstream の複数ファイルには元リポジトリの絶対パス表記が含まれるため、この repo での入口はこのファイルを優先

## 読み始める順番

1. [README.md](README.md)
2. [product/index.md](product/index.md)
3. [business/index.md](business/index.md)
4. [product/scope/mvp-boundaries.md](product/scope/mvp-boundaries.md)
5. [product/fan/core-experience.md](product/fan/core-experience.md)

## この repo での参照先

- ルートの全体マップ: [index.md](index.md)
- プロダクト領域: [product/index.md](product/index.md)
- 事業領域: [business/index.md](business/index.md)
- 参考資料: [refs/reference_content.md](refs/reference_content.md)

## 取り扱い

- `docs/ssot/` 配下の upstream ファイルは vendor されたドキュメントとして扱う
- この repo 固有の補足や同期メタデータは `LOCAL_INDEX.md` と `SOURCE.md` に寄せる
- `/Users/yonedazen/Projects/short-fans/_project/short-fans/ssot/...` は、この repo では `docs/ssot/...` と読み替える
- `/Users/yonedazen/Projects/short-fans/memo/...` は参照元のままで、この repo には取り込んでいない
