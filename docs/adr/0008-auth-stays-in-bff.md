# 0008: ブラウザ向け認証は BFF に閉じる

- 状態: 採用
- 日付: 2026-09-06

## 判断

ログイン・セッション発行・CSRF・ログアウトは `apps/bff` だけで完結する。SPA にアクセストークンを渡さない。`apps/api` にログインや Cookie 検証を持ち込まない。API は BFF が付けた `X-User-ID` を、ネットワーク境界を信頼して受け取る。

セッションの保存先は [ADR 0004](0004-bff-session-store.md)、転送ヘッダは [ADR 0005](0005-bff-identity-forwarding.md)、Cookie と同一オリジン前提は [docs/bff.md](../bff.md)。

## 理由

トークンを SPA に置くと XSS の影響面が広がる。認証を API にも分散すると、Cookie 属性・CSRF・失効の実装が 2 箇所になり、抜けが出る。BFF に閉じるとブラウザ向けの境界が 1 つになり、API は内部サービスとして小さく保てる。

## 代替案

| 案 | 結果 |
| --- | --- |
| SPA が JWT を保持し API に直接付ける | XSS でトークンが漏れる。失効が難しい。却下 |
| API 自身が Cookie セッションを持つ | 認証と業務 API が同じプロセスに混ざる。ブラウザ公開面が API になる。却下 |
| 共有秘密ヘッダで API が BFF を検証する | 初期値のネットワーク前提では過剰。前提が崩れた案件で [ADR 0005](0005-bff-identity-forwarding.md) に従って足す |

参考: Phase 3 `design.md` D1–D4、[docs/bff.md](../bff.md)。
