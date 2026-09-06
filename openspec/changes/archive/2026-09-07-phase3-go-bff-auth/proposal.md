## Why

SPA にトークンを持たせると、XSS が起きた瞬間に認証情報が漏れる。ブラウザに渡してよいのは httpOnly Cookie のセッションだけであり、その判断はアプリケーションごとにブレさせたくない。認証を BFF に閉じ込め、SPA 側には「認証済みかどうか」しか見えない構造をテンプレートの初期値にする。

Cookie セッションを使う以上 CSRF 対策は必須であり、後から入れると全ての更新系エンドポイントに手が入る。認証の実装と同時に組み込む。

## What Changes

- `apps/bff` に Go / Echo の BFF を実装する。認証はここで完結させる
- セッションを httpOnly / Secure / SameSite=Lax の Cookie で保持する。SPA 側にトークンを一切渡さない
- ログイン / ログアウト / セッション確認のエンドポイントを追加する
- CSRF 対策を実装する。トークンを持たない更新系リクエストを拒否する
- backend への転送時に、認証済みユーザ情報とリクエスト ID を付与する（リクエスト ID のヘッダ名は Phase 2 で決めたものに従う）
- 未認証時は 401 を返す。リダイレクトは返さない（遷移は SPA 側で処理するため）

## Capabilities

### New Capabilities

- `session-auth`: BFF が提供する認証の振る舞い。Cookie セッションの発行と失効、セッション状態の確認、未認証リクエストの扱い、CSRF 保護、backend への認証済みコンテキストの転送を含む。

### Modified Capabilities

なし。`sample-resource-api` と `service-runtime` の要求は変更しない。BFF は API の前段に立つ別のコンポーネントであり、その振る舞いを新たに定義する。

## Impact

- **新規ファイル**: `apps/bff/` 配下（Echo サーバ、認証ハンドラ、セッションストア、CSRF ミドルウェア、転送プロキシ）、`apps/bff/AGENTS.md`
- **変更ファイル**: `go.work`（`apps/bff` の登録が Phase 0 で済んでいなければ追加）、`Makefile`（BFF の起動ターゲット）、`.github/workflows/go.yml`（BFF のテスト追加）
- **依存追加**: セッションストアの実装、CSRF 対策の実装。いずれも選定して ADR に記録する
- **後続への影響**: Phase 4 の SPA は BFF の 401 応答と CSRF トークンの取得方法に依存する。Phase 5 の画面操作はすべて BFF を経由する
