# BFF の実装規約

- ブラウザ向けの認証はここで完結する。セッション発行・確認・失効と CSRF 検証を
  BFF が行う。`apps/api` にログインや Cookie 検証を持ち込まない。
- `internal/session`: セッションの発行・取得・失効。本番は PostgreSQL。
  識別子だけを `session_id` Cookie（httpOnly / Secure / SameSite=Lax）に入れる。
- `internal/identity`: 資格情報の検証。テンプレート初期値はデモユーザのみ。
- `internal/handler`: Echo の入口。リクエスト ID、クライアント由来 `X-User-ID` の
  削除、セッション読み込み、状態変更メソッドへの CSRF を `e.Use` で一律に掛ける。
  ルートごとの CSRF opt-in は禁止。`POST /auth/login` だけは発行前のため対象外。
- `internal/proxy`: `/api/*` を backend へ転送する。outbound で `X-User-ID` を
  削除してからセッションのユーザ ID を設定し、`X-Request-ID` を付ける。
- `cmd/bff`: 設定と DI。`cmd/migrate` は goose（version テーブル
  `bff_goose_db_version`）。起動時に自動 migrate しない。
- `internal/testdb`: 統合テスト専用。production からの依存は禁止。

Cookie と同一オリジン配信の前提は [docs/bff.md](../../docs/bff.md)。
ストア選定は [ADR 0004](../../docs/adr/0004-bff-session-store.md)、
転送ヘッダは [ADR 0005](../../docs/adr/0005-bff-identity-forwarding.md)。

検証: `make test-go` が BFF を含む。HTTP の認証・CSRF・転送はメモリストアの
統合テスト、PostgreSQL ストアは testdb。最後に `make check`。
