## Why

[ADR 0006](../../../docs/adr/0006-shared-go-platform-module.md) は「API と BFF で同じ実装を複製しない」ために `packages/go-platform` を作った。しかし実際に移したのは `logging` / `server` / `database` の 3 つだけで、その外側では複製が続いている。

重複のコストは仮定ではなく既に発生している。`internal/testdb/database.go` は 63 行が重複しているが、BFF 側だけが Go 1.24 以降の `t.Context()` を使い、API 側は `context.Background()` のまま取り残されている。`traceRequest` も同様で、panic 時に `debug.Stack()` を残す修正は API 側にしか入っていない。**片方だけが直る事故が、既に 2 件起きている。**

テンプレートは複製されて使われる。この重複は案件数だけ増える。

## What Changes

- `apps/{api,bff}/cmd/migrate/main.go`（43 行が import 1 行を除いて完全一致）の共通部分を `packages/go-platform` へ移す
- `apps/{api,bff}/internal/testdb/database.go`（63 行の重複。既にドリフト済み）を `packages/go-platform` へ統合し、`t.Context()` に揃える
- 両サービスの `traceRequest`（各 ~28 行）と health チェック 2 本（shallow / deep）を `packages/go-platform` へ移す。Echo 依存は専用サブパッケージに閉じる
- エラー応答の封筒（`{code, message}` の書き出し）だけを共有する。**HTTP ステータスとコード語彙の対応表は各サービスに残す**（API は 422 の検証エラー、BFF は 403 CSRF / 413 payload、と語彙が異なるため）
- `apps/bff/internal/handler/router.go`（285 行 / 8 責務）を `router.go` / `middleware.go` / `auth.go` / `errors.go` に分割する
- 環境変数の読み出しを 1 つのイディオムに揃える。現在は `if x == ""` と `cmp.Or` が混在し、`BFF_COOKIE_SECURE` は `strconv.ParseBool` を使わず自前判定している
- UUID 生成を標準ライブラリの `uuid` に揃える。`github.com/google/uuid` は生成コードが要求する型変換だけに残す
- `apps/api/internal/handler/items.go` の `uuid.MustParse` をエラー返却に変える（panic を経由せず 500 を返す）
- 両 `.go-arch-lint.yml` の `anyVendorDeps: true` を `canUse:` の明示に置き換える。現在 `handler` / `main` / `migrations` / `testdb` は外部依存が無検査になっている

## Capabilities

### New Capabilities

なし。

### Modified Capabilities

大半は実装の置き場所が変わるだけで、観測可能な振る舞いは変えない。要求が変わるのは次の 2 つに限る。

- `service-runtime`: リクエスト ID の伝播に「下流サービスへの引き継ぎ」を追加する。現在これを行っているのは BFF のプロキシだけで、API 側の `traceRequest` はリクエストヘッダに ID を書かない。統一実装は BFF 側を採るため、要求として明文化する
- `sample-resource-api`: 依存方向の機械的検査に「各層が利用してよい外部モジュールの許可リスト」を加える。現在 `handler` / `main` / `migrations` / `testdb` は `anyVendorDeps: true` で外部依存が無検査であり、検査の範囲が要求より狭い

health チェックの深浅分離、graceful shutdown、`session-auth` の要求は変わらない。既存のテストがそのまま回帰検査になる。

## Impact

- **新規**: `packages/go-platform/echox/`（Echo 依存の middleware と health ルート）、`packages/go-platform/migrate/`、`packages/go-platform/testdb/`、`packages/go-platform/config/`
- **変更**: `apps/api/cmd/{api,migrate}/main.go`、`apps/bff/cmd/{bff,migrate}/main.go`、両 `internal/handler/router.go`、両 `internal/testdb/database.go`、両 `migrations/migrations.go`、両 `.go-arch-lint.yml`、`apps/api/internal/handler/items.go`
- **削除**: `apps/{api,bff}/internal/testdb/`（go-platform へ移動）
- **依存**: 追加なし。`packages/go-platform/go.mod` に Echo が direct require として増える（両アプリで既に使用中）
- **ドキュメント**: `docs/adr/0006` に Echo 依存の扱いを追記。`packages/go-platform/AGENTS.md`、`apps/{api,bff}/AGENTS.md` の記述を実態に合わせる
