## Why

SPA は「ディレクトリ構成以外の 9 割」が最初の設計で決まる。環境ごとにビルドを作り直す構成にすると、ステージングと本番で別々の成果物を管理することになり、案件が増えるほど破綻する。エラー処理と認証ガードも、後から入れると全画面に手が入る。骨格の段階で通しておく。

## What Changes

- Vite + React + TypeScript の構成を `apps/web` に作る。`tsconfig` は `strict: true` に加えて `noUncheckedIndexedAccess` を有効にする
- 環境ごとの設定をビルド時に埋め込まない。ビルド成果物は全環境で 1 つとし、起動時に `/config.json` を fetch して API のベース URL などを読む
- `config.json` の型定義と、読み込み失敗時のフォールバック挙動を実装する
- `.env.example` に、SPA には秘密情報を置けない旨をコメントで明記する
- TanStack Router をファイルベースルーティングで設定し、認証ガードを `beforeLoad` に実装する
- TanStack Query の `QueryClient` を 1 か所で設定する。`staleTime` / `retry` / `refetchOnWindowFocus` の既定値を明示的に決め、理由をコメントに書く
- グローバルな `onError` を実装する。401 はセッション切れとしてログインへ、5xx はトースト、それ以外は握りつぶさず表示する
- Tailwind CSS を設定する
- 汎用 UI コンポーネント（ボタン、入力、モーダル、テーブル、トースト）を `apps/web/src/components/ui/` に自分のコードとして置く。UI ライブラリを依存に追加しない
- ルート単位の ErrorBoundary と、最上位の ErrorBoundary を置く。最上位には監視サービス送信フックを空実装で用意する
- ローディングの扱い（Suspense を使うか否か）を決めて統一する
- react-hook-form + zod を導入し、サーバ側バリデーションエラーをフォームのフィールドエラーへマッピングする共通処理を書く
- ダークモードは実装しない

## Capabilities

### New Capabilities

- `web-app-shell`: SPA の骨格が提供する振る舞い。実行時設定の読み込みとフォールバック、認証ガードを含むルーティング、サーバ状態の既定方針とグローバルエラー処理、エラー境界による障害時表示、フォームの検証とサーバ検証エラーの反映、UI 部品の内製方針を含む。

### Modified Capabilities

なし。既存の `session-auth` / `api-contract` の要求は変更せず、それらを利用する側の振る舞いを新たに定義する。

## Impact

- **新規ファイル**: `apps/web/` 配下一式（Vite 設定、`tsconfig`、ルーティング、`QueryClient` 設定、`components/ui/`、ErrorBoundary、フォーム共通処理、`public/config.json` のサンプル、`.env.example`）、`apps/web/AGENTS.md`
- **変更ファイル**: `pnpm-workspace.yaml`（既に登録済みなら変更なし）、`Makefile`（web の dev / build ターゲット）、`.github/workflows/web.yml`（テストとビルドの追加）
- **依存追加**: React、Vite、TanStack Router、TanStack Query、Tailwind CSS、react-hook-form、zod、Vitest、Testing Library。UI コンポーネントライブラリは追加しない
- **後続への影響**: Phase 5 の画面はこの骨格の上に載る。エラー表示・フォーム・認証ガードの実装はここで確定する
