# BFF の配信・Cookie・セッション

ブラウザ向けの認証は `apps/bff` で完結する。SPA にトークンを渡さず、
httpOnly Cookie のセッションだけを使う。API は BFF が付与した `X-User-ID` を
信頼する内部サービスとして扱う。

## 配信構成の前提（D6）

初期値は CloudFront（または同等の同一ホスト配信）で SPA と BFF を同一オリジンにする。

- SPA の静的ファイルをデフォルトオリジン（S3 等）に置く
- `/api/*` と `/auth/*` を BFF オリジンへルーティングする
- Cookie は `Path=/`、`SameSite=Lax`、`Secure`、`HttpOnly`

同一オリジンなら Lax でも同一サイトの fetch に Cookie が付く。CSRF トークンは
Lax に加えてセッション紐付けで検証する。

### この前提を採らない場合

SPA が `https://app.example`、BFF が `https://api.example` のように別サイトになる
と、`SameSite=Lax` ではクロスサイトの XHR/fetch に Cookie が乗らない。

その構成では次が必要になる。

- Cookie を `SameSite=None; Secure` にする
- BFF で CORS を開き `Access-Control-Allow-Credentials: true` と明示 Origin を返す
- CSRF の防御が実質トークン検証だけになる（Lax による補助が消える）

案件で別オリジンにするなら、上記の差分をまとめて変更する。Cookie だけ
`None` にすると、CORS 未設定のままブラウザが Cookie を送らない。

ローカルで HTTP のままブラウザ確認する場合、`Secure` Cookie は乗らない。
`make run-bff` は開発用に `BFF_COOKIE_SECURE=false` を付ける。本番と CI の
既定は `Secure=true` のままにする。

## セッションの有効期限

| 項目 | 初期値 | 根拠 |
| --- | --- | --- |
| アイドル | 12 時間 | 業務時間をまたいでも再ログイン過多にならない。放置端末の窓は営業日以内 |
| 絶対上限 | 24 時間 | スライディングだけだと無期限相当になるため、発行から 1 日で切る |
| 更新 | 取得のたびにアイドルを延長し、絶対上限を超えない | 操作中の切断を避けつつ、上限は `created_at` 基準で固定する |

CSRF トークンはセッション発行時に作り、セッションと同じ寿命にする。
GET のたびに回すと並行リクエストで先行側が失敗するため、初期値では回さない。

ログアウトはサーバ側の行を削除する。Cookie 削除だけに依存しない。

## ヘッダ

| ヘッダ | 役割 |
| --- | --- |
| `session_id` Cookie | セッション識別子。httpOnly / Secure / SameSite=Lax |
| `X-CSRF-Token` | 状態変更リクエスト。値は `GET /auth/session` の `csrfToken` |
| `X-Request-ID` | SPA → BFF → API の追跡。未指定なら BFF が生成 |
| `X-User-ID` | BFF が API へ付けるユーザ識別。クライアント値は破棄する |

選定理由は [ADR 0004](adr/0004-bff-session-store.md) と
[ADR 0005](adr/0005-bff-identity-forwarding.md) を参照。

## ローカル起動

```sh
make db-up
make migrate-up
make migrate-bff-up
make run-api
make run-bff
```

BFF は `http://127.0.0.1:8081`。API は `http://127.0.0.1:8080`。
開発用デモユーザは `demo` / `demo`。`POST /auth/login` は CSRF 対象外
（セッション発行前のため）。それ以外の POST/PUT/PATCH/DELETE はトークン必須。

```sh
curl -sS -D - -o /tmp/login.json -X POST http://127.0.0.1:8081/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","password":"demo"}'
# Set-Cookie の session_id を Cookie ヘッダで送り直す（Secure を HTTP で確認するため）
curl -sS -o /dev/null -w '%{http_code}\n' \
  -H "Cookie: session_id=<value>" http://127.0.0.1:8081/api/items
curl -sS -o /dev/null -w '%{http_code}\n' -X POST http://127.0.0.1:8081/auth/logout \
  -H "Cookie: session_id=<value>" -H "X-CSRF-Token: <csrfToken>"
curl -sS -o /dev/null -w '%{http_code}\n' \
  -H "Cookie: session_id=<value>" http://127.0.0.1:8081/api/items   # 401
curl -sS -o /dev/null -w '%{http_code}\n' -X POST http://127.0.0.1:8081/api/items \
  -H "Cookie: session_id=<value>" -H 'Content-Type: application/json' \
  -d '{"name":"x"}'   # 403（CSRF トークンなし）
```
