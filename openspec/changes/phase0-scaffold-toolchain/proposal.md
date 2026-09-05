## Why

受託案件を立ち上げるたびに、ディレクトリ構成・タスクランナー・CI・依存更新設定を作り直している。1 人で 5〜8 案件を並行保守する前提では、この初期設定のばらつきがそのまま保守コストになる。以降のフェーズ（契約生成・Go 実装・SPA）はすべてこの土台の上に乗るため、最初に「すべての作業の入口」と「CI の走らせ方」を固定する。

## What Changes

- モノレポのディレクトリ骨格を作成する（`apps/api`、`apps/bff`、`apps/web`、`packages/api-client`、`api/`、`docs/`。中身は空でよい）
- ワークスペース定義を配置する: `go.work`、`pnpm-workspace.yaml`、`.node-version`
- `Makefile` をすべてのタスクの単一の入口として定義する: `setup` / `gen` / `lint` / `fmt` / `test` / `build` / `check`
- GitHub Actions を 3 本追加する（`go.yml` / `web.yml` / `contract.yml`）。パスフィルタで Go 側と TypeScript 側のジョブを分離する
- `renovate.json` を追加する。patch/minor は週次グループ化＋CI 通過を条件に自動マージ、major は個別 PR で自動マージしない。Go modules / npm / GitHub Actions を対象にし、lockfile maintenance を有効にする
- `.editorconfig`、`.gitignore`、`.gitattributes` を整備する
- ツールチェーンのバージョンを完全一致でピンし、lockfile をコミットする
- `scripts/check-test-plan.sh`（E2E ガード）に、テンプレートリポジトリ自身では無効化する経路を追加する。`.openspec-e2e-kit.json` の `templateRepo` フラグで判定し、フラグがない案件リポジトリでは既定で有効とする

## Capabilities

### New Capabilities

- `build-toolchain`: リポジトリの開発タスク実行契約。`make` を唯一の入口とするタスク定義、CI とローカルの手順一致、言語別のパス分離実行、ツールチェーンバージョンの固定、依存更新の取り込み方針を含む。

### Modified Capabilities

なし（初回変更のため既存 spec は存在しない）。

## Impact

- **新規ファイル**: `Makefile`、`go.work`、`pnpm-workspace.yaml`、`.node-version`、`renovate.json`、`.editorconfig`、`.gitattributes`、`.github/workflows/{go,web,contract}.yml`
- **変更ファイル**: `.gitignore`、`scripts/check-test-plan.sh`、`.openspec-e2e-kit.json`
- **ディレクトリ**: `apps/{api,bff,web}/`、`packages/api-client/`、`api/`、`docs/`
- **依存**: pnpm（`packageManager` で固定）、Node.js、Go toolchain。アプリケーション依存はこのフェーズでは追加しない
- **後続への影響**: Phase 1 以降のすべてのターゲット（`make gen` の中身など）はこのフェーズで定義した入口に追記する形になる
