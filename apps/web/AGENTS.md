# SPA（`apps/web`）の骨格規約

このファイルがフロントエンド作業の正。ルートの [AGENTS.md](../../AGENTS.md) は案内だけ。API / BFF のファイルは開かない。

生成物 `src/routeTree.gen.ts` は手編集しない。ルートファイルを直して Vite プラグインに再生成させる。

## 実行時設定

- ビルド時に環境値を埋め込まない。`import.meta.env` の `VITE_*` も使わない。
- 起動時に `/config.json` を 1 回取得し、zod で検証してから React を描画する。
- 取得失敗・スキーマ不一致はエラー画面にする。既定値で動かさない。
- `apiBaseUrl`（業務 API）と `authBaseUrl`（ログイン・セッション確認）は両方必須。
  片方だけ差し替えると接続先が割れるため、既定値は持たせない。
- 公開してよい接続先だけを書く。秘密情報は SPA に置かない。
- 配信時は `config.json` を CloudFront で `Cache-Control: no-cache` にする。詳細は [docs/web.md](../../docs/web.md)。

## 認証

- ログイン状態のソースは `authBaseUrl` の `/session` のみ。SPA にフラグを持たない。
- 保護ルートの判定は TanStack Router の `beforeLoad` で行う。描画後に消す実装は禁止。
- セッション結果は TanStack Query のキャッシュ経由で読む。画面ごとに再問い合わせしない。
- CSRF トークンはセッション応答の `csrfToken` を状態変更リクエストの `X-CSRF-Token` に付ける。

## サーバ状態とエラー

- `QueryClient` の既定値は `src/query/createQueryClient.ts` にだけ置く。
- 作成・更新・削除のあとキャッシュを無効化するキーは `api-client` の `itemKeys` から取る。
  `queryKey: ['items']` のような文字列リテラルで組み立てない。
- 401 はセッション切れとしてログインへ。5xx はトースト。それ以外も握りつぶさず表示する。
- フォームが契約のフィールドエラーを描くミューテーションは `meta: { formHandlesValidation: true }`
  を付ける。付けないと 422 が項目とトーストの二重表示になる。5xx は meta があっても通知する。
- ローディングは TanStack Query の `isPending` を画面内で扱う。`React.Suspense` と混在させない。

## UI

- ボタン・入力・複数行入力・モーダル・テーブル・トーストは `src/components/ui/` の内製コードを使う。
- UI コンポーネントライブラリを依存に追加しない。
- ボタンの見た目を `<button>` 以外で使うときは `buttonClasses` を使う。クラス文字列を写さない。
- ダークモードは実装しない。

## フォーム

- クライアント検証は react-hook-form + zod。契約の制約に寄せたスキーマの例は `src/forms/itemInputSchema.ts`。
  文字数と禁止文字は `api-client` が公開する生成制約を参照する。契約を変えたら `make gen` してからスキーマとテストを直す。
- ページが共有する文言は `src/pages/<resource>/messages.ts` に置く。ページ実装同士を import しない。
- サーバの `validation_error.errors[].field` は `mapServerErrors` でフォーム項目へ変換する。対応しない項目はフォーム全体のエラーにする。
- 項目のエラーは入力欄と `aria-describedby` で結ぶ。隣に置くだけでは関連付かない。

## テスト

- 骨格の振る舞いは Vitest + Testing Library で確認する。ブラウザ操作は `tests/e2e` の Playwright に任せる。
- アクセシブルネームに依存する。UI 文言の変更は仕様変更として扱う。変更する前に
  該当する OpenSpec change の `test-plan.md` を先に更新し、E2E の期待結果と揃えてから
  画面を直す。
