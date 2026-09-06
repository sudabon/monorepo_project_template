# SPA の実行時設定と静的配信

`apps/web` の成果物は環境に依存しない。API の接続先は起動時に読む `/config.json` だけが持つ。

```json
{ "apiBaseUrl": "/api" }
```

同一オリジン配信（CloudFront で SPA と BFF を同じホストにする）では `/api` のままでよい。
接続先だけを変えるときは、ビルドし直さず `config.json` を差し替える。

## CloudFront のキャッシュ

`config.json` は CloudFront（または同等の CDN）で **`Cache-Control: no-cache`** にする。
HTML / JS / CSS はハッシュ付きで長くキャッシュしてよいが、設定ファイルまでキャッシュすると
接続先の差し替えが届かない。Phase 6 の配信設定と揃える。

## 差し替えの確認

```sh
pnpm --filter @monorepo-project-template/web build
# dist/config.json の apiBaseUrl を書き換えて静的配信する
pnpm --filter @monorepo-project-template/web verify:config
```

`verify:config` はビルド済み `dist/` を一時配信し、`config.json` の差し替えが
その成果物から読めることを確認する。アプリケーションが実際にその URL へ通信することは
Vitest（`src/app/shell.test.tsx`）で確認する。
