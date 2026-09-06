## 1. 一覧画面

- [x] 1.1 一覧ルートの検索条件・ページ位置を zod スキーマ付きの型付き search params として定義し、不正な URL パラメータが既定値にフォールバックすることをユニットテストで確認する
- [x] 1.2 一覧画面を実装し、`packages/api-client` の `queryOptions` 経由でデータを取得して表示されることを確認する
- [x] 1.3 検索条件の変更が URL の書き換えとして行われ、コンポーネント state に二重管理されていないことをコードレビューで確認する
- [x] 1.4 ページネーション UI を実装し、ページ移動で URL が変わり対応する内容が表示されることを Testing Library で確認する
- [x] 1.5 該当データがない場合に、エラーではなく該当なしの表示になることを Testing Library で確認する

## 2. 詳細画面

- [x] 2.1 詳細ルートを実装し、一覧の項目から遷移して当該項目の内容が表示されることを Testing Library で確認する

## 3. 作成・編集フォーム

- [x] 3.1 作成フォームを実装し、Phase 4 の検証スキーマとサーバ検証エラーのマッピング共通処理を実際に接続する
- [x] 3.2 作成成功後に一覧のクエリを無効化して再取得し、作成した項目が一覧に現れることを Testing Library で確認する
- [x] 3.3 編集フォームを実装し、更新成功後に詳細の内容が更新されることを Testing Library で確認する
- [x] 3.4 サーバ検証エラーが該当フォーム項目に表示され、リソースが作成・更新されないことを Testing Library で確認する
- [x] 3.5 送信中の多重送信が抑止されることを Testing Library で確認する

## 4. 削除

- [x] 4.1 Phase 4 の内製モーダルを用いた削除確認を実装し、`window.confirm` を使っていないことをコードで確認する
- [x] 4.2 確認を承認すると削除され一覧から消えること、取り消すと削除されないことを Testing Library で確認する

## 5. E2E 実行環境とシード

- [x] 5.1 API・BFF・SPA と DB を一括起動する `make` ターゲットを追加し、ローカルで起動できることを確認する
- [x] 5.2 `playwright.config.ts` の `webServer` から 5.1 のターゲットを呼ぶ設定を追加し、`make` 経由の E2E 実行で起動から実行まで通ることを確認する
- [x] 5.3 シード fixture `authenticated-user`(ログイン済み状態)を `tests/e2e/fixtures/` に追加し、テスト本体でログイン操作を書かずに認証済みで開始できることを確認する
- [x] 5.4 シード fixture `sample-resources-multi-page`(複数ページに跨る件数のサンプルリソース)を追加し、2 ページ目が存在する状態になることを確認する
- [x] 5.5 追加した fixture 名と作られる状態の対応を `tests/e2e/fixtures/README.md` に記録する
- [x] 5.6 一覧・詳細・フォーム・削除確認の Page Object を `tests/e2e/pages/` に作成し、セレクタがテスト本体に漏れていないことを確認する(ロケーターは getByRole / getByLabel / getByText を優先)

## 6. E2Eテスト実装タスク(必須)

test-plan.md の TP-001〜TP-007 を、単一のスモークテストの `test.step` として実装する。
テストには `@phase5-fullstack-vertical-slice` と `@TP-001`〜`@TP-007` のすべてのタグを付与する。
実装規約は `docs/E2E_CONVENTIONS.md` に従う。

- [x] 6.1 スモークテストのファイルを作成し、`{ tag: ['@phase5-fullstack-vertical-slice', '@TP-001', ..., '@TP-007'] }` を付与して、タグで絞り込み実行できることを確認する
- [x] 6.2 TP-001: `authenticated-user` fixture で開始し、一覧から作成フォームで有効な内容を送信して一覧に現れることを `test.step` で検証する
- [x] 6.3 TP-002: 作成した項目の詳細へ遷移し、内容が表示されることを `test.step` で検証する
- [x] 6.4 TP-003: 詳細から編集して更新し、詳細に更新後の内容が反映されることを `test.step` で検証する
- [x] 6.5 TP-004: 削除を実行して確認を取り消し、一覧に残ることを `test.step` で検証する
- [x] 6.6 TP-005: 削除を実行して確認を承認し、一覧から消えることを `test.step` で検証する
- [x] 6.7 TP-006: `sample-resources-multi-page` を前提に、検索条件を変更してリロードし条件が保持されることを `test.step` で検証する
- [x] 6.8 TP-007: 2 ページ目へ移動してリロードし、2 ページ目が表示されることを `test.step` で検証する
- [x] 6.9 `page.waitForTimeout` / `page.locator` / XPath を使っていないことを確認し、テストを 3 回連続実行してフレークしないことを確認する

## 7. CI と完了条件の検証

- [x] 7.1 `.github/workflows/web.yml` にスモーク実行を追加し、CI で緑になることを確認する
- [x] 7.2 `scripts/check-test-plan.sh` が `@phase5-fullstack-vertical-slice` タグ付きテストを検出して成功することを確認する
- [x] 7.3 `make check` が全て通ることを確認する
- [x] 7.4 一覧画面で検索条件を変えた後にブラウザをリロードし、条件が保持されることを実機で確認する
- [x] 7.5 UI 文言の変更を仕様変更として扱う運用(test-plan への反映が先)を `apps/web/AGENTS.md` に明記する
- [x] 7.6 妥協した実装の `// TODO(template):` 一覧を報告に含める
