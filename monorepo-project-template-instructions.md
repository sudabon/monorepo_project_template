# monorepo_project_template 構築指示書

Coding Agent 向けの作業指示書。受託案件の初期リポジトリとして使うモノレポテンプレートを新規作成する。

---

## 0. 前提と原則

### このリポジトリの目的

受託案件を新規に立ち上げるとき、`Use this template` から clone して即座に開発を始められる状態を提供する。運用は原則ひとりで行い、同時に 5〜8 案件を並行保守することを想定している。

### 参考リポジトリ

`github.com/sudabon/monorepo_go_project_template` を参照する。Go 側の構成・命名・Clean Architecture の層の切り方は、原則そちらを踏襲する。ただし機械的な移植はしない。TypeScript ワークスペースが同居する前提で配置を見直すこと。

### 貫く原則

1. **依存は少ないほど良い。** ライブラリを 1 つ足すたびに、5 年後のメジャーアップグレード工数を買っている。
2. **契約は 1 か所にしか置かない。** Go と TypeScript の境界は OpenAPI 仕様ファイルのみ。型を二重管理しない。
3. **テンプレートは初期値であって規約ではない。** clone 後に案件側で変更されることを前提に書く。「全案件で後から一括更新できること」は目指さない。
4. **ローカルで再現できない CI を作らない。** CI が実行するコマンドは、すべて `make` 経由でローカルからも同じものを実行できること。
5. **生成物はコミットする。** ビルド時生成に頼らず、リポジトリを見ただけで型が読める状態を保つ。

### やらないこと（明示的な非目標）

以下は採用しない。提案もしない。

- Next.js、Remix、TanStack Start、その他 SSR / メタフレームワーク
- Nx、Turborepo などのモノレポビルドツール（pnpm workspace + Makefile で足りる）
- Redux、MobX などの大掛かりな状態管理ライブラリ
- npm 公開前提の独自 UI コンポーネントライブラリ
- 生成コードの `.gitignore` 化
- Docker による本番フロントエンド配信（フロントは静的ファイルとして S3 + CloudFront に置く）

---

## 1. 技術スタック

### 確定事項

| 領域 | 採用 |
|---|---|
| Backend | Go / Echo / Clean Architecture |
| BFF | Go / Echo |
| Frontend | React + TypeScript（Next.js なし） |
| ビルド | Vite |
| ルーティング | TanStack Router |
| サーバ状態 | TanStack Query |
| スタイル | Tailwind CSS |
| UI コンポーネント | コピー型（shadcn/ui 方式。依存パッケージにしない） |
| フォーム | react-hook-form + zod |
| パッケージマネージャ | pnpm（workspace 利用） |
| ユニットテスト | Vitest + Testing Library |
| E2E | Playwright（スモーク 1 本のみ） |
| 依存更新 | Renovate |
| CI | GitHub Actions |

### バージョンの決め方

**具体的なバージョン番号はこの指示書では指定しない。** 作業時点の最新安定版を調べて採用し、以下を必ず行うこと。

- すべて完全一致でピンする（`^` や `~` を使わない）
- lockfile をコミットする
- Node.js は `.node-version` と `package.json` の `engines.node` を完全一致で固定し、両者を一致させる。pnpm は `packageManager` の `pnpm@<version>` で完全一致で固定する
- Go のバージョンは `go.mod` と CI で一致させる
- 採用した主要ライブラリのバージョンと選定理由を `docs/STACK.md` に記録する。次のメジャー移行で何が壊れそうかも 1 行ずつ添える

---

## 2. ディレクトリ構成

```
monorepo_project_template/
├── AGENTS.md                  # ルート。Coding Agent 向けの作業指示
├── README.md                  # 新規案件での使い方
├── Makefile                   # すべてのタスクの入口
├── renovate.json
├── .node-version
├── go.work
├── pnpm-workspace.yaml
├── .github/
│   └── workflows/
│       ├── go.yml             # パスフィルタで Go 変更時のみ
│       ├── web.yml            # パスフィルタで TS 変更時のみ
│       └── contract.yml       # OpenAPI 検証と生成物の差分チェック
├── api/
│   └── openapi.yaml           # Go と TS の唯一の契約
├── apps/
│   ├── api/                   # Go backend
│   │   └── AGENTS.md
│   ├── bff/                   # Go BFF
│   │   └── AGENTS.md
│   └── web/                   # React SPA
│       └── AGENTS.md
├── packages/
│   └── api-client/            # OpenAPI から生成した TS 型 + fetch クライアント
└── docs/
    ├── STACK.md               # 採用バージョンと選定理由
    ├── RUNBOOK.md             # 障害対応と定期メンテ手順
    ├── UPGRADING.md           # 主要ライブラリのメジャー更新手順
    └── adr/                   # 設計判断の記録
```

`AGENTS.md` を階層に分けているのは、フロントエンドの作業時に Go の文脈まで読ませないため。ルートには共通ルールと「どの AGENTS.md を読むべきか」だけを書く。

---

## 3. フェーズ分割

**各フェーズの完了条件を満たしたら、そこで作業を止めて報告すること。** 次のフェーズには進まない。フェーズ 1 本がおおよそ PR 1 本の分量になる想定。

---

### Phase 0 — 骨格とツールチェーンの固定

**やること**

- 上記ディレクトリ構成の骨組みを作る（中身は空でよい）
- `go.work`、`pnpm-workspace.yaml`、`.node-version` を配置
- Makefile に以下のターゲットを定義する。中身は最小で構わないが、すべて実行可能にすること
  - `make setup` — 依存インストール
  - `make gen` — コード生成
  - `make lint` / `make fmt` / `make test` / `make build`
  - `make check` — 上記を CI と同じ順序で全部実行
- GitHub Actions のワークフローを 3 本作る。パスフィルタで Go 側と TS 側を分離すること
- `.editorconfig`、`.gitignore`、`.gitattributes`
- `renovate.json`：
  - patch と minor は週次でグループ化、CI 通過を条件に自動マージ
  - major は個別 PR、自動マージしない
  - Go modules、npm、GitHub Actions のすべてを対象にする
  - lockfile maintenance を有効にする

**完了条件**

- clone 直後に `make setup && make check` が成功する
- CI が全ジョブ緑になる
- Go だけを変更した PR で、TS の単体検証とビルドを行う Web ジョブが走らないことを確認できる（ブラウザ E2E は Go を貫通するため走る。[ADR 0001](docs/adr/0001-ci-path-filters.md)）

---

### Phase 1 — OpenAPI 契約と生成パイプライン

このフェーズがテンプレート全体の中核。ここが機能すれば残りは肉付けになる。

**方針**

`api/openapi.yaml` を手書きの単一ソースとする。Go のコードから仕様を生成する方式は採らない（契約が実装に引きずられるため）。仕様を先に書き、Go と TypeScript の両方をそこから生成する。

**やること**

- `api/openapi.yaml` にサンプルリソース 1 つ分の定義を書く。一覧取得 / 単体取得 / 作成 / 更新 / 削除、エラーレスポンスのスキーマ、ページネーション、バリデーションエラーの形式を含めること
- Go 側：`oapi-codegen` で Echo のサーバインタフェースとリクエスト / レスポンス型を生成
- TS 側：`openapi-typescript` で型定義、`openapi-fetch` で型付き fetch クライアントを生成し、`packages/api-client` に配置
- `packages/api-client` に、TanStack Query の `queryOptions` を返す薄いラッパを手書きで用意する。生成物の上に 1 枚被せる形にして、生成物そのものは編集しない
- `api/openapi.yaml` の lint を CI に入れる（Spectral か Redocly）
- **生成物はすべてコミットする。** CI に `make gen && git diff --exit-code` を入れ、仕様と生成物がずれた PR を落とす

**完了条件**

- `api/openapi.yaml` にフィールドを 1 つ足して `make gen` すると、Go と TS の両方の生成物に反映される
- 生成物を編集せずに `make gen` を実行しても差分が出ない（冪等）
- 仕様だけ変更して生成し忘れた PR が CI で落ちる

---

### Phase 2 — Go backend の参照実装

**やること**

- 参考リポジトリの Clean Architecture 構成を踏襲し、Phase 1 のサンプルリソースを一通り実装する
- 層は最低限、handler / usecase / domain / repository に分ける。依存の向きが内側を向いていることをテストか lint で保証する
- DB アクセスの実装を 1 つ入れる。マイグレーションツールも決めて組み込む
- 構造化ログ（`log/slog`）とリクエスト ID の伝播
- ヘルスチェックエンドポイント（ALB のターゲットグループ用に、依存先の生死を見る deep check と、プロセスの生死だけを見る shallow check を分ける）
- graceful shutdown（ECS のタスク停止時にリクエストを取りこぼさないこと）

**完了条件**

- `make test` でユニットテストが通る
- ローカルで起動して、サンプルリソースの CRUD が curl で一通り叩ける
- SIGTERM 送出後、処理中のリクエストが完了してから終了する

---

### Phase 3 — Go BFF

**やること**

- 認証はここで完結させる。セッションを httpOnly / Secure / SameSite=Lax の Cookie で持ち、**SPA 側にはトークンを一切渡さない**
- ログイン / ログアウト / セッション確認のエンドポイント
- backend への転送時に、認証済みユーザ情報とリクエスト ID を付与する
- CSRF 対策（Cookie セッションを使うため必須）
- 未認証時は 401 を返す。リダイレクトは返さない（SPA 側で処理するため）

**完了条件**

- ログイン → 保護されたリソースの取得 → ログアウト → 401 が curl で再現できる
- CSRF トークンなしの POST が拒否される

---

### Phase 4 — React SPA の骨格

ここが「ディレクトリ構成以外の 9 割」にあたる部分。順に全部入れること。

**4-1. ビルドと環境設定**

- Vite + React + TypeScript。`tsconfig` は `strict: true` に加えて `noUncheckedIndexedAccess` を有効にする
- **環境ごとの設定はビルド時に埋め込まない。** ビルド成果物は全環境で 1 つだけ作り、起動時に `/config.json` を fetch して API のベース URL などを読む方式にする。これをやらないと、開発 / ステージング / 本番でビルドをやり直すことになる
- `config.json` の型定義と、読み込み失敗時のフォールバック挙動を書く
- SPA なので、秘密情報はどこにも置けない。`.env.example` にその旨をコメントで明記する

**4-2. ルーティングと状態**

- TanStack Router をファイルベースルーティングで設定
- 認証ガードを `beforeLoad` に実装。未認証なら遷移前にログイン画面へ
- TanStack Query の `QueryClient` を 1 か所で設定。`staleTime`、`retry`、`refetchOnWindowFocus` の既定値を明示的に決めてコメントで理由を書く
- グローバルな `onError`：401 ならセッション切れとして扱いログインへ、5xx ならトースト表示、それ以外は握りつぶさず表示

**4-3. UI 基盤**

- Tailwind を設定
- 汎用 UI コンポーネント（ボタン、入力、モーダル、テーブル、トースト）を `apps/web/src/components/ui/` に**自分のコードとして**置く。UI ライブラリを依存に追加しない
- ダークモードは実装しない（案件側で必要になったら足す）

**4-4. エラーとローディング**

- ルート単位の ErrorBoundary
- 予期しない例外を拾う最上位の ErrorBoundary。ここに監視サービスへの送信フックを空実装で用意し、差し込み方を `docs/RUNBOOK.md` に書く
- ローディングは Suspense を使うか使わないかを決めて統一する

**4-5. フォーム**

- react-hook-form + zod。バリデーションスキーマは、可能な範囲で `api/openapi.yaml` の制約と対応させる
- サーバ側のバリデーションエラー（Phase 1 で定義した形式）をフォームのフィールドエラーにマッピングする共通処理を書く

**完了条件**

- `pnpm dev` で起動し、ログイン画面が表示される
- `pnpm build` の成果物を静的サーバで配信し、`config.json` を差し替えるだけで接続先 API が変わる
- API を停止した状態でもアプリが白画面にならず、エラー表示になる

---

### Phase 5 — フルスタックの貫通

参照実装として、1 つのリソースを DB からブラウザまで通す。

**やること**

- 一覧画面（ページネーション、検索条件を URL に持たせる。TanStack Router の型付き search params を使う）
- 詳細画面
- 作成 / 編集フォーム（サーバ側バリデーションエラーの表示まで）
- 削除（確認ダイアログ付き）
- Playwright でこの一連の流れをスモークテスト 1 本にする

**完了条件**

- `make check` が全部通る
- Playwright のスモークが CI で緑になる
- 一覧画面で検索条件を変えた後、ブラウザをリロードしても条件が保持される

---

### Phase 6 — 運用ドキュメントとテンプレート化

**やること**

- `README.md`：新規案件での使い方。clone 後に書き換えるべき箇所のチェックリストを含める（モジュールパス、パッケージ名、リソース名、CI のシークレット名など）
- `docs/STACK.md`：採用バージョン一覧と選定理由
- `docs/UPGRADING.md`：React、Tailwind、TanStack Router、Go、Echo それぞれのメジャー更新手順と、過去に踏んだ落とし穴を書き足していく場所
- `docs/RUNBOOK.md`：
  - Aurora のマイナー / メジャーアップグレード手順
  - Terraform provider のメジャー更新手順
  - 依存更新 PR のレビュー基準（本番バンドルに入るか入らないかで扱いを変えること）
  - `npm audit` の警告をどう切り分けるか。devDependencies の脆弱性は本番配信物に届かないため、対応優先度を分けること
  - 障害時の初動（ログの見方、リクエスト ID での追跡方法）
- `docs/adr/0001-*.md`：ここまでの主要な判断を ADR として記録
- 各 `AGENTS.md` を書く
- CloudFront 配信時のキャッシュ設定を `docs/` に記載：`index.html` は no-cache、ハッシュ付きアセットは immutable、403/404 は `/index.html` に 200 で fallback

**完了条件**

- README を読んだだけで、新規案件を 30 分以内に立ち上げられる
- リポジトリを GitHub の Template repository に設定できる状態

---

## 4. 作業時の注意

- 各フェーズの完了時に、**変更点の要約と、判断に迷った箇所を明示して報告すること**。勝手に次へ進まない
- 依存を新規追加するときは、追加理由と代替案を報告に含める。自明でないものは ADR に残す
- 「とりあえず動く」実装で妥協した箇所には `// TODO(template):` を付け、報告に一覧を含める
- 指示書と現実が食い違ったら、勝手に解釈を変えず、食い違いを報告して判断を仰ぐ
