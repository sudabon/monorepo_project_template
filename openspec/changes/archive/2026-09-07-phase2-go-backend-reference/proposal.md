## Why

Phase 1 で契約と生成パイプラインができたが、それを実装する側の型が決まっていない。案件ごとに層の切り方を考え直すと、複数案件を並行保守するときに毎回コードの読み方を思い出す必要がある。参照実装を 1 本通し、以降はそれをコピーして肉付けする状態にする。

また、ECS / ALB 上で運用する前提で必要になる要素（ヘルスチェックの粒度、graceful shutdown、リクエスト ID による追跡）は、後から入れると全ハンドラに手が入る。最初に骨格へ組み込む。

## What Changes

- 参考リポジトリ `monorepo_go_project_template` の Clean Architecture 構成を踏襲し、Phase 1 のサンプルリソースを handler / usecase / domain / repository の 4 層で実装する
- 依存の向きが内側（domain）を向いていることを、テストまたは lint で機械的に保証する
- DB アクセスの実装を 1 つ入れる。マイグレーションツールを選定して組み込む
- 構造化ログ（`log/slog`）を導入し、リクエスト ID を生成してログとダウンストリームに伝播する
- ヘルスチェックを 2 種類に分ける: プロセスの生死だけを見る shallow check と、依存先（DB 等）の生死を見る deep check
- graceful shutdown を実装する。SIGTERM 受信後、処理中のリクエストを完了させてから終了する
- Phase 1 で生成した Echo サーバインタフェースを実装し、実装漏れがコンパイルエラーになる構成にする
- ユーザ承認に基づき、name / description の NUL（U+0000）禁止を契約に追加し、作成・更新は 422 のフィールドエラーで拒否する

## Capabilities

### New Capabilities

- `sample-resource-api`: 契約に定義されたサンプルリソースの CRUD を HTTP 越しに提供する振る舞い。永続化、ページネーション、バリデーションエラー応答、存在しないリソースの扱いを含む。
- `service-runtime`: Go サービスの実行時の振る舞い。ヘルスチェックの粒度分離、graceful shutdown、構造化ログとリクエスト ID の伝播を含む。

### Modified Capabilities

- `api-contract`: PostgreSQL text に保存できない NUL（U+0000）を Item / ItemInput の name / description で禁止する。2026-09-06 にユーザ承認済み。エンドポイント、型、エラーの形式は維持する。

## Impact

- **新規ファイル**: `apps/api/` 配下の各層（handler / usecase / domain / repository）、DI 組み立て、マイグレーションファイル、`apps/api/AGENTS.md`
- **変更ファイル**: `Makefile`（マイグレーション実行と DB を要するテストのターゲット追加）、`.github/workflows/go.yml`（DB を伴うテストの実行環境追加）、`api/openapi.yaml` と Go / TypeScript の生成物（NUL 禁止の制約と説明）
- **依存追加**: Echo、DB ドライバ、マイグレーションツール、依存方向検査ツール（またはテストによる代替）。テスト用の DB 起動手段
- **後続への影響**: Phase 3 の BFF はこの API に転送する。Phase 5 の画面はこの API の応答形式に依存する
