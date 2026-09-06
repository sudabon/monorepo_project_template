## 1. ディレクトリ骨格とワークスペース定義

- [x] 1.1 `apps/{api,bff,web}/`、`packages/api-client/`、`api/`、`docs/adr/` を作成し、`find . -type d -not -path './.git/*'` で指示書のディレクトリ構成と一致することを確認する
- [x] 1.2 採用する Go / Node.js / pnpm のバージョンを調査し、完全一致でピンする値を決定して `docs/STACK.md` に採用理由とともに記録する
- [x] 1.3 `go.work` を配置し、`go work sync` が成功することを確認する
- [x] 1.4 `pnpm-workspace.yaml` とルート `package.json`(`engines` / `packageManager`)を配置し、`pnpm install` が成功して lockfile が生成されることを確認する
- [x] 1.5 `.node-version` と `engines.node` が同一の完全一致バージョンを指し、`packageManager` が `pnpm@<version>` で完全一致で固定されていることをスクリプトで照合する

## 2. Makefile

- [x] 2.1 `make setup` を実装し(Go と pnpm の依存取得)、clone 直後の環境で終了コード 0 になることを確認する
- [x] 2.2 `make gen` を Phase 1 の生成処理を差し込める空ターゲットとして定義し、実行が成功することを確認する
- [x] 2.3 `make fmt` / `make lint` を実装し、Go 側と TypeScript 側の両方が呼ばれていることを実行ログで確認する
- [x] 2.4 `make test` / `make build` を実装し、対象がない状態でも終了コード 0 になることを確認する
- [x] 2.5 `make check` を CI と同じ順序(fmt チェック → lint → gen 差分 → test → build)で全ターゲットを直列実行するよう定義し、`make setup && make check` が成功することを確認する

## 3. CI ワークフロー

- [x] 3.1 `.github/workflows/go.yml` を追加し、Go のパスフィルタを設定して各ステップが `make <target>` 呼び出しのみで構成されていることを確認する
- [x] 3.2 `.github/workflows/web.yml` を追加し、TypeScript のパスフィルタを設定して同様に `make <target>` のみで構成されていることを確認する
- [x] 3.3 `.github/workflows/contract.yml` を追加し、`api/` のパスフィルタで起動することを確認する(検証内容の実装は Phase 1)
- [x] 3.4 CI 上で Go / Node のバージョンが `go.mod` / `.node-version` から読み取られるよう設定し、値のハードコードがないことを確認する
- [x] 3.5 required status check とパスフィルタの併用による PR ブロックを回避する集約ジョブの要否を判断し、必要なら追加して判断理由を `docs/adr/` に記録する

## 4. 設定ファイルと依存更新

- [x] 4.1 `.editorconfig`、`.gitattributes` を追加し、`.gitignore` を更新して生成物が除外対象に含まれないことを確認する
- [x] 4.2 `renovate.json` を追加し(patch/minor 週次グループ化＋CI 通過で自動マージ、major は個別 PR で自動マージなし、Go modules / npm / GitHub Actions 対象、lockfile maintenance 有効)、`renovate-config-validator` が成功することを確認する

## 5. E2E ガードの適用範囲対応

- [x] 5.1 `.openspec-e2e-kit.json` に `"templateRepo": true` を追加し、既存フィールドを壊さずパースできることを確認する
- [x] 5.2 `scripts/check-test-plan.sh` の冒頭に、`templateRepo` が真のときスキップして終了 0 を返す処理を追加し、スキップした旨が出力されることを確認する
- [x] 5.3 フラグが偽・フィールド自体がない・設定ファイルがない の 3 ケースでガードが有効(スキップしない)になることを確認する
- [x] 5.4 ガードが有効な状態で、test-plan.md を持たない change(E2E 不要)がタグ付きテスト必須の対象から除外され、test-plan.md がある change では従来どおりタグ付きテストの不在で失敗することを両ケースで確認する
- [x] 5.5 環境変数によるスキップ経路を設けていないことをコードで確認する(design.md D6)
- [x] 5.6 `scripts/check-test-plan.sh` を CI から実行するよう配線し(タスク 3.5 の集約ジョブに含めるか専用ワークフローを追加するかは実装時に判断)、`openspec/changes/` に差分がある PR で実際に起動することを CI 実行履歴で確認する
- [x] 5.7 CI の shallow clone 環境で `origin/main` が解決でき、`git diff origin/main...HEAD` が意図した差分を返すことを確認する(fetch-depth の設定が必要なら合わせて行う)

## 6. 完了条件の検証

- [x] 6.1 クリーンな clone で `make setup && make check` が成功することを確認する
- [x] 6.2 CI の全ジョブが緑になることを確認する
- [x] 6.3 Go のみを変更した検証用 PR で TypeScript 側のジョブが起動しないこと、TypeScript のみを変更した PR で Go 側が起動しないことを CI 実行履歴で確認する

## E2Eテスト実装タスク(必須)

本 change は E2E 不要のため test-plan.md を持たない(ビルドツールチェーンと CI 設定のみで、
ブラウザから観測できる振る舞いを含まないため)。したがって `@phase0-scaffold-toolchain` タグを
持つ E2E テストは実装しない。各シナリオの検証手段は specs のシナリオごとにタスク 1〜4 と 6 で
指定している。対応するガード側の修正と検証はタスク 5.1〜5.7 で行う。

## 検証記録（2026-09-05）

- クリーンなスナップショット clone で `make setup` と `make -j8 check` が成功し、実行後も Git 差分なし。Go / Node.js / pnpm は固定版を利用。
- Node.js 組み込みテストランナーによる 18 件の回帰テストがローカル・CI の両方で成功。フラグの true / false / 未定義 / 設定なし、plan の有無、対象差分なし、JSON 不正、shallow clone からの履歴取得を検証。
- 一時 clone に実 Go ソースを追加し、fmt / vet / test / build の実行、未整形とテスト失敗の伝播、固定 toolchain の gofmt 利用、未追跡生成ファイルの検出を確認。
- `renovate-config-validator 44.65.2 --strict`、`actionlint 1.7.12`、`openspec validate phase0-scaffold-toolchain --strict` が成功。Renovate 検証は Node.js 24.18.0 で実行（任意依存 RE2 は未ビルドのため RegExp へフォールバック）。
- 実装 [PR #1](https://github.com/sudabon/monorepo_project_template/pull/1): [Go](https://github.com/sudabon/monorepo_project_template/actions/runs/33952904482)、[Web](https://github.com/sudabon/monorepo_project_template/actions/runs/33952904455)、[Contract](https://github.com/sudabon/monorepo_project_template/actions/runs/33952904472)、[E2E plan](https://github.com/sudabon/monorepo_project_template/actions/runs/33952904448) がすべて成功。
- E2E plan の checkout ログで `fetch-depth: 0` と `main -> origin/main` を確認。同じ CI 内で、depth 1 の clone から全履歴を取得し、`origin/main...HEAD` が意図した change だけを返す回帰テストが成功。
- Go のみの [PR #2](https://github.com/sudabon/monorepo_project_template/pull/2): 差分は `apps/api/ci_probe.go` のみ。[Go](https://github.com/sudabon/monorepo_project_template/actions/runs/33953137298) と E2E plan が成功し、Web / Contract は起動しない。
- TypeScript のみの [PR #3](https://github.com/sudabon/monorepo_project_template/pull/3): 差分は `apps/web/ci-probe.ts` のみ。[Web](https://github.com/sudabon/monorepo_project_template/actions/runs/33953140088) と E2E plan が成功し、Go / Contract は起動しない。
- 検証用 PR は実装ブランチを base にして、骨格追加の差分を除いて確認した。検証用ソースは実装ブランチに含めない。
