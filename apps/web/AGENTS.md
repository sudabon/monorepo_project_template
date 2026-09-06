# SPA（`apps/web`）の骨格規約

## 実行時設定

- ビルド時に環境値を埋め込まない。`import.meta.env` の `VITE_*` も使わない。
- 起動時に `/config.json` を 1 回取得し、zod で検証してから React を描画する。
- 取得失敗・スキーマ不一致はエラー画面にする。既定値で動かさない。
- `apiBaseUrl` は公開してよい接続先だけを書く。秘密情報は SPA に置かない。
- 配信時は `config.json` を CloudFront で `Cache-Control: no-cache` にする。詳細は [docs/web.md](../../docs/web.md)。

## 認証

- ログイン状態のソースは `GET /auth/session` のみ。SPA にフラグを持たない。
- 保護ルートの判定は TanStack Router の `beforeLoad` で行う。描画後に消す実装は禁止。
- セッション結果は TanStack Query のキャッシュ経由で読む。画面ごとに再問い合わせしない。
- CSRF トークンはセッション応答の `csrfToken` を状態変更リクエストの `X-CSRF-Token` に付ける。

## サーバ状態とエラー

- `QueryClient` の既定値は `src/query/createQueryClient.ts` にだけ置く。
- 401 はセッション切れとしてログインへ。5xx はトースト。それ以外も握りつぶさず表示する。
- ローディングは TanStack Query の `isPending` を画面内で扱う。`React.Suspense` と混在させない。

## UI

- ボタン・入力・モーダル・テーブル・トーストは `src/components/ui/` の内製コードを使う。
- UI コンポーネントライブラリを依存に追加しない。
- ダークモードは実装しない。

## フォーム

- クライアント検証は react-hook-form + zod。契約の制約に寄せたスキーマの例は `src/forms/itemInputSchema.ts`。
- サーバの `validation_error.errors[].field` は `mapServerErrors` でフォーム項目へ変換する。対応しない項目はフォーム全体のエラーにする。

## テスト

- 骨格の振る舞いは Vitest + Testing Library で確認する。Playwright は Phase 5 のスモークに任せる。
- アクセシブルネームに依存する。文言変更は仕様変更として扱う。
