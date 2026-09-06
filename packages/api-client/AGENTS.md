# API クライアント（`packages/api-client`）の実装規約

このファイルが api-client 作業の正。ルートの [AGENTS.md](../../AGENTS.md) は案内だけ。
Go のファイルは開かない。

- `src/generated/` は `api/openapi.yaml` からの生成物。**手編集しない。** 型を変えたい
  ときは契約を直して `make gen` する。Biome の対象外にしてある。
- 公開するのは package entry からの `createItemQueries` / `createItemMutations` と
  `ApiError` だけ。アプリケーションに生成物を直接 import させない。
- ラッパは薄く保つ。TanStack Query の `queryOptions` / `mutationOptions` を返すに留め、
  再試行・エラー表示・キャッシュ無効化の方針を持たない。それはアプリ側の QueryClient が決める。
- HTTP エラーは `ApiError` に status と契約の body を載せて送出する。通信エラーは
  そのまま伝播させる。握りつぶさない。
- GET は Query の `AbortSignal` を fetch に渡す。
- 契約に無いエンドポイントをここに足さない。BFF の `/auth/*` は SPA 側が扱う。

## テスト

- `tests/` に置く。`tsconfig.json` の `include` は `src` と `tests` の両方。
  テストが型検査から外れる状態を作らない。
- 型の検証は `tests/*.test-d.ts`、通信の検証は `node:test` の `tests/*.test.ts`。
- 契約が変わったらテストの期待値も同じ PR で直す。

検証: `make gen-check-web` で生成差分、`make test-web` でテスト、`make check` で全体。
