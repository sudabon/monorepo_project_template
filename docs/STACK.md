# ツールチェーン

2026-09-05 に公式配布元で確認した最新安定版（プレリリースを除く）を採用する。

| ツール | 固定バージョン | 採用理由 | 次のメジャー更新で確認する点 |
| --- | --- | --- | --- |
| Go | 1.27.1 | 最新安定版。`go.work` と各 `go.mod` を統一し、Make は同じ toolchain を選択する | Go 1.x の更新でもコンパイラ・vet の診断と標準ライブラリの変更を確認 |
| Node.js | 26.8.1 | 最新安定版の Current 系を採用。LTS 系ではない。`.node-version` と `engines.node` を統一 | ESM、組み込みテストランナー、開発ツールの対応範囲 |
| pnpm | 11.25.0 | 最新安定版。workspace と単一 lockfile を利用し、`packageManager` で固定 | lockfile 形式、設定キー、依存のビルドスクリプト実行方針 |
| Biome | 2.5.12 | TypeScript の formatter と linter を開発依存 1 つで提供 | 設定スキーマ、推奨ルール、整形結果の変更 |

配布元: [Go](https://go.dev/dl/)、[Node.js](https://nodejs.org/dist/index.json)、[pnpm](https://registry.npmjs.org/pnpm/latest)、[Biome](https://registry.npmjs.org/@biomejs/biome/latest)。

Biome は ESLint + Prettier より依存と設定を少なくできるため採用した。アプリケーション依存は Phase 0 では追加しない。
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
`make gen` は Phase 1 で生成処理を追加するための空ターゲット。
`make gen-check` の差分検出範囲は `GENERATED_PATHS` に列挙した生成物に限定する。Phase 0 では未設定のため成功としてスキップし、未コミットの手書きコードでは失敗しない。

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
