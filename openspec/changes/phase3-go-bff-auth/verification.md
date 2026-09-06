# 実装検証記録

2026-09-06、branch `feat/phase3-go-bff-auth`。

## 自動検証の範囲

- `make check`: Go / Web 整形、go vet、`go mod tidy -diff`、依存方向 lint、Spectral、
  生成差分、実 PostgreSQL を含む Go テスト、api-client テスト、型検査、build。
  このマシンの pnpm shim が壊れているため、固定版 11.25.0 を呼ぶ既存の
  `/tmp/phase1-toolchain/bin/pnpm` を PATH に追加して実行した。
- Cookie 属性（httpOnly / Secure / SameSite=Lax / Path）、誤資格情報での Cookie 不発行、
  応答本文とヘッダへのセッション ID 非露出、セッション確認の認証済み / 未認証、
  ログアウト後の Cookie 再利用拒否、未認証 401 かつ非リダイレクト、改竄 Cookie の 401、
  CSRF の未添付 / 不一致 / 一致、`X-User-ID` の詐称破棄と `X-Request-ID` の引き継ぎ。
  いずれも `apps/bff/internal/handler` の統合テストで固定。
- `curl` による login → 保護リソース → logout → 401 → CSRF 403 の一連は
  `curl_test.go` が実プロセス相手に実行する。手作業の記録に頼らない。
- セッションストアの契約はメモリと PostgreSQL の両方を同じ `exerciseStore` に通し、
  スライディング延長と絶対上限を注入クロックで検証する。
- CSRF が per-route opt-in になっていないことを `csrf_mw_test.go` がソース検査で固定する。

## レビュー指摘の修正 (2026-09-06)

[PR #7 のレビュー](https://github.com/sudabon/monorepo_project_template/pull/7#issuecomment-5558397869)
の Should Fix 5 件を修正した。

- **S-1** `loadSession` が `session.ErrNotFound` だけを「未ログイン」として扱い、
  それ以外のストアエラーを 5xx にするようにした。修正前は接続断でも 401 を返し、
  DB 停止中は全ユーザがログアウト扱いになっていた。回帰テスト
  `TestStoreOutageIsServerErrorNotUnauthenticated` を追加。
- **S-2** `proxy` が転送前に `Cookie` と `Authorization` を削除するようにした。
  修正前は backend が `Cookie: session_id=...` とクライアント由来の
  `Authorization` を受け取っていた。`TestBrowserCredentialsAreNotForwardedToBackend`
  で backend 側の受信ヘッダを検証する。
- **S-3** `POST /auth/login` に 8KiB の `http.MaxBytesReader` を入れ、超過を 413 /
  `payload_too_large` にした。修正前は 8MB の本文が最後まで読まれて 401 を返していた。
  `TestLoginBodyIsCapped` を追加。
- **S-4** `docs/bff.md` の curl 例のパスワードを `secret` から `demo` に直した。
  `make run-bff` と `docs/RUNBOOK.md` の値と食い違い、手順どおりに叩くと 401 だった。
- **S-5** `internal/platform` の重複を `packages/go-platform` の共有モジュールに寄せた。
  `logging.go` は 2 module にバイト単位で同一のコピーがあり、`server.go` と
  `database.go` はコピー時に根拠コメントが落ちていた。判断は
  [ADR 0006](../../../docs/adr/0006-shared-go-platform-module.md)。
  接続プールの見積もりを「API と BFF の合計タスク数 × 10」に更新した。
  API 側にしかなかった SIGTERM の drain / 待機上限テストが共有モジュールに移り、
  BFF の `Serve` も同じテストで覆われるようになった（申し送り N-2 も解消）。

### 検証の証跡

- S-1 / S-2 / S-3 は修正前に再現を実測してから直した。ストア障害は 401、
  backend は `Cookie` と `Authorization` を受信、8MB の login 本文は 413 ではなく 401。
- go-arch-lint の `canUse: [platform]` が空振りしていないことを確認した。
  `apps/bff/internal/proxy` に echo の import を一時的に足すと
  `Component proxy shouldn't depend on github.com/labstack/echo/v4 ... proxy.go:8`
  で失敗する（exit 1）。共有モジュールだけを許可し、vendor 全体は開いていない。
- 修正後に `make check` が全通過。`go-platform` も fmt-check / lint / test の対象に入る。

## 制約

- 未解決の `TODO(template)` は 2 件。`identity/static.go` のデモユーザ差し替えと、
  `session/postgres.go` の期限切れセッションの定期削除。いずれも案件側で埋める前提。
- `packages/go-platform` は未公開モジュール。各 `go.mod` の `replace` と `go.work` で
  解決する。案件がコンテナを個別ビルドする場合も `replace` があれば動く。
- `internal/testdb` は API と BFF で別実装のまま。各 module の `migrations` に依存し、
  テスト専用のため共有していない。
