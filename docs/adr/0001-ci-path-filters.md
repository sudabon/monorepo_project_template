# 0001: パス別 CI と required status check

- 状態: 採用
- 日付: 2026-09-05
- 更新: 2026-09-07（Phase 5 で追加した `e2e.yml` と、パス分離が保証する範囲を追記）

## 判断

ワークフローは 6 本。Go、Web、Contract、E2E は `on.push.paths` / `on.pull_request.paths` で分離する。
共通 Makefile・設定を変更した場合は関係するワークフローを実行する。
Renovate 設定は専用 `renovate.yml` が Node.js 24.x で `make validate-renovate` を実行する。
E2E 計画ガードは専用 `test-plan.yml` で全 PR に対して実行し、差分がない場合はスクリプト内で成功としてスキップする。

パスフィルタはワークフロー単位にしか書けず、ジョブ単位には書けない。
そのため E2E スモークは `web.yml` の 1 ジョブにできず、独立した `e2e.yml` になる。
E2E は SPA から BFF を通って API まで貫通するため、`apps/api/**` と `apps/bff/**` も対象パスに含める。

### パス分離が保証する範囲

Go のソースだけを変更した Pull Request で走らないのは、TypeScript の単体検証とビルドを行う
`web.yml` と、契約検証の `contract.yml` である。`e2e.yml` は Go の変更を貫通して検証するため走り、
`test-plan.yml` はパスフィルタを持たないため常に走る。
**「Go の変更で Node.js が一切動かない」ことは保証しない。** 保証するのは、Go の変更が
TypeScript の単体検証と契約生成物の検査を巻き込まないことである。
E2E を Go の変更から外せば PR あたり数分は縮むが、BFF のヘッダ転送や API の応答形式を壊した PR が
緑になる。実行時間より、貫通経路の破壊を PR で捕まえることを優先する。

確認時点のリポジトリには ruleset・main のブランチ保護・required status check がないため、集約ジョブは追加しない。
パスフィルタでスキップしたワークフローを required check に指定すると Pending のまま PR をブロックし得る。
案件側で required check を導入する際は、フィルタ付きの個別ジョブを直接必須にせず、
必要なワークフローの結果を確認して失敗を伝播する、パスフィルタのない集約ワークフローを追加する。
`test-plan.yml` の E2E 計画ガードは計画の有無しか見ない。これを `e2e.yml` のスモークや
他の検証の成功とみなしてはならない。

`test-plan.yml` の checkout は `fetch-depth: 0` を指定する。PR の merge commit とベースの共通祖先、
`origin/main` を取得し、`git diff origin/main...HEAD` をローカルと同じ意味で実行する。

全ワークフローに `concurrency` を設定し、Pull Request では先行実行を打ち切る。
main への push は打ち切らない。main の各コミットに対する検証結果を残すため。
Web ジョブは `pnpm/action-setup` を `setup-node` より前に置き、`cache: pnpm` で pnpm ストアを再利用する。

## 根拠とトレードオフ

独立したワークフロー間で結果を集約するには待機・API 問い合わせなどの追加実装が必要になる。
現状は required check がないため、その複雑さを持ち込まない。案件側でブランチ保護を有効にする際に判断を更新する。
Renovate は `platformAutomerge: false` によって全チェック成功を自身で確認してから自動マージする。

参考: [GitHub: Troubleshooting required status checks](https://docs.github.com/en/pull-requests/how-tos/merge-and-close-pull-requests/troubleshooting-required-status-checks)、
[Renovate: automerge](https://docs.renovatebot.com/configuration-options/#automerge)。
