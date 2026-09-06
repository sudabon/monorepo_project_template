# 実装検証記録

2026-09-06、branch `feat/phase2-go-backend-reference`。

## 成功した検証

- `make check`: Go / Web 整形、go vet、依存方向 lint、Spectral、生成差分、
  Go 実 DB テスト、toolchain 22 テスト、api-client 7 テスト、型検査、build。
  このマシンの pnpm shim が壊れているため、固定版 11.25.0 を呼ぶ既存の
  `/tmp/phase1-toolchain/bin/pnpm` を PATH に追加して実行した。
- `TEST_DATABASE_URL=... make test`: CI と同じ外部 DB 指定方式でも成功。
  URL 未指定の `make test-go` は Compose を起動して同じテストを実行する。
- `make migrate-up` → `make migrate-down` → `make migrate-up`: すべて成功。
- 一時コピーで domain に repository 配下の package の import を追加すると
  go-arch-lint が失敗し、`internal/domain/forbidden.go:2` を出力。
- 一時コピーで `Items.DeleteItem` を外すと build が
  `does not implement generated.ServerInterface (missing method DeleteItem)` で失敗。
- 実 API バイナリへの curl: POST / GET / PUT / DELETE、404、ページ取得、
  複数フィールドの 422 と永続化されないことを確認。
- 実 API プロセスを終了・再起動し、作成済み item が一致することを確認。
- 実 PostgreSQL コンテナ停止中の shallow=200、deep=503 を確認。DB は再起動済み。
- Go subprocess テストで実 SIGTERM の drain / 待機上限を検証。
- レビュー指摘の応答書き込み失敗を再現し、ERROR ログと request_id を検証する
  テストの失敗を確認後に修正。再実行と scoped 再レビューで解消を確認。
- [PR #6](https://github.com/sudabon/monorepo_project_template/pull/6) の
  [Go CI](https://github.com/sudabon/monorepo_project_template/actions/runs/34022732951) が成功。
  commit `963461d` で PostgreSQL サービスを起動し、`make test`、lint、生成差分、build が
  すべて成功した。契約 / Web / E2E 計画ガード / Renovate も同じ commit で成功。
  初回 CI で検出した DB ヘルスチェックコマンドの引用符を修正後に再実行した。

2026-09-06 のユーザ承認に基づき、OpenAPI の Item / ItemInput の name / description に
NUL 禁止の pattern と 422 の説明を追加し、Go / TypeScript を再生成した。
型とエンドポイントに変更はなく、生成差分はフィールドの説明コメントのみ。
domain / HTTP / 契約バリデータの回帰テストで、修正前の失敗と修正後の成功を確認した。
POST / PUT の単一・複数フィールド拒否、他の制約との集約、永続化されないこと、既存 item が
変わらないことを検証した。通常の改行と文字列としての `\u0000` の保存・取得も成功。
E2E は tasks.md の指定に従い対象外。`test-plan.md` は追加していない。

## 制約

- NUL 入力は承認済みの追加制約として 422 で拒否する。未解決の `TODO(template)` は 0 件。
- offset pagination はリクエスト間の挿入・削除に対する snapshot 維持を保証しない。
  契約どおりの offset 方式であり、同一応答の total / items は同じ snapshot。
