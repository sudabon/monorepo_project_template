# API と BFF のローカル起動と運用

```sh
make setup
make db-up
make migrate-up
make migrate-bff-up
make run-api
make run-bff
make dev-web
```

API は `http://localhost:8080/api/items`、BFF は `http://localhost:8081`、SPA は Vite の開発サーバ（既定 `http://localhost:5173`）で待ち受ける。
`/api` と `/auth` は Vite が BFF へプロキシする。ブラウザ向けの呼び出しは BFF 経由にする。DB はローカルの `localhost:55432`。
開発用のユーザ・パスワード・DB 名は `template`。BFF のデモログインは `demo` / `demo`。
停止は各プロセスに Ctrl-C、DB に `make db-stop`。DB の volume は保持される。
直前の migration を戻すコマンドは `make migrate-down` / `make migrate-bff-down`。
items の初期 migration の rollback はテーブルを削除するため、必要なデータのある
DB では実行しない。

| 設定 | 既定値・用途 |
| --- | --- |
| `DATABASE_URL` | Make 経由は開発 DB。バイナリを直接起動する場合は必須。PostgreSQL 接続 URL |
| `HTTP_ADDR` | API 既定 `:8080`、BFF 既定 `:8081`。`make run-bff` は `:8081` |
| `SHUTDOWN_TIMEOUT` | `20s`。SIGTERM 後の処理完了待ち上限。正の Go duration |
| `TEST_DATABASE_URL` | 未設定なら Compose を起動。設定済みならその専用 DB に一時 schema を作成 |
| `POSTGRES_PORT` | Compose のホスト側ポート。既定 `55432`。変更時は `DATABASE_URL` も合わせる |
| `BACKEND_URL` | BFF が転送する API の絶対 URL。`make run-bff` は `http://127.0.0.1:8080` |
| `BFF_DEMO_USERNAME` / `BFF_DEMO_PASSWORD` | テンプレートのデモユーザ。`make run-bff` は `demo` / `demo` |
| `BFF_COOKIE_SECURE` | 既定 `true`。`make run-bff` はローカル HTTP 用に `false` |

`GET /health/shallow` は DB に接続せず 200 を返す。ALB のヘルスチェックには
このパスを指定する。`GET /health/deep` は DB の Ping を最大 2 秒待ち、
依存先ごとの結果を `dependencies.database` に返す。正常は 200 / `healthy`、
接続できない場合は 503 / `unhealthy`。deep は監視・調査に利用する。

`X-Request-ID` は受信値を引き継ぎ、未指定なら UUID を生成する。成功・失敗とも
応答ヘッダに返し、リクエスト関連の JSON ログの `request_id` で検索できる。
BFF はこのヘッダを API に渡す。内部エラーの詳細はログだけに記録する。

name / description に NUL（U+0000）は使用できない。JSON の `\u0000` を含む
作成・更新リクエストは、DB に保存する前に 422 / `validation_error` で拒否する。
複数フィールドに含まれる場合はすべて `errors` に返し、既存データも変更しない。
文字としてのバックスラッシュを含む `\\u0000` や通常の改行は使用できる。

SIGTERM / Ctrl-C 後は新規受付を閉じ、進行中の応答を待つ。上限到達時は接続を閉じ、
`shutdown deadline reached` を記録して終了する。ECS の `stopTimeout` が 30 秒なら
既定の 20 秒を使い、残りを DB の切断とログ転送に確保する。変更時も必ず
`SHUTDOWN_TIMEOUT < stopTimeout` とする。

```sh
make test-unit   # domain/usecase のテスト
make test-go     # 実 DB の CRUD/validation、migration、SIGTERM、ログ捕捉
make test        # Go と TypeScript のテスト
make check      # 整形、lint、生成差分、テスト、build、E2E 計画ガード
```

## SPA の監視サービス差し込み

最上位の ErrorBoundary は捕捉した例外を `apps/web/src/monitoring/reportError.ts` の
`reportError` に渡す。既定実装は何も送信しない。Sentry や Datadog を足すときは
`setErrorReporter` に送信関数を渡し、PII を送らないこと。詳細は
[apps/web/AGENTS.md](../apps/web/AGENTS.md)。

認証は BFF が担当する。API は信頼できる内部ネットワークに置く。
BFF の Cookie 属性・CSRF・セッション寿命は [bff.md](bff.md) を参照。
offset pagination は固定されたデータ集合で重複・欠落がなく、同一応答の total/items は
DB の同じ snapshot を使う。複数リクエストの間に挿入・削除があればページ境界は動く。
