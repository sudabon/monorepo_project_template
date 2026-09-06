# E2E シード fixture 一覧

test-plan.md の「前提(fixture)」列に書いた fixture 名は、必ずこの表に登録すること。
表を見れば「そのテストがどんな状態から始まるか」が読み手に分かる状態を維持する。

## fixture 名 → 作られる状態

| fixture 名 | 作られる状態 | 使用する TP-ID | 方式 |
|-----------|-------------|---------------|------|
| `authenticated-user` | デモユーザ (`demo` / `demo`) でログイン済み。保護ルートを開ける | TP-001〜TP-007 | fixture 直接（ログイン画面操作） |
| `sample-resources-multi-page` | 一意プレフィックス付きのサンプルリソース 21 件。既定 pageSize=20 で 2 ページ目が存在する | TP-006, TP-007 | fixture 直接（認証済み API `POST /api/items`） |

`authenticated-user` は auto fixture のため、テスト本体にログイン操作を書かない。
`sample-resources-multi-page` は `authenticated-user` のあとに API で投入する。

## 方式について

- **シードAPI方式**: テスト用エンドポイント(例 `POST /__test__/seed`)にシード名を渡し、
  アプリ側のトランザクションで状態を作る。本番コードと同じ経路を通るので不整合が起きにくく、
  DB スキーマ変更にも追従しやすい。原則こちらを使う。
- **fixture 直接方式**: Playwright の fixture から DB やストレージへ直接書き込む。
  シードAPIを用意できない外部依存や、API では作れない異常状態(壊れたレコード等)に限って使う。

どちらの方式でも、fixture は各テストの前に**べき等に**状態を作り直し、テスト間で状態を
共有しないこと(順序依存の禁止は `docs/E2E_CONVENTIONS.md` を参照)。

テンプレートにはシードAPIが無いため、本フェーズの fixture はログイン UI と公開 CRUD API
で状態を作る。案件でシードAPIを足したら、この表の方式列を更新する。
