## Why

Go と TypeScript の境界で型を二重管理すると、片方だけ直して気付かないまま本番に出る事故が起きる。テンプレート全体の中核はここで、契約を 1 か所に置いて両言語のコードを生成する仕組みが機能すれば、Phase 2 以降は肉付けになる。

## What Changes

- `api/openapi.yaml` を Go と TypeScript の唯一の契約として手書きで定義する。Go 実装からの仕様生成は採用しない（契約が実装に引きずられるため）
- サンプルリソース 1 つ分の定義を書く: 一覧取得 / 単体取得 / 作成 / 更新 / 削除、エラーレスポンススキーマ、ページネーション、バリデーションエラー形式
- Go 側: `oapi-codegen` で Echo のサーバインタフェースとリクエスト / レスポンス型を生成する
- TypeScript 側: `openapi-typescript` で型定義、`openapi-fetch` で型付き fetch クライアントを生成し `packages/api-client` に配置する
- `packages/api-client` に TanStack Query の取得系 `queryOptions` と更新系 `mutationOptions` を返す薄いラッパを手書きで追加する。生成物の上に 1 枚被せる形とし、生成物そのものは編集しない
- `api/openapi.yaml` の lint を CI に組み込む（Spectral または Redocly）
- 生成物をすべてコミットし、CI に `make gen && git diff --exit-code` を入れて仕様と生成物がずれた PR を落とす

## Capabilities

### New Capabilities

- `api-contract`: API 契約の単一ソース化と、そこから Go / TypeScript のコードを生成するパイプラインの振る舞い。契約が定義するサンプルリソースの操作・エラー形式・ページネーション、生成の冪等性、仕様と生成物の不一致検出を含む。

### Modified Capabilities

なし。`make gen` が空ターゲットから実際の生成処理に変わり、契約検証ワークフローが検証内容を持つが、`build-toolchain` の要求（単一のタスク入口・パス分離・バージョン固定・依存更新方針）そのものは変わらないため、delta spec は作らない。生成物の差分検出は本 change の `api-contract` 側の要求として定義する。

## Impact

- **新規ファイル**: `api/openapi.yaml`、`api/.spectral.yaml`（または Redocly 設定）、Go 生成物（`apps/api` 配下の generated パッケージ）、`packages/api-client/`（生成物 + 手書きラッパ + `package.json` + `tsconfig.json`）
- **変更ファイル**: `Makefile`（`gen` ターゲットの実体化）、`.github/workflows/contract.yml`（lint と差分検証の追加）、`.github/workflows/web.yml`（`packages/api-client` をパスフィルタに追加）
- **依存追加**: `oapi-codegen`（Go、ツール依存）、`openapi-typescript`（TS、dev 依存）、`openapi-fetch`（TS、実行時依存）、`@tanstack/react-query`（`queryOptions` / `mutationOptions` の型のため peer として扱う）、契約 lint ツール
- **後続への影響**: Phase 2 は生成された Echo サーバインタフェースを実装する。Phase 4/5 は `packages/api-client` の `queryOptions` / `mutationOptions` を使う
