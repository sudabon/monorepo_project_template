# 実装検証記録

2026-09-06、branch `feat/phase6-ops-docs-template`。

本 change の成果物は文書であり、自動テストが守れる範囲が小さい。したがって手動検証の
記録がほぼ唯一の証跡になる。以下は実際に実行した内容であり、実行していないものは
「未検証」として明示する。

## 複製ドライラン（tasks 1.5 / 1.6）

`feat/phase6-ops-docs-template` を別ディレクトリへ `git clone` し、README の
チェックリストを適用した。置き換えた値は次のとおり。

| チェックリスト | テンプレートの値 | 適用した値 |
| --- | --- | --- |
| 1 Go モジュールパス | `github.com/sudabon/monorepo_project_template` | `github.com/acme/orders_platform` |
| 2 npm パッケージ名 | `@monorepo-project-template/*` / `monorepo-project-template` | `@acme-orders/*` / `acme-orders` |
| 5 テンプレート専用ガード | `"templateRepo": true` | キーごと削除 |
| 作者向けファイル | `monorepo-project-template-instructions.md` | 削除 |

その後 `go mod tidy` を 3 モジュール（`apps/api` / `apps/bff` / `packages/go-platform`）で
実行し、`make fmt` を通した。`compose.yaml` の `name:` は識別子の置換で
`acme-orders` に変わった。

### 結果

- README の全文検索がヒット 0 件（`rg` の exit 1）。
- 追加した `test ! -f monorepo-project-template-instructions.md` が成功。
- `make setup && make check` が成功（Vitest 15 ファイル / 72 テスト）。
  ローカルの Compose が既に 55432 を使っているため、CI と同じ方式で
  `TEST_DATABASE_URL` を渡した。
- 複製先で `templateRepo` を削除した状態に `test-plan.md` を置き、タグ付きテストを
  用意せずに `make check-test-plan` を実行すると失敗する。

```
::error::@phase6-ops-docs-template タグ付きの E2E テストが tests/e2e/ にありません
make: *** [check-test-plan] Error 1
```

  `test-plan.md` を消すと `test-plan.md なし（E2E 不要）。skip` で成功に戻る。
  テンプレート側では `templateRepo=true: E2E 計画ガードを skip` となる。
  ガードが複製先でだけ有効になることを、両方向で確認した。

### このドライランが**覆っていない**範囲

- チェックリスト 3（サンプルリソース名の置き換え）は設計作業であり、機械的な置換に
  馴染まないため適用していない。`items` / `Item` は README の検索パターンに含まれないので、
  未適用でも検証は通る。案件の最初のリソースに置き換える判断は複製先で行う。
- チェックリスト 4（CI のシークレット名）は、現状 workflow に `secrets.*` 参照が無く
  置き換える対象が存在しない。DB 認証の `template` は変更していない。
- チェックリスト 6（認証の差し替え）は案件の IdP が要るため適用していない。
  機械的に検出できないことを README に明記した。
- `api/openapi.yaml` の `info.title` は識別子の検索に掛からないため残る。README の
  「合わせて次も直す」に記載済み。

## 未検証

- **tasks 7.2 の「README のみを読んで 30 分以内に立ち上げられる」**: 所要時間は読み手に
  依存し、本記録では再現できない。上記ドライランは README の手順に沿っているが、
  時間の計測は行っていない。

## レビュー指摘の修正 (2026-09-06)

[PR #10 のレビュー](https://github.com/sudabon/monorepo_project_template/pull/10#issuecomment-5559936687)
の Should Fix 5 件を修正した。

- **S-1** チェックリストに 6 項目目「認証の実装」を追加した。デモユーザのまま本番に
  出ると何が起きるかを同じ行に書いた。
- **S-2** `docs/TEMPLATE_TODO.md` を追加し、`TODO(template)` 7 件を「案件で必ず解消」
  2 件と「テンプレートに残す」5 件に仕分けた。README のドキュメント表から参照する。
  棚卸し用の `rg` も置いた。
- **S-3** `packages/api-client/AGENTS.md` と `packages/go-platform/AGENTS.md` を追加し、
  ルート `AGENTS.md` の案内表に 2 行と契約・E2E の行を足した。`packages/` を触る
  エージェントが読むものが無い状態を解消した。
- **S-4** 作者向けファイルの削除確認を `test ! -f` として README の検証手順に足した。
  内容検索では検出できないため、目視依存になっていた。
- **S-5** 本ファイル。tasks 1.5 / 1.6 を実際に複製して実行した記録に置き換えた。

## 制約

- README の検証コマンドは `docs/**` を除外する。ADR は履歴として識別子を残す判断だが、
  案件で書き換える文書も同じ除外に入る（レビューの申し送り N-1）。
- `docs/UPGRADING.md` の「実施記録」は全節が未記入。最初のメジャー更新時に書式が
  機能するかが分かる。
