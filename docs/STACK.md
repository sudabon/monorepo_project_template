# ツールチェーン

基本ツールチェーンは 2026-09-05、契約ツールは 2026-09-06 に公式配布元で確認した
安定版（プレリリースを除く）を採用する。依存の対応範囲による制約は以下に記す。

| ツール | 固定バージョン | 採用理由 | 次のメジャー更新で確認する点 |
| --- | --- | --- | --- |
| Go | 1.27.1 | 最新安定版。`go.work` と各 `go.mod` を統一し、Make は同じ toolchain を選択する | Go 1.x の更新でもコンパイラ・vet の診断と標準ライブラリの変更を確認 |
| Node.js | 26.8.1 | 最新安定版の Current 系を採用。LTS 系ではない。`.node-version` と `engines.node` を統一 | ESM、組み込みテストランナー、開発ツールの対応範囲 |
| pnpm | 11.25.0 | 最新安定版。workspace と単一 lockfile を利用し、`packageManager` で固定 | lockfile 形式、設定キー、依存のビルドスクリプト実行方針 |
| Biome | 2.5.12 | TypeScript の formatter と linter を開発依存 1 つで提供 | 設定スキーマ、推奨ルール、整形結果の変更 |
| oapi-codegen | 2.8.0 | `apps/api/go.mod` の tool directive で固定し、Echo の通常・strict インタフェースと型を生成 | 設定スキーマ、生成インタフェース、Echo / runtime の対応範囲 |
| Echo | 4.15.4 | API の生成サーバと BFF の実行時依存 | v5 は生成設定と handler の同時移行が必要 |
| oapi-codegen/runtime | 1.7.0 | 生成されたパラメータ・strict middleware の共通処理 | バインド処理と strict handler のシグネチャ |
| openapi-typescript | 7.13.0 | 読める型定義だけを生成し、実行時コードを増やさない | OpenAPI の解釈、生成型名、TypeScript peer の対応範囲 |
| openapi-fetch | 0.17.0 | 生成型に従う薄い fetch クライアント | 0.x の minor 更新でも返り値・middleware・パスの型推論を確認 |
| TypeScript | 5.9.3 | openapi-typescript 7.13.0 の peer が 5.x のため、その最新 patch を固定 | 6 / 7 への更新は生成器の peer 対応を待ち、型テストで確認 |
| TanStack React Query | 5.102.8 | 手書きラッパから型付き options を公開。api-client の peer と開発依存に固定 | options helper、query key の型推論、React peer の対応範囲 |
| React / @types/react | 19.2.8 / 19.2.18 | Query の開発時 peer と型検証に使用 | React peer と JSX / hook 型の互換性 |
| @types/node | 26.4.1 | `node:test` を使う通信テストを `tsc --noEmit` の対象に含めるため。Node.js 26 系に合わせる | Node.js のメジャー更新への追随、DOM lib との global 衝突 |
| Spectral CLI | 6.16.3 | OAS 推奨ルールと契約固有ルールを pnpm から実行 | 推奨ルールの追加、Node 対応、診断と終了コード |
| PostgreSQL | 18.3-alpine | ローカルの Compose と CI のサービスコンテナで同じ版を使う | SQL とデータ型の非互換、`pg_isready` の挙動、major 間のダンプ移行手順 |
| pgx | 5.10.0 | `database/sql` ドライバとして SQL を直接書く。ORM を持ち込まない | DSN の解釈、`stdlib` の接続プール設定、型変換の変更 |
| goose | 3.28.0 | 単一の SQL に Up / Down を書き、Provider API を統合テストでも使う | Provider API のシグネチャ、version テーブルの構成、埋め込み FS の扱い |
| go-arch-lint | 1.18.0 | 層の依存許可グラフを宣言で書き、`make lint-go` で検査 | 設定 version、component 記法、`deepScan` の判定 |
| google/uuid | 1.6.0 | リクエスト ID の生成と生成型の UUID 変換 | 生成方式と文字列表現 |

配布元: [Go](https://go.dev/dl/)、[Node.js](https://nodejs.org/dist/index.json)、[pnpm](https://registry.npmjs.org/pnpm/latest)、[Biome](https://registry.npmjs.org/@biomejs/biome/latest)。

Biome は ESLint + Prettier より依存と設定を少なくできるため採用した。
契約ツールの固定値は [Go module proxy](https://proxy.golang.org/)、
[npm registry](https://registry.npmjs.org/) で確認した。選定理由と lint の無効化ルールは
[ADR 0002](adr/0002-openapi-contract-tools.md) を参照。
Spectral の推移依存 `@scarf/scarf` は analytics 用のインストール処理が不要なため、
`pnpm-workspace.yaml` の `allowBuilds` で明示的に無効化する。
Playwright の既存設定・レポートスクリプトは後続フェーズで利用し、ブラウザテストや Playwright のインストールはここでは行わない。
E2E の実装規約は [E2E_CONVENTIONS.md](E2E_CONVENTIONS.md) に置く。

## ローカルと CI

Git、Bash、Make、指定の Node.js / pnpm、および Go の toolchain 自動取得に対応した Go を用意して、ルートで `make setup`、`make check` を実行する。
Node.js の既存バージョンマネージャで `.node-version` の値をインストールし、pnpm は `packageManager` の値に合わせる。
Go は Make が `go.work` から `GOTOOLCHAIN` を設定するため、インストール済み Go が異なる場合は固定版を自動取得する。
CI の Node.js は `.node-version`、Go は `apps/api/go.mod`、pnpm は `package.json` から取得する。
`make lint-web` は Node.js と pnpm の固定値・実行バージョン・依存宣言を検証し、`make lint-go` は全 Go module と workspace の固定値を検証する。

`make check` は fmt チェック → lint → gen 差分 → test → build → E2E 計画ガードの順に直列実行する。
Go と TypeScript のソースがまだない workspace は成功としてスキップし、追加後は同じ入口で処理する。
`make gen` は `api/openapi.yaml` から Go と TypeScript の両方を生成する。
`make lint-contract` は Spectral の warning 以上を失敗とする。
`make gen-check` の差分検出範囲は `GENERATED_PATHS` に列挙した生成物に限定し、
HEAD との差分と未追跡の生成物を検出する。生成物は仕様と一緒にコミットする。
手書きコードだけの未コミット変更では失敗しない。
Go / Web CI は `make gen-check-go` / `make gen-check-web` を使い、他言語の実行環境を
必要としない。Go CI が実行するのは `make test-go` であり、`make test` は使わない。
Contract CI は両方の実行環境を用意し、lint と `make gen-check`、
`git diff --exit-code` を実行する。

`make lint-go` は go vet に加えて go-arch-lint で層の依存方向を検査し、
`go mod tidy -diff` で go.mod の未整理を検出する。Go の統合テストは実 PostgreSQL を使う。
`make test-go` は `TEST_DATABASE_URL` が未設定なら Compose を起動し、CI は同じ環境変数で
サービスコンテナを指す。DB・マイグレーション・依存方向検査の選定理由は
[ADR 0003](adr/0003-api-database-and-architecture.md)、起動手順は [RUNBOOK](RUNBOOK.md) を参照。

TypeScript の生成物 `packages/api-client/src/generated/` は Biome の対象外。
生成型を手編集せず、契約を修正して `make gen` を再実行する。
Go の生成先は `apps/api/internal/generated/`。生成は `scripts/go-task.sh gen` 経由で行い、
GOTOOLCHAIN の導出をこのスクリプト 1 か所に閉じる。サーバ実装は `apps/api/` と
`apps/bff/` にあり、層の責務は [apps/api/AGENTS.md](../apps/api/AGENTS.md) と
[apps/bff/AGENTS.md](../apps/bff/AGENTS.md) に置く。BFF のセッションは API と同じ
PostgreSQL を goose で管理する。選定理由は [ADR 0004](adr/0004-bff-session-store.md)。
`packages/api-client` の型検証は `src` と `tests` の両方を対象にする。

アプリケーションは `@monorepo-project-template/api-client` の package entry から
`createItemQueries({ baseUrl: '/api' })` を作り、`items.list({ page: 1, pageSize: 20 })`
や `items.get(id)` を Query に渡す。baseUrl は Phase 4 の実行時設定から渡す。
更新系は `createItemMutations({ baseUrl })` の `create()` / `update()` / `delete()` を
`useMutation` に渡す。`mutate` の引数はそれぞれ body、`{ id, body }`、id。
作成・更新は Item、削除はレスポンス本文なし（void）を返す。
生成物の直接 import やアプリケーション側での型の二重定義は不要。
HTTP エラーは `ApiError` に status と契約の body を保持して送出する。
通信エラーはそのまま伝播し、再試行・表示・ログイン遷移・キャッシュ無効化の方針は
アプリケーションの QueryClient が決める。GET は Query の AbortSignal を fetch に渡す。

`make check-test-plan` は PR の E2E 計画ガード。既定の比較先は `origin/main`、別ブランチは `make check-test-plan BASE=origin/<branch>` で指定する。
テンプレートでは `.openspec-e2e-kit.json` の `templateRepo: true` によりスキップし、案件作成時にこのフラグを削除する。
フラグがない場合、test-plan.md を持つ change にタグ付き E2E テストを要求する。

## 依存更新

Renovate は Go modules / npm / GitHub Actions の patch・minor を毎週月曜の 0〜6 時（Asia/Tokyo）にまとめ、全 CI 成功後にマージする。
`platformAutomerge: false` とし、ブランチ保護が未設定でも Renovate 自身がチェック結果を確認する。
major は個別 PR とし自動マージしない。lockfile maintenance も週次で行う。
バージョン固定の整合性チェックが失敗した更新 PR は、`.node-version` / `engines.node`、または `go.work` / 各 `go.mod` を同時に更新する。

Renovate 設定の破損は CI の `renovate.yml` が Node.js 24.x で `make validate-renovate` を実行して検出する。
Renovate 44.65.2 は Node.js 24.x 向けのため、リポジトリ固定の Node.js 26 では実行しない。
通常の `make setup` に Renovate 本体は追加しない。CI の required check の判断は [ADR 0001](adr/0001-ci-path-filters.md) を参照。
