# Phase 1 検証記録

実施日: 2026-09-06

## ローカル

- `make check`: 成功。Go vet / test / build、Biome、Spectral、TypeScript 型検証、
  既存 toolchain テスト 22 件、クライアント通信テスト 7 件、生成差分検査を実行した。
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
- 作成・更新・削除の入力と戻り値を生成型で検証し、不正な body / id と誤った戻り値の
  型をコンパイルで検出する。通信テストで POST / PUT / DELETE のパスと本文、
  mutation 実行前に通信しないこと、204 の本文なし応答、422 の全フィールドエラーの
  保持を確認した。
- `openspec validate phase1-openapi-contract-pipeline --strict`: 成功。

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

検証 PR はマージせずに close 済み。比較用コミットと CI 実行履歴は PR に残している。

## 最終レビュー

取得系 `queryOptions`、更新系 `mutationOptions` への変更はユーザー承認済み。
既存の `createItemQueries` のシグネチャと挙動を保ち、`createItemMutations` を追加した。
主担当が差分をレビューし、ラッパは key・通信関数・型の受け渡しに限定され、
リトライ、認証遷移、表示、キャッシュ無効化の方針を持たないことを確認した。
HTTP エラーは status / body を保持する既存の通信アダプタに通して呼び出し側へ伝える。
全 25 タスクを完了した。実サーバとの E2E は計画どおり Phase 2 / 5 の検証範囲とする。

## レビュー指摘の修正 (2026-09-06)

PR #5 のレビューで挙がった Should Fix 2 件を修正した。

- `packages/api-client/tests/client.test.ts` が `tsconfig.json` の `include` から漏れており、
  型検査されていなかった。`@types/node` 26.4.1 を追加して `include` を `["src", "tests"]` に広げ、
  `assert.rejects` のコールバック引数を `(error: unknown)` にした。`tsc --listFiles` に
  当該ファイルが現れることと、`items.list({ page: '2' })` を一時的に追加すると
  TS2322 で失敗することを確認した。DOM lib と `@types/node` の global 衝突は発生しない。
- `Makefile` の `gen-go` が GOTOOLCHAIN の導出を `scripts/go-task.sh` から複製しており、
  `go.work` が `X.Y` 形式のとき `make gen-go` だけが `invalid toolchain` で落ちる状態だった。
  `go-task.sh` に `gen` タスクを追加して導出を 1 か所に閉じた。`go 1.99` で `go1.99.0` に
  正規化されること、`go oops` で `must be X.Y or X.Y.Z` として落ちることを確認した。
  `gen-go` が `go-task.sh` に依存するようになったため、`contract.yml` のパスフィルタに
  `scripts/go-task.sh` を追加した。

修正後に `make check` と `make gen` を実行し、すべて成功・生成物の差分なしを確認した。
