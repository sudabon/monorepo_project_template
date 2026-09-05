# 0001: パス別 CI と required status check

- 状態: 採用
- 日付: 2026-09-05

## 判断

Go、Web、Contract のワークフローは `on.push.paths` / `on.pull_request.paths` で分離する。
共通 Makefile・設定を変更した場合は関係するワークフローを実行する。
Renovate 設定は専用 `renovate.yml` が Node.js 24.x で `make validate-renovate` を実行する。
E2E 計画ガードは専用 `test-plan.yml` で全 PR に対して実行し、差分がない場合はスクリプト内で成功としてスキップする。

確認時点のリポジトリには ruleset・main のブランチ保護・required status check がないため、集約ジョブは追加しない。
パスフィルタでスキップしたワークフローを required check に指定すると Pending のまま PR をブロックし得る。
案件側で required check を導入する際は、フィルタ付きの個別ジョブを直接必須にせず、
必要なワークフローの結果を確認して失敗を伝播する、パスフィルタのない集約ワークフローを追加する。
E2E 計画ガードだけを他の検証の成功とみなしてはならない。

`test-plan.yml` の checkout は `fetch-depth: 0` を指定する。PR の merge commit とベースの共通祖先、
`origin/main` を取得し、`git diff origin/main...HEAD` をローカルと同じ意味で実行する。

## 根拠とトレードオフ

独立したワークフロー間で結果を集約するには待機・API 問い合わせなどの追加実装が必要になる。
現状は required check がないため、その複雑さを持ち込まない。案件側でブランチ保護を有効にする際に判断を更新する。
Renovate は `platformAutomerge: false` によって全チェック成功を自身で確認してから自動マージする。

参考: [GitHub: Troubleshooting required status checks](https://docs.github.com/en/pull-requests/how-tos/merge-and-close-pull-requests/troubleshooting-required-status-checks)、
[Renovate: automerge](https://docs.renovatebot.com/configuration-options/#automerge)。
