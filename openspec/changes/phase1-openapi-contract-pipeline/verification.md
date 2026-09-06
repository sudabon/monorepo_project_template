# Phase 1 検証記録

実施日: 2026-09-06

## ローカル

- `make check`: 成功。Go vet / test / build、Biome、Spectral、TypeScript 型検証、
  既存 toolchain テスト 22 件、クライアント通信テスト 3 件、生成差分検査を実行した。
  BFF はソース未追加のため Go 処理をスキップ。E2E 計画ガードは既存の
  `templateRepo: true` に従ってスキップする。
- `make gen` を 2 回実行し、各回 `git diff --exit-code` が成功。
- `make fmt` 後も `make gen-check` が成功。TypeScript の生成物は Biome が整形しない。
- 一時仕様で operationId を重複させると Spectral が失敗し、
  `paths./items.post.operationId` と行番号を表示した。
- 一時的な未追跡ファイルを生成先に追加すると `make gen-check-web` が
  `Untracked generated files` で失敗した。検証後に一時ファイルを削除した。
- `Item.contractProbe` を隔離した検証用 clone の仕様だけに追加し、`make gen` により
  Go の `*string` と TypeScript の省略可能な `string` の両方に反映されることを確認した。
  このフィールドは実装ブランチに含めない。
- パス・メソッド・入力・ページ引数の誤りは型エラーとして検出する。description を
  省略した POST / PUT の正常入力も型テストに含めた。
- 取得のページ引数のシリアライズ、HTTP エラーの status / body の保持、Query の
  キャンセルが fetch に伝播することをテストした。

ローカルでは既存の pnpm ランチャーの固定版切り替えが停止したため、npm 配布の
同じ pnpm 11.25.0 を使用した。一時 PATH の `/tmp/phase1-toolchain/bin` にランチャーを
置き、既存の `/tmp/phase0-toolchain/pnpm-store` を利用してコマンドを実行した。
プロジェクトの Node / pnpm / Go の固定値は変更していない。

## CI の生成差分検出

検証用 [Draft PR #4](https://github.com/sudabon/monorepo_project_template/pull/4) は
`test/phase1-contract-drift` ブランチを用い、実装ブランチから分離した。

- `3f9edca`: 仕様だけに検証フィールドを追加。Contract の lint は成功し、
  `make gen-check` が両言語の差分を検出して失敗。
  [失敗実行](https://github.com/sudabon/monorepo_project_template/actions/runs/34000634418)
- `0244d02`: 同じ仕様に対応する Go / TypeScript の生成物だけを追加。
  Contract / Go / Web / Renovate / E2E plan がすべて成功。
  [成功実行](https://github.com/sudabon/monorepo_project_template/actions/runs/34000722278)

検証 PR はマージせずに close する。比較用コミットと CI 実行履歴は PR に残す。

## 未完了

タスク 4.4 の全操作 `queryOptions` 指定について、取得系を `queryOptions`、
作成・更新・削除を `mutationOptions` に変更する設計確認を依頼中。
回答後に更新系ラッパと型・通信テストを実装し、タスク 4.5 の責務分離を最終確認する。
