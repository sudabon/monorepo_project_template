## Why

参照実装は「動くこと」ではなく「真似されること」で評価される。SPA の参照実装に、真似されると困る欠陥が 3 つ残っている。いずれも `docs/TEMPLATE_TODO.md` に載っておらず、テストも通っていない。

1. **削除に失敗すると未処理の Promise 拒否が出る。** 実測で確認した。DELETE が 500 を返す状況を作ると、画面表示は正しい（トーストが出て確認ダイアログは開いたまま）が、`unhandledRejections: 1 / ApiError` が同時に発生する。原因は `ItemDetailPage` が非同期処理を `void` で捨てていること。`docs/RUNBOOK.md` が案内するとおり Sentry や Datadog を差し込むと、**それらは既定で `unhandledrejection` を拾うため、削除失敗 1 回がトースト（想定内）とクラッシュイベント（想定外）の 2 件として届く**
2. **リンクがボタンの見た目だけを写していて、フォーカスリングが出ない。** 一覧の「新規作成」と詳細の「編集」は `components/ui/button.tsx` のクラス文字列をコピーしているが、コピーされたのは基本クラスだけで `focus-visible:` と `disabled:` が落ちている。UI 部品を自分のコードとして持つ方針の下で、その部品を使えない場面が生まれ、コピーで回避されている
3. **2000 文字の説明が単一行入力で、エラーがフィールドに関連付いていない。** 契約は `description` に `maxLength: 2000` を与えているのに、フォームは 1 行の `<input>` を出す。`components/ui/` に複数行入力が無い。あわせて `Input` はエラー文と `aria-describedby` で結ばれておらず、支援技術が項目とエラーを対応付けられない

## What Changes

- 削除の失敗経路で未処理の Promise 拒否を発生させない。エラー通知は現在どおりグローバルな `onError` が 1 回だけ行う
- `components/ui/button.tsx` からボタンの見た目を関数として公開し、リンクなど `<button>` 以外の要素でも同じ見た目・同じフォーカス表現・同じ無効表現を再利用できるようにする。クラス文字列のコピーを消す
- `components/ui/` に複数行入力を追加し、契約が長い文字列を許す項目で使う
- `Input` と新しい複数行入力に、エラー文を `aria-describedby` で結ぶ経路を通す
- 上記 3 つを回帰テストで固定する。**E2E は増やさない**（スモーク 1 本のみという方針を維持し、Testing Library の単体テストで確認する）

## Capabilities

### Modified Capabilities

- `sample-resource-ui`: 削除の失敗時の振る舞いを要求として追加する。作成・編集で長い文字列の項目が複数行で編集できることを追加する
- `web-app-shell`: UI 部品の内製に「見た目の再利用はコピーではなく共有で行う」ことを追加する。フォーム検証に「エラーと項目が支援技術から対応付けられる」ことを追加する

### New Capabilities

なし。

## Impact

- **変更**: `apps/web/src/pages/items/ItemDetailPage.tsx`、`apps/web/src/components/ui/button.tsx`、`apps/web/src/components/ui/input.tsx`、`apps/web/src/forms/ItemForm.tsx`、`apps/web/src/pages/items/ItemListPage.tsx`
- **新規**: `apps/web/src/components/ui/textarea.tsx`
- **新規テスト**: 削除失敗時に未処理の Promise 拒否が出ないこと、リンクがボタンと同じフォーカス表現を持つこと、エラー文が項目に関連付くこと
- **依存追加**: なし
- **ドキュメント**: `docs/TEMPLATE_TODO.md` の UI 系 `TODO(template)` と重複しないことを確認する。`apps/web/AGENTS.md` に見た目の再利用の作法を追記する
