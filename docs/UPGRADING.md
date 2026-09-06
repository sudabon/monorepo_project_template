# メジャー更新手順

このファイルは、実施時に判明した問題を書き足す場所である。手順の本文を「完成した手順書」として凍結しない。

## 追記の書式

各ライブラリの節の末尾 **「実施記録」** に、新しい見出しを古いものの上（新しい順）で足す。

```md
### YYYY-MM-DD: <ライブラリ> <旧> → <新>

- 作業者:
- 手順からの差分:
- 問題:
- 対処:
- 残課題:
```

バージョンの正は依存宣言（[STACK.md](STACK.md) の生成元）である。ここには手順と落とし穴だけを書く。

---

## React

対象: `react` / `react-dom` / `@types/react` / `@types/react-dom`。SPA と `packages/api-client` の peer / dev を同じ系列に揃える。

1. 公式のアップグレードガイドと Testing Library / Vite プラグイン / TanStack Query の peer を読む。
2. 上記 4 パッケージを同時に上げる。片側だけ 19 のままにしない。
3. `@vitejs/plugin-react`、`@testing-library/react`、`@types/react` の peer 警告を解消する。
4. JSX runtime、`ref`、`useFormStatus` などランタイムの破壊的変更をアプリとテストで確認する。
5. `make test-web && make build-web`、最後に `make check`。

壊れそうな点は [STACK.md](STACK.md) の React 行。

### 実施記録

（まだメジャー更新していない）

---

## Tailwind CSS

対象: `tailwindcss` と `@tailwindcss/vite`。PostCSS 設定は置かない前提のまま上げる。

1. v4 系の次メジャーの移行ガイドで `@theme inline`、プラグイン指令、Vite プラグインの読み込み方を確認する。
2. 2 パッケージを同じ版にする。
3. `apps/web` の CSS エントリとユーティリティクラスのビルド結果を目視する。
4. `make build-web`。

### 実施記録

（まだメジャー更新していない）

---

## TanStack Router

対象: `@tanstack/react-router` と `@tanstack/router-plugin`。生成物 `apps/web/src/routeTree.gen.ts` はコミットする。

1. プラグイン API、`beforeLoad` のシグネチャ、生成ファイル名の変更を changelog で確認する。
2. 2 パッケージを同時に上げ、開発サーバまたは `make build-web` で生成物を作り直す。
3. 認証ガード（`beforeLoad`）とファイルベースルートが型エラーなく解決することを確認する。
4. `make test-web && make check`。

### 実施記録

（まだメジャー更新していない）

---

## Go

対象: `go.work` と各 `go.mod` の `go` 行。Make は `go.work` から `GOTOOLCHAIN` を決める。

1. リリースノートの言語・vet・標準ライブラリの非互換を読む。
2. `go.work` と `apps/api` / `apps/bff` / `packages/go-platform` の `go` 行を同じ版にする。Renovate の `go toolchain` グループと同じ範囲。
3. `make lint-go && make test-go && make build-go`。
4. 固定値チェックが失敗したら、`.github/workflows/go.yml` の `go-version-file` が指す `apps/api/go.mod` と揃っているか見る。

Go は 1.x のままメジャーが長く動かない。コンパイラ診断の変化を「次のメジャー」相当として扱う。

### 実施記録

（まだメジャー更新していない）

---

## Echo

対象: `github.com/labstack/echo/v4`。API の生成サーバと BFF の実行時依存。oapi-codegen の Echo 生成設定と同時に見る。

1. Echo の次メジャー（v5）のハンドラシグネチャ、コンテキスト、ミドルウェアを読む。
2. oapi-codegen がそのメジャーに対応した版を先に確認する。生成設定と handler を同時に移行する。
3. `make gen-go` の差分をレビューし、手書き handler を生成インタフェースに合わせる。
4. BFF の router / CSRF / proxy も同じ Echo メジャーにする。片側だけ残さない。
5. `make test-go && make check`。

### 実施記録

（まだメジャー更新していない）
