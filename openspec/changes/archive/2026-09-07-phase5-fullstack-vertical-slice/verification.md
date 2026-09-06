# 実装検証記録

2026-09-06、branch `feat/phase5-fullstack-vertical-slice`。

## 自動検証の範囲

- `make check`: Go / Web 整形、go vet、`go mod tidy -diff`、依存方向 lint、Spectral、
  生成差分、実 PostgreSQL を含む Go テスト、Vitest、型検査、build、E2E 計画ガード。
  このマシンの pnpm shim が壊れているため、固定版 11.25.0 を呼ぶ既存の
  `/tmp/phase1-toolchain/bin/pnpm` を PATH に追加し、`CI=true` を付けて実行した
  （TTY が無いと `pnpm install` が node_modules の再作成を中断するため）。
- Vitest: 一覧の URL 反映、ページ移動、空一覧、詳細遷移、作成後の再取得、編集反映、
  サーバ検証エラー、多重送信の抑止、削除の確認と取り消し、search params スキーマの
  フォールバック、`createAuthedFetch` の CSRF 付与。
- Playwright スモーク 1 本が TP-001〜TP-007 を `test.step` で通す。CI の E2E ジョブが
  `make e2e` を実行する。

## レビュー指摘の修正 (2026-09-06)

[PR #9 のレビュー](https://github.com/sudabon/monorepo_project_template/pull/9#issuecomment-5559406610)
の Should Fix 5 件を修正した。

- **S-1** CI の E2E が SPA の起動待ちで落ちていた。Vite が `server.host` 未指定で
  `localhost` に bind するため、ubuntu runner では `::1` を掴み、
  `scripts/e2e-serve.mjs` の `fetch('http://127.0.0.1:5173')` が繋がらなかった。
  `vite.config.ts` に `host: '127.0.0.1'` を明示し、API / BFF の待ち受けと表記を揃えた。
- **S-2** 検索がクライアント側の絞り込みなのに、ページャがサーバの全件数を基準に
  出ていた。「該当なし」と「次のページあり」が同時に成立していた。絞り込み中は
  ページャを出さず、絞り込みがページ内に閉じることを画面に明示する。
  契約に `q` を足すサーバ側検索は公開契約の変更にあたるため、ユーザ承認が必要で
  本 PR では行わない。`TODO(template)` にその旨を残した。
- **S-3** 詳細・編集画面が `!query.data` だけを見て「見つかりません」と表示していた。
  500 や通信断でも同じ文言だった。`itemLoadMessage` で 404 とそれ以外を分ける。
- **S-4** 「作成されない」ことの検証が `const created = false` に対する
  `expect(created).toBe(false)` で恒真だった。モックの保存配列を一覧が読む形にし、
  送信後に一覧へ戻って項目が無いことを確認する。
- **S-5** `web.yml` に Go のパスを足したため、Go だけの変更で web ジョブ（Vitest・
  型検査・build）まで起動していた。E2E を `e2e.yml` に分離し、`web.yml` のフィルタを
  元に戻した。path フィルタはワークフロー単位なので、ジョブを分けないと分離できない。

### 検証の証跡

- S-1: ローカル（macOS）では `localhost` が IPv4 に解決されるため
  `TCP 127.0.0.1:5173 (LISTEN)` となり、この不整合は再現しない。CI のログでは
  DB が 2 秒で healthy、API と BFF が 21 秒で起動した後、SPA だけが 120 秒待って
  `Timed out waiting for http://127.0.0.1:5173` で落ちていた。修正の確認は CI で行う。
- S-2: 修正前は total=21・1 ページ目に `Widget 1`〜`20`・検索語 `Widget 21` で
  「該当なし」表示と「次のページ」有効が同時に成立した。同じ条件の回帰テストを
  `items.test.tsx` に追加し、絞り込み中はページャが存在しないことを固定した。
- S-3: 修正前は詳細の GET が 500 でも `サンプルリソースが見つかりません` と表示した。
  404 と 500 の 2 ケースを回帰テストで固定した。
- S-4: 修正後のテストが実際に効くことを、POST ハンドラに保存を混ぜて確認した。
  `Unable to find an element with the text: 該当するサンプルリソースはありません`
  で失敗する。
- 修正後に `make check` が全通過。Vitest は 15 ファイル / 72 テスト。

## 制約

- 一覧の検索は表示中ページ内の絞り込みに留まる（`TODO(template)`）。契約の
  `listItems` に `q` を追加すればサーバ側検索にできるが、公開契約の変更のため
  ユーザ承認が要る。
- E2E の `sample-resources-multi-page` fixture は毎回 21 件を追加し、削除しない。
  `make e2e-serve` は開発用 DB を使うため、ローカルで繰り返すと行が増える
  （レビューの申し送り N-1）。
- 詳細・編集で採用した 404 の判定は `ApiError.status` に依存する。BFF が別の形で
  エラーを返すようになったら `itemLoadMessage` も直す必要がある。
