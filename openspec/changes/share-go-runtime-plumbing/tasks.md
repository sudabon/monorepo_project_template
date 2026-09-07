## 1. 共有パッケージの追加

- [ ] 1.1 `packages/go-platform/config/` に `Require(name) (string, error)` / `Or(name, fallback) string` / `Bool(name, fallback) (bool, error)` を追加し、未設定・空文字・不正な bool のそれぞれで期待どおりの結果になる単体テストが通ることを確認する
- [ ] 1.2 `packages/go-platform/migrate/` に `Run(ctx, provider, arg string) error` を追加し、`up` / `down` / 不正な引数の 3 ケースを単体テストで確認する
- [ ] 1.3 `packages/go-platform/echox/` に `TraceRequest` を追加する。実装は BFF 側（`Clone(ctx)` とリクエストヘッダへの書き込み）を基にし、panic 整形は API 側（`debug.Stack()` を含む）を採る（design D2）
- [ ] 1.4 `TraceRequest` の単体テストで、(a) `X-Request-ID` が無いとき採番される (b) あるとき引き継がれる (c) レスポンスとリクエストの両方にヘッダが載る (d) panic がスタック付きのエラーになり本文には漏れない、の 4 点を確認する
- [ ] 1.5 `packages/go-platform/echox/` に `HealthRoutes(e *echo.Echo, ping func(context.Context) error)` を追加し、ping 成功で shallow/deep とも 200、ping 失敗で shallow は 200 のまま deep が 503 になることを単体テストで確認する
- [ ] 1.6 `packages/go-platform/echox/` に `WriteError(c, status, code, message)` を追加し、`Committed` 済みのレスポンスでは本文を書かずログのみ出すことを単体テストで確認する（design D3）
- [ ] 1.7 `packages/go-platform/testdb/` に `Open(t, newProvider)` を追加する。コンテキストは `t.Context()` に揃える（design D4）
- [ ] 1.8 `go-platform` に Echo を direct require として追加し、`go mod tidy -diff` が差分を出さないことを確認する

## 2. BFF の切り替え

- [ ] 2.1 `apps/bff/cmd/migrate/main.go` を `go-platform/migrate` 利用に書き換え、`make migrate-bff-up` / `make migrate-bff-down` が従来どおり動くことを確認する
- [ ] 2.2 `apps/bff/internal/testdb/database.go` を削除し、`go-platform/testdb` の利用に切り替えて `make test-go` の BFF 統合テストが通ることを確認する
- [ ] 2.3 `apps/bff/internal/handler/router.go` の `traceRequest` と health 2 本を `go-platform/echox` の利用に置き換え、`curl_test.go` と `auth_test.go` が通ることを確認する
- [ ] 2.4 `apps/bff/cmd/bff/main.go` の環境変数読み出しを `go-platform/config` に統一し、`BFF_COOKIE_SECURE` を `Bool` 経由にする。不正値でプロセスが起動失敗することを確認する
- [ ] 2.5 `router.go` を `router.go`（`New` と DI）/ `middleware.go`（`stripClientUserHeader` / `loadSession` / `protectCSRF` / `requireAuth`）/ `auth.go`（`login` / `logout` / `sessionStatus` と view 型と cookie）/ `errors.go`（`handleError`）に分割し、既存テストを 1 つも変えずに `make test-go` が通ることを確認する
- [ ] 2.6 `apps/bff/.go-arch-lint.yml` の `anyVendorDeps: true` を `canUse:` の明示に置き換え、`make lint-go` が通ることを確認する
- [ ] 2.7 許可していない外部パッケージの import を一時的に足すと `make lint-go` が失敗することを確認し、確認後に取り除く

## 3. API の切り替え

- [ ] 3.1 `apps/api/cmd/migrate/main.go` を `go-platform/migrate` 利用に書き換え、`make migrate-up` / `make migrate-down` が従来どおり動くことを確認する
- [ ] 3.2 `apps/api/internal/testdb/database.go` を削除し、`go-platform/testdb` の利用に切り替える。統合テストのコンテキストが `t.Context()` になったことを差分で確認する
- [ ] 3.3 `apps/api/internal/handler/router.go` の `traceRequest` と health 2 本を `go-platform/echox` に置き換え、`runtime_test.go` の panic テストと `items_test.go` が通ることを確認する
- [ ] 3.4 `apps/api/cmd/api/main.go` の環境変数読み出しを `go-platform/config` に統一する
- [ ] 3.5 `handler/items.go` の `uuid.MustParse` を `uuid.Parse` とエラー返却に変え、不正な ID を持つ行が panic ではなく 500 になることを単体テストで確認する（design D5）
- [ ] 3.6 `apps/api/.go-arch-lint.yml` の `anyVendorDeps: true` を `canUse:` の明示に置き換え、`make lint-go` が通ることを確認する

## 4. 後片付けと記録

- [ ] 4.1 `apps/{api,bff}` に残った旧実装が無いことを確認する（`traceRequest` / health / `testdb` / migrate の重複が grep で 0 件）
- [ ] 4.2 リクエスト ID 生成が標準ライブラリの `uuid` に揃い、`github.com/google/uuid` の参照が `apps/api/internal/handler` の型変換だけになっていることを確認する（design D5）
- [ ] 4.3 `docs/adr/0006-shared-go-platform-module.md` に、Echo 依存を `echox` に閉じる判断と、エラー語彙を共有しない判断を追記する
- [ ] 4.4 `packages/go-platform/AGENTS.md` に「置いてよいのは両サービスに同じ形で存在する配線だけ。業務ロジックと片側だけの都合は入れない」を明記する
- [ ] 4.5 `apps/api/AGENTS.md` と `apps/bff/AGENTS.md` の記述を移動後の実態に合わせる
- [ ] 4.6 `docs/STACK.md` の `google/uuid` 行を、用途を生成型の変換に限定した記述に直す
- [ ] 4.7 `make check` と `git diff --check` が通ることを確認する
