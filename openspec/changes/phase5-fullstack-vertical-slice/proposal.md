## Why

Phase 2〜4 で層ごとの実装は揃うが、DB からブラウザまで 1 本通っていないと「この組み合わせで本当に動くのか」が分からない。案件側が最初に見るのはこの貫通した 1 リソースであり、以降の画面はこれをコピーして作る。参照実装として通し、同時に回帰を検出するスモークを 1 本置く。

## What Changes

- サンプルリソースの一覧画面を実装する。ページネーションと検索条件を持ち、条件は URL に保持する（TanStack Router の型付き search params を使う）
- 詳細画面を実装する
- 作成 / 編集フォームを実装する。Phase 4 で用意したサーバ側バリデーションエラーのマッピングを実際に通す
- 削除を実装する。確認ダイアログを挟む
- 上記の一連の流れを Playwright のスモークテスト 1 本にする
- CI でスモークを実行する

## Capabilities

### New Capabilities

- `sample-resource-ui`: サンプルリソースをブラウザから操作する振る舞い。一覧の検索条件とページネーションの URL 保持、詳細表示、作成・編集時のサーバ検証エラー表示、確認を伴う削除を含む。

### Modified Capabilities

なし。`sample-resource-api` / `session-auth` / `web-app-shell` の要求は変更せず、それらを組み合わせた画面側の振る舞いを新たに定義する。

## Impact

- **新規ファイル**: `apps/web/src` 配下の一覧 / 詳細 / 作成 / 編集ルートとコンポーネント、`tests/e2e/` のスモークテスト・Page Object・シード fixture、`tests/e2e/fixtures/README.md` への追記
- **変更ファイル**: `.github/workflows/web.yml`（スモーク実行の追加）、`Makefile`（E2E 実行ターゲット）、`playwright.config.ts`（必要に応じて webServer 設定）
- **依存追加**: Playwright（Phase 0 時点で導入済みなら追加なし）。アプリケーション側の新規依存は追加しない想定
- **前提**: Phase 2 の API、Phase 3 の BFF、Phase 4 の SPA 骨格が動作すること。E2E は 3 つを同時に起動できる構成を必要とする
