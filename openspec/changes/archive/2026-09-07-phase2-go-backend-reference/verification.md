# 実装検証記録

2026-09-06、branch `feat/phase2-go-backend-reference`。

## 成功した検証

- `make check`: Go / Web 整形、go vet、依存方向 lint、Spectral、生成差分、
  Go 実 DB テスト、toolchain 22 テスト、api-client 7 テスト、型検査、build。
  このマシンの pnpm shim が壊れているため、固定版 11.25.0 を呼ぶ既存の
  `/tmp/phase1-toolchain/bin/pnpm` を PATH に追加して実行した。
- `TEST_DATABASE_URL=... make test`: CI と同じ外部 DB 指定方式でも成功。
  URL 未指定の `make test-go` は Compose を起動して同じテストを実行する。
- `make migrate-up` → `make migrate-down` → `make migrate-up`: すべて成功。
- 一時コピーで domain に repository 配下の package の import を追加すると
  go-arch-lint が失敗し、`internal/domain/forbidden.go:2` を出力。
- 一時コピーで `Items.DeleteItem` を外すと build が
  `does not implement generated.ServerInterface (missing method DeleteItem)` で失敗。
- 実 API バイナリへの curl: POST / GET / PUT / DELETE、404、ページ取得、
  複数フィールドの 422 と永続化されないことを確認。
- 実 API プロセスを終了・再起動し、作成済み item が一致することを確認。
- 実 PostgreSQL コンテナ停止中の shallow=200、deep=503 を確認。DB は再起動済み。
- Go subprocess テストで実 SIGTERM の drain / 待機上限を検証。
- レビュー指摘の応答書き込み失敗を再現し、ERROR ログと request_id を検証する
  テストの失敗を確認後に修正。再実行と scoped 再レビューで解消を確認。
- [PR #6](https://github.com/sudabon/monorepo_project_template/pull/6) の
  [Go CI](https://github.com/sudabon/monorepo_project_template/actions/runs/34022732951) が成功。
  commit `963461d` で PostgreSQL サービスを起動し、`make test`、lint、生成差分、build が
  すべて成功した。契約 / Web / E2E 計画ガード / Renovate も同じ commit で成功。
  初回 CI で検出した DB ヘルスチェックコマンドの引用符を修正後に再実行した。

2026-09-06 のユーザ承認に基づき、OpenAPI の Item / ItemInput の name / description に
NUL 禁止の pattern と 422 の説明を追加し、Go / TypeScript を再生成した。
型とエンドポイントに変更はなく、生成差分はフィールドの説明コメントのみ。
domain / HTTP / 契約バリデータの回帰テストで、修正前の失敗と修正後の成功を確認した。
POST / PUT の単一・複数フィールド拒否、他の制約との集約、永続化されないこと、既存 item が
変わらないことを検証した。通常の改行と文字列としての `\u0000` の保存・取得も成功。
E2E は tasks.md の指定に従い対象外。`test-plan.md` は追加していない。

## レビュー指摘の修正 (2026-09-06)

[PR #6 のレビュー](https://github.com/sudabon/monorepo_project_template/pull/6#issuecomment-5558179439)
の Should Fix 5 件を修正した。

- **S-1** `.github/workflows/go.yml` の `make test` を `make test-go` に戻し、
  `pnpm/action-setup` / `setup-node` / `make setup-web` を削除した。`make -n test` が
  `pnpm run test` を含むため、Go だけの変更で Web のテストが二重に走っていた。
  ADR 0001 の言語別分離と `docs/STACK.md` の記述に戻した。`apps/api/AGENTS.md` も更新。
- **S-2** `docs/STACK.md` に PostgreSQL 18.3-alpine / pgx 5.10.0 / goose 3.28.0 /
  go-arch-lint 1.18.0 / google/uuid 1.6.0 を追記し、「サーバ実装は Phase 2 で追加する」
  の陳腐化した記述を差し替えた。CI と DB の実行方法も追記。
- **S-3** `renovate.json` の `enabledManagers` に `docker-compose` を追加し、
  `compose.yaml` と `go.yml` の PostgreSQL を 1 つの PR にまとめる packageRule を足した。
  追加前は compose.yaml が更新対象外で、CI 側だけ版が上がる状態だった。
  `make validate-renovate` が成功。
- **S-4** `apps/api/go.mod` を tidy し、`contract_test.go` が直接 import している
  `github.com/getkin/kin-openapi` を direct 依存に直した。再発防止に
  `scripts/go-task.sh` の lint へ `go mod tidy -diff` を追加。修正前の go.mod で
  `make lint-go` 相当が exit 1 になることを確認してから tidy を実行した。
- **S-5** `internal/handler/router.go` の recover で `debug.Stack()` を error に含め、
  500 のログに panic 発生箇所が残るようにした。応答本文は従来どおり
  `internal_error` のみ。回帰テスト `TestPanicIsLoggedWithStackAndHiddenFromResponse`
  を追加し、スタックを落とした実装では `panic log lost the stack` で失敗することを
  確認してから修正版で成功させた。

修正後に `make check` の全通過を確認した。

## 制約

- NUL 入力は承認済みの追加制約として 422 で拒否する。未解決の `TODO(template)` は 0 件。
- offset pagination はリクエスト間の挿入・削除に対する snapshot 維持を保証しない。
  契約どおりの offset 方式であり、同一応答の total / items は同じ snapshot。
