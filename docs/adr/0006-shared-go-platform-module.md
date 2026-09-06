# 0006: Go の実行時基盤を共有モジュールに置く

- 状態: 採用
- 日付: 2026-09-06

## 判断

`log/slog` の context ハンドラ、graceful shutdown、pgx の接続プールを
`packages/go-platform` の 1 モジュールに置き、`apps/api` と `apps/bff` の両方が
import する。各 module 配下の `internal/platform/` は廃止する。

モジュールは公開しないため、`apps/api/go.mod` と `apps/bff/go.mod` に
`replace` を書き、`go.work` にも登録する。`replace` があるので、workspace の
外（案件のコンテナビルド等）でも module 単体で解決できる。

## 根拠

Phase 3 で BFF を追加した時点で、`platform/logging/logging.go` は 2 つの module に
**バイト単位で同一**のコピーが並んだ。`server.go` と `database.go` も同一だったが、
コピーの過程で ECS の `stopTimeout` と接続プール本数の根拠コメントが BFF 側から
落ちていた。コピー直後の 1 フェーズで既に劣化しており、規約で防げる性質ではない。

同時に、API と BFF が同じ PostgreSQL に各 10 本ずつ接続するようになった。
プール本数の見積もりは 1 か所に書かないと、片方だけ直して整合が崩れる。

## トレードオフ

- go.work に module が 1 つ増え、CI のパスフィルタと go-arch-lint の設定が増える。
  `web.yml` の `packages/**` は Go の module を拾わないよう除外する
- 案件が BFF を捨てる構成にすると、利用者 1 つの共有モジュールが残る。
  その場合は `internal/` に戻すのが自然であり、`replace` を消すだけで済む

## 依存方向の扱い

go-arch-lint から見ると共有モジュールは vendor になる。`vendors.platform` として
宣言し、外部依存を広く許していない component（BFF の `proxy` 等）には
`canUse: [platform]` で個別に許可する。`anyVendorDeps: true` に緩めない。

参考: [ADR 0003](0003-api-database-and-architecture.md)、[docs/bff.md](../bff.md)。
