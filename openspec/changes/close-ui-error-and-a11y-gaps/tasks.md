## 1. 削除失敗時の未処理拒否

- [ ] 1.1 削除が失敗する状況（DELETE が 500）を再現する単体テストを追加し、修正前に未処理の Promise 拒否が 1 件発生することを確認する
- [ ] 1.2 `ItemDetailPage` の `void confirmDelete();` を拒否を受け取る形に変える（design D1）
- [ ] 1.3 1.1 のテストで、未処理の Promise 拒否が 0 件になり、失敗の通知が 1 回だけ表示され、確認ダイアログが開いたままで、対象が一覧に残ることを確認する
- [ ] 1.4 削除が成功する既存テストが変更なしで通ることを確認する

## 2. ボタンの見た目の共有

- [ ] 2.1 `components/ui/button.tsx` から `buttonClasses` を export し、`Button` 自身がそれを使うように変更する（design D2）
- [ ] 2.2 `ItemListPage` の「新規作成」と `ItemDetailPage` の「編集」のクラス文字列を `buttonClasses()` に置き換える
- [ ] 2.3 ボタンの見た目を表すクラス文字列が `button.tsx` 以外に存在しないことを grep で確認する
- [ ] 2.4 リンクとボタンが同じクラス集合（フォーカス表現と無効表現を含む）を持つことを単体テストで確認する
- [ ] 2.5 `buttonClasses` から `focus-visible:` を一時的に取り除くと 2.4 のテストが失敗することを確認し、確認後に戻す

## 3. 複数行入力とエラーの関連付け

- [ ] 3.1 `components/ui/textarea.tsx` を追加し、`label` と `invalid` の扱いを `Input` と揃える（design D3）
- [ ] 3.2 `ItemForm` の `description` を `Textarea` に置き換え、改行を含む内容を入力・保存・再表示できることを単体テストで確認する
- [ ] 3.3 `Input` と `Textarea` にエラー要素の id を受け取る経路を追加する（design D4）
- [ ] 3.4 `ItemForm` のエラー `<p>` に id を与え、入力欄のアクセシブルな説明としてエラー文が取得できることを単体テストで確認する
- [ ] 3.5 クライアント側検証のエラーとサーバ側検証のエラーの両方で 3.4 が成り立つことを確認する
- [ ] 3.6 `components/ui/ui.test.tsx` に `Textarea` の分を追加する

## 4. 記録と検証

- [ ] 4.1 `apps/web/AGENTS.md` に「見た目の再利用は `buttonClasses` を使い、クラス文字列を写さない」「項目のエラーは入力欄と `aria-describedby` で結ぶ」を明記する
- [ ] 4.2 `docs/TEMPLATE_TODO.md` の UI 系 `TODO(template)` を読み直し、今回直した内容と重複していないこと、残す判断が今も妥当であることを確認する
- [ ] 4.3 E2E テストを増やしていないこと（`tests/e2e/` の `test(` が 1 件のまま）を確認する
- [ ] 4.4 `make check` と `git diff --check` が通ることを確認する
