# SPA の実行時設定と静的配信

`apps/web` の成果物は環境に依存しない。API の接続先は起動時に読む `/config.json` だけが持つ。

```json
{ "apiBaseUrl": "/api", "authBaseUrl": "/auth" }
```

`apiBaseUrl` は業務 API、`authBaseUrl` はログインとセッション確認の接続先。どちらも BFF が
受けるが、片方だけ差し替えられると「API は別オリジン、ログインは SPA 自身のオリジン」に
割れるため、両方とも必須にしている。既定値は持たせない。

同一オリジン配信（CloudFront で SPA と BFF を同じホストにする）では `/api` と `/auth` の
ままでよい。接続先を変えるときは、ビルドし直さず `config.json` の 2 つを揃えて差し替える。

## CloudFront のキャッシュ

Terraform はテンプレートに含めない。案件の CDN 設定で次を守る。理由がない設定は、壊れたときに原因が分からない。

| オブジェクト | Cache-Control | 理由 | 守られないと起きること |
| --- | --- | --- | --- |
| `/index.html` | `no-cache` | デプロイ後もエントリ HTML が新しいアセット名を指すようにする | 古い HTML が消えたハッシュ付き JS を参照し、画面が白くなる |
| 内容ハッシュ付きアセット（Vite の `assets/*`） | `public, max-age=31536000, immutable` | ファイル名が内容と一対一なので再検証が不要 | 再検証が増えて配信が遅くなる。短くしても正しさは保たれるが、デプロイのたびに帯域を無駄にする |
| `/config.json` | `no-cache` | 接続先の差し替えを再ビルドなしで届ける。[ADR 0009](adr/0009-runtime-config-fetch.md) | ハッシュ付きアセットと同じ長期キャッシュにすると、古い `apiBaseUrl` のまま通信する。**アセットと同じ扱いにしない** |

存在しないパス（直接開いた `/items/…` やブラウザリロード）は、S3 の 403/404 を CloudFront のカスタムエラーで **`/index.html` に 200 で fallback** する。SPA のクライアントルータがパスを解釈するため。200 以外だとアセットや API の失敗と区別できず、クローラもエラーとみなす。`/api/*` と `/auth/*` はこの fallback の対象にしない（BFF オリジンへルーティングする。[bff.md](bff.md)）。

## 差し替えの確認

```sh
pnpm --filter @monorepo-project-template/web build
# dist/config.json の apiBaseUrl を書き換えて静的配信する
pnpm --filter @monorepo-project-template/web verify:config
```

`verify:config` はビルド済み `dist/` を一時配信し、`config.json` の差し替えが
その成果物から読めることを確認する。アプリケーションが実際にその URL へ通信することは
Vitest（`src/app/shell.test.tsx`）で確認する。
