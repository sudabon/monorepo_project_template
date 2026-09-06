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

## 障害時の初動

ログは JSON、標準出力。リクエストに紐づく行は `request_id` を持つ。

1. 利用者・ALB・CSP から時刻とパスを聞く。可能なら応答ヘッダ `X-Request-ID` を取る。
2. なければ再現時に自分で付ける。未指定ならサーバが UUID を生成し、応答ヘッダに返す。

```sh
make db-up && make migrate-up
make run-api
# 別端末
id=ops-trace-1
curl -sS -D - -o /dev/null -H "X-Request-ID: $id" http://127.0.0.1:8080/health/shallow
# make run-api の出力から:
# {"time":"...","level":"INFO","msg":"request completed","request_id":"ops-trace-1",...}
```

3. 同じ ID で BFF と API のログを突合する。BFF は受信ヘッダを API へ転送する。SPA → BFF → API が 1 本になる。
4. `status` が 5xx の行、`msg` が `request completed` 以外の ERROR / WARN を先に見る。内部エラーの詳細はレスポンス本文ではなくログだけにある。
5. `GET /health/deep` で DB を切り分ける。503 ならアプリより先にデータベースを疑う。

`X-Request-ID` は受信値を引き継ぎ、未指定なら UUID を生成する。成功・失敗とも
応答ヘッダに返す。BFF はこのヘッダを API に渡す。

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
`reportError` に渡す。既定実装は何も送信しない。送信先の SDK や DSN はテンプレートに含めない。

差し込み:

1. 案件の監視 SDK を `apps/web` の依存に追加する（本番バンドルに入る。後述のレビュー基準を適用する）。
2. 起動時（`config.json` 取得の前後どちらでも、PII を送る前）に `setErrorReporter` へ送信関数を渡す。
3. 引数は `unknown` の error と任意の `componentStack`。ユーザ名・メール・トークン・リクエスト本文を付けない。
4. `setErrorReporter` 以外の経路（window.onerror の二重送出など）を足すなら、同じ ErrorBoundary と重複しないことをテストする。

詳細は [apps/web/AGENTS.md](../apps/web/AGENTS.md)。認証は BFF が担当する。API は信頼できる内部ネットワークに置く。
BFF の Cookie 属性・CSRF・セッション寿命は [bff.md](bff.md) を参照。
offset pagination は固定されたデータ集合で重複・欠落がなく、同一応答の total/items は
DB の同じ snapshot を使う。複数リクエストの間に挿入・削除があればページ境界は動く。

## 依存更新 PR のレビュー基準

判断の軸は **本番配信物に入るか**。入るものの脆弱性と破壊的変更を先に見る。

| 入るか | 例 | レビュー |
| --- | --- | --- |
| 入る | `apps/web` の `dependencies`、Go の実行時 require（Echo、pgx、goose 等） | 差分・CVE・互換を通常どおり見る。自動マージ対象でも changelog の破壊的変更は目を通す |
| 入らない | `devDependencies`（Vitest、Biome、Playwright、型パッケージ）、Go の `tool`（oapi-codegen、go-arch-lint） | 本番の実行パスは変わらない。CI と開発者マシンの再現性、ライセンス、**ビルド時に実行されること**を見る |

ビルド時ツールのサプライチェーンは別軸である。本番に入らないことは「無視してよい」ではない。
`postinstall` が動くパッケージ、コード生成器、GitHub Actions は、実行時依存より優先度を下げてよいが、放置しない。
major は個別 PR のまま自動マージしない（[STACK.md](STACK.md)）。

バージョン表が依存宣言とずれていないかの確認は、[STACK.md](STACK.md) 先頭の突き合わせ先を PR の対象ファイルと照合する。

## npm audit の切り分け

```sh
pnpm audit
```

1. 本番配信物に入る依存（`apps/web` / `packages/api-client` の `dependencies`）の脆弱性を先に直す。
2. `devDependencies` の脆弱性は本番バンドルには届かない。優先度は下げる。ただし前項のサプライチェーン軸は残る。開発時・CI でそのツールがコードを実行するなら、修正または回避策を issue にする。
3. 修正できない推移依存は、影響面（入力は何か、ネットワークに出るか）を一行で PR に書く。黙って close しない。

## Aurora のアップグレード

ローカルと CI の PostgreSQL 版は [STACK.md](STACK.md)。本番を Aurora PostgreSQL にする案件向け。テンプレートにクラスタ定義はない。

### マイナー

1. リリースノートで拡張・レプリケーション・パラメータの注意を読む。
2. 非本番クラスタでバージョンを上げ、`make test-go` 相当の結合とマイグレーションを通す。
3. 本番はメンテナンスウィンドウで適用する。フェイルオーバーが起きる前提で、ALB の shallow が 200 に戻ることを確認する。
4. アプリの `pgx` と goose の版は、サーバのマイナーだけでは通常変えない。接続エラーが出たらドライバ側を疑う。

### メジャー

1. メジャー間の非互換（予約語、廃止関数、ダンプ形式）を読む。
2. スナップショットまたはブルー/グリーンで新メジャーのクラスタを作る。
3. パラメータグループを新メジャー用に作り直す。旧グループを流用しない。
4. 非本番で `make migrate-up` 済みのスキーマを載せ、CRUD と BFF セッションを確認する。
5. アプリの接続先を切り替え、deep ヘルスとリクエスト ID で数本トレースしてから旧クラスタを残期間だけ保持する。
6. ダウングレードはしない。戻すならスナップショット復元。

## Terraform provider のメジャー更新

インフラコードはテンプレート範囲外。案件で Terraform を足したあとの手順。

1. provider のメジャー changelog と `required_providers` の制約を読む。
2. ロックファイル（`.terraform.lock.hcl`）を含む作業ブランチで `terraform init -upgrade` する。
3. `terraform plan` の destroy / replace を先に見る。state 移動が要る変更は `moved` ブロックを同 PR に含める。
4. 非本番 workspace で apply し、shallow / deep と静的配信（[web.md](web.md)）を確認する。
5. 本番 apply は plan ファイルを保存してから行う。provider とモジュールの major を同じ PR に混ぜない。
