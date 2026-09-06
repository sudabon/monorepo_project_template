## Why

テンプレートは「動くこと」より「引き継げること」で価値が決まる。1 人で 5〜8 案件を並行保守する前提では、半年後の自分が読んで思い出せるかがそのまま保守コストになる。Phase 0〜5 で下した判断は、記録しなければ次の案件で同じ検討をやり直すことになる。

また、テンプレートとして公開する以上、clone 後に何を書き換えるべきかが明示されていないと、案件ごとにテンプレート由来の識別子が残る。

## What Changes

- `README.md` を新規案件での使い方として書き直す。clone 後に書き換えるべき箇所のチェックリスト（モジュールパス、パッケージ名、リソース名、CI のシークレット名など）を含める
- チェックリストに、テンプレート専用に無効化している E2E ガード（`.openspec-e2e-kit.json` の `templateRepo` フラグ）の解除を含める。解除漏れは全文検索の検証手順で検出できるようにする
- `docs/STACK.md` に採用バージョン一覧と選定理由、次のメジャー移行で壊れそうな点を記録する
- `docs/UPGRADING.md` に React / Tailwind / TanStack Router / Go / Echo それぞれのメジャー更新手順を書き、踏んだ落とし穴を書き足していく場所として位置づける
- `docs/RUNBOOK.md` を書く: Aurora のマイナー / メジャーアップグレード手順、Terraform provider のメジャー更新手順、依存更新 PR のレビュー基準（本番バンドルに入るか否かで扱いを変える）、`npm audit` 警告の切り分け（devDependencies の脆弱性は本番配信物に届かないため優先度を分ける）、障害時の初動（ログの見方、リクエスト ID での追跡）
- `docs/adr/` にここまでの主要な判断を ADR として記録する
- 各階層の `AGENTS.md` を書く。ルートには共通ルールと「どの AGENTS.md を読むべきか」だけを書く
- CloudFront 配信時のキャッシュ設定を `docs/` に記載する: `index.html` は no-cache、ハッシュ付きアセットは immutable、403/404 は `/index.html` に 200 で fallback
- リポジトリを GitHub の Template repository に設定できる状態にする

## Capabilities

### New Capabilities

- `template-bootstrap`: テンプレートから新規案件を立ち上げるための振る舞い。書き換え箇所のチェックリストと、その適用によってテンプレート由来の識別子が残らないこと、運用・更新・障害対応の手順が参照可能であること、Coding Agent 向け指示が階層ごとに分離されていること、静的配信の要件が明示されていることを含む。

### Modified Capabilities

なし。既存の各 capability の要求は変更しない。

## Impact

- **新規ファイル**: `docs/STACK.md`（Phase 0〜5 で追記済みなら再構成）、`docs/UPGRADING.md`、`docs/RUNBOOK.md`、`docs/adr/0001-*.md` ほか、`AGENTS.md`（ルート / `apps/api` / `apps/bff` / `apps/web`。各フェーズで作成済みなら整合を取る）
- **変更ファイル**: `README.md`
- **依存追加**: なし
- **前提**: Phase 0〜5 が完了していること。ADR には各フェーズで下した判断を集約する
