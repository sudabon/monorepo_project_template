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
  リクエスト ID と health ルートは `packages/go-platform/echox` を使う。
- `cmd/api`: 設定、DB、repository、usecase、handler の DI とプロセス起動。
  環境変数は `packages/go-platform/config`、DB プール・ログ・shutdown は
  `packages/go-platform` の共有モジュールを使う。BFF と同じ実装であり、複製しない
  （[ADR 0006](../../docs/adr/0006-shared-go-platform-module.md)）。
- `migrations`: goose の SQL Up / Down。`cmd/migrate` は `packages/go-platform/migrate`
  経由で実行し、API 起動で自動適用しない。
- 統合テストの PostgreSQL スキーマ分離は `packages/go-platform/testdb`。
  production からの依存は禁止。

## リソースを 1 つ追加するときに書くもの

公開契約の変更はユーザ承認を得る。承認後、次の順で足す。`readInput` の JSON 解析をコピーしない。

1. **契約** — `api/openapi.yaml` に path・schema・エラー応答を足す。文字列の `minLength` / `maxLength` / `pattern` はここにだけ書く。`make gen` で `ServerInterface` と `internal/domain/constraints.gen.go` を更新する。生成物は手編集しない。`strict-server` は使わない（生成コードの既定バインドでは JSON null とフィールド欠損を区別できない）。
2. **handler** — 生成インタフェースを実装する。書き込み入力は `decodeFields` に契約順のフィールド名を渡し、返った `map[string]*string` を domain 入力へ詰め替え、`mergeFieldErrors` で型エラーと `Validate()` を契約順に合成する。Content-Type・サイズ・多重 JSON・null 判定は `decodeFields` が担う。
3. **usecase** — domain の型と repository インタフェースで操作する。書き込み前に `Validate()` を呼ぶ。HTTP と SQL を持ち込まない。
4. **repository** — SQL と domain エラーへの変換。handler/usecase を import しない。引数はプレースホルダで渡す。
5. **migration** — goose の SQL Up / Down。API 起動では自動適用しない。

操作追加時は生成インタフェースへのコンパイル時束縛を維持する。

依存方向は `.go-arch-lint.yml` に宣言し、`make lint` で検査する。

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
