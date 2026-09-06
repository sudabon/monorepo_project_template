# Go 共有基盤（`packages/go-platform`）の実装規約

このファイルが go-platform 作業の正。ルートの [AGENTS.md](../../AGENTS.md) は案内だけ。
TypeScript のファイルは開かない。

`apps/api` と `apps/bff` の両方が import する。判断の背景は
[ADR 0006](../../docs/adr/0006-shared-go-platform-module.md)。

- `logging`: `log/slog` の JSON ハンドラと、context 経由のリクエスト ID 伝播。
  ヘッダ名 `X-Request-ID` はここが唯一の定義。
- `server`: シグナル受信と graceful shutdown。`SHUTDOWN_TIMEOUT` の既定 20 秒は
  ECS の `stopTimeout` より短いことが前提。根拠のコメントを消さない。
- `database`: pgx の `database/sql` プール設定。接続本数は API と BFF の合計タスク数で
  見積もる。片方だけを見た数字にしない。

## 禁止

- **サービス固有のロジックを置かない。** 片方だけが使うものは、その `apps/*/internal/` に置く。
- **`apps/*/internal/platform/` に複製し直さない。** 複製がコメントごと劣化したのが
  この module を作った理由。
- `apps/api` / `apps/bff` を import しない。依存の向きは常にこちらが内側。

## 変更するとき

- 両サービスに同時に効く。片方だけの都合で signature を変えない。
- module は未公開。各 `go.mod` の `replace` と `go.work` で解決している。
  パスを変えるなら 3 か所（`go.work`、`apps/api/go.mod`、`apps/bff/go.mod`）を同時に直す。
- 依存方向の lint からは vendor に見える。各サービスの `.go-arch-lint.yml` は
  `vendors.platform` として宣言し、`canUse: [platform]` で個別に許可している。
  `anyVendorDeps: true` に緩めない。

検証: `make lint-go`（go vet・`go mod tidy -diff`）、`make test-go`、最後に `make check`。
