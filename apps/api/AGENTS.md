# Go API の実装規約

このファイルが API 作業の正。ルートの [AGENTS.md](../../AGENTS.md) は案内だけ。

- `internal/domain`: エンティティ、ビジネスルール、repository インタフェース。
  Go 標準ライブラリだけに依存する。生成型、Echo、DB ドライバを持ち込まない。
- `internal/usecase`: domain の型と repository インタフェースで業務操作を実行する。
  具象 repository、HTTP、SQL に依存しない。書き込み前に domain の検証を行う。
- `internal/repository`: SQL、永続化、DB エラーから domain エラーへの変換。
  handler/usecase を import しない。引数はプレースホルダで渡す。
- `internal/handler`: 生成された `ServerInterface`、入力検証、生成型と domain 型の変換、
  HTTP ステータスと契約のエラー形式への変換。具象 repository を import しない。
- `cmd/api`: 設定、DB、repository、usecase、handler の DI とプロセス起動。
  DB プール、context 対応 slog、シグナルと終了処理は `packages/go-platform`
  の共有モジュールを使う。BFF と同じ実装であり、複製しない（[ADR 0006](../../docs/adr/0006-shared-go-platform-module.md)）。
- `migrations`: goose の SQL Up / Down。`cmd/migrate` で実行し、API 起動で自動適用しない。
- `internal/testdb`: 統合テスト専用。production からの依存は禁止。

依存方向は `.go-arch-lint.yml` に宣言し、`make lint` で検査する。
公開契約の変更はユーザ承認を得る。このフェーズでは承認済みの NUL（U+0000）禁止を
`api/openapi.yaml` の Item / ItemInput の name / description に明記している。
生成物は手編集せず `make gen` で更新する。
操作追加時は生成インタフェースへのコンパイル時束縛を維持する。

リクエストに関連するログは必ず `slog.InfoContext(ctx, ...)` 等を使い、
`X-Request-ID` を保持した context を DB/ダウンストリームにも渡す。
認証は BFF が担当する。API を直接インターネットに公開しない。
BFF が付ける `X-User-ID` はネットワーク境界を信頼する。クライアント由来の
同名ヘッダを API が信用してはならない。

検証: `make test-unit` は DB 不要、`make test-go` は実 PostgreSQL と SIGTERM のテストを含む。
`TEST_DATABASE_URL` を指定しなければ Compose の DB を起動する。
テスト DB は専用 DB を使い、個々のテストが作った schema のみを削除する。
CI の Go ジョブは `make test-go` だけを実行し、Node.js の実行環境を必要としない。
最後に `make check` と `git diff --check` を実行する。
