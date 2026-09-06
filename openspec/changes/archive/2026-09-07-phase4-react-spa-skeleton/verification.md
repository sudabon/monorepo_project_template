# 実装検証記録

2026-09-06、branch `feat/phase4-react-spa-skeleton`。

## 自動検証の範囲

- `make check`: Go / Web 整形、go vet、`go mod tidy -diff`、依存方向 lint、Spectral、
  生成差分、実 PostgreSQL を含む Go テスト、Vitest、型検査、build、E2E 計画ガード。
  このマシンの pnpm shim が壊れているため、固定版 11.25.0 を呼ぶ既存の
  `/tmp/phase1-toolchain/bin/pnpm` を PATH に追加して実行した。
- 実行時設定: 取得失敗とスキーマ不一致でエラー画面になること、スキーマ不一致のときに
  発行された fetch が `/config.json` だけであること（通信を試みない）、
  差し替えた接続先へ通信することを Vitest で固定。
- 認証: 未認証で保護 URL を直接開いた際に、セッション応答を保留したまま `ホーム` が
  存在しないことを確認してからログイン画面になること。認証済みで内容が出ること。
  `ensureQueryData` を 2 回呼んでもセッション取得が 1 回であること。
- エラー処理: 401 でログインへ、5xx でトースト、未分類も表示されること。
  401 / 4xx を再試行せず 5xx を 3 回試すこと。
- フォーム: クライアント検証で送信が止まること、契約のフィールドエラーが項目に出ること、
  対応の取れないエラーがフォーム全体のエラーとして残ること。
- `make build-web` が成果物を静的配信し、`config.json` の差し替えがその成果物から
  読めることを `verify:config` で確認する。アプリが実際にその URL へ通信することは
  `src/app/shell.test.tsx` が担当する（`docs/web.md` に記載）。

## レビュー指摘の修正 (2026-09-06)

[PR #8 のレビュー](https://github.com/sudabon/monorepo_project_template/pull/8#issuecomment-5559025753)
の Should Fix 5 件を修正した。

- **S-1** テストファイルが型検査の対象外だった。`tsconfig.test.json` を追加して
  `src` 全体を対象にし、`tsconfig.json` から参照する。アプリ側の
  `tsconfig.app.json` は Node の型を持たないまま残し、アプリコードが Node API へ
  手を伸ばせない状態を保つ。テストだけ `types` に `node` を足している。
- **S-2** `/auth/session` と `/auth/login` が実行時設定を通っていなかった。
  `runtimeConfigSchema` に `authBaseUrl` を必須で追加し、`fetchSession` /
  `sessionQueryOptions` / `LoginPage` が設定から組み立てるようにした。既定値は
  持たせない。片方だけ差し替えられる状態を残すと、接続先が割れたまま起動できてしまう。
- **S-3** 422 がフィールドエラーとトーストの二重表示になっていた。`classifyError` に
  `validation` を追加し、`MutationCache.onError` は
  `meta: { formHandlesValidation: true }` が付いたミューテーションの検証エラーだけを
  抑止する。meta の無い 422 と、meta がある場合の 5xx はこれまでどおり通知する
  （spec の「握りつぶしてはならない」を維持するため）。
- **S-4** `itemInputSchema` に契約の最大長がなかった。name 100 / description 2000 を
  追加し、境界値のテストを足した。
- **S-5** Suspense 禁止の検査がファイル名の直書きリストだった。`src` を走査する
  方式に変え、`routeTree.gen.ts` とテストだけ除外する。

### 検証の証跡

- S-1: 修正前は `src/config/schema.test.ts` に `const zzTypeError: number = "..."`
  を入れても `make typecheck-web` が exit 0 だった。修正後は
  `src/config/schema.test.ts(42,7): error TS2322` で失敗する。
- S-5: 修正前は `Suspense` を使うページを追加しても `2 tests passed` だった。
  修正後は `pages/ZzScratchPage.tsx uses Suspense` で失敗し、違反ファイル名が出る。
- S-3: `formHandlesValidation` 付きの 422 で通知が呼ばれないこと、meta の無い 422 と
  meta ありの 5xx では呼ばれることを 3 ケースで固定した。
- S-2: 別オリジンの設定でログインし、発行された URL がすべて
  `https://staging.example/` 始まりであることを確認するテストを追加した。
  同一オリジンの `/auth/...` が 1 本でも出れば失敗する。
- 修正後に `make check` が全通過。Vitest は 12 ファイル / 55 テスト。

## 制約

- 未解決の `TODO(template)` は 4 件。`modal.tsx` の背景スクロール固定と opener 消失時の
  フォーカス復帰、`toast.tsx` のフォーカス移動と `aria-live="assertive"`。design の
  Risks に挙げたアクセシビリティの積み残しであり、案件側で埋める前提。
- `test/renderApp.tsx` は `bootstrap.tsx` のプロバイダ構成を複製している。骨格を変えた
  ときに両方を直す必要がある（レビューの申し送り N-1）。
- E2E は tasks.md の指定どおり対象外。`test-plan.md` は追加していない。
