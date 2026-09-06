# ADR

覆すコストが高い判断を、判断・理由・代替案つきで残す。change の `design.md` は archive されるため、長期に参照するものだけをここに置く。

## 運用

- ファイル名は `NNNN-kebab-case.md`。採番は追加順。
- 1 判断 1 ファイル。状態は `採用` / `廃止`。
- **覆された判断は削除しない。** 後続の ADR で上書きし、古いファイルの状態を `廃止` にして新しい ADR へリンクする。
- 既に ADR がある判断を繰り返さない。参照する。

## 索引

| 番号 | 判断 | 由来 |
| --- | --- | --- |
| [0001](0001-ci-path-filters.md) | パス別 CI と required status check | Phase 0 |
| [0002](0002-openapi-contract-tools.md) | 契約の lint と生成ツール | Phase 1 |
| [0003](0003-api-database-and-architecture.md) | 層構造、依存方向検査、DB、マイグレーション | Phase 2 |
| [0004](0004-bff-session-store.md) | セッションストア | Phase 3 |
| [0005](0005-bff-identity-forwarding.md) | BFF から API へのユーザ識別 | Phase 3 |
| [0006](0006-shared-go-platform-module.md) | Go 実行時基盤の共有モジュール | Phase 3 |
| [0007](0007-handwritten-openapi-contract.md) | 契約は手書きし、実装からは生成しない | Phase 1 D1 |
| [0008](0008-auth-stays-in-bff.md) | ブラウザ向け認証は BFF に閉じる | Phase 3 |
| [0009](0009-runtime-config-fetch.md) | 環境設定は実行時に fetch する | Phase 4 D1 |
| [0010](0010-copy-owned-ui-components.md) | UI 部品はコピー型で持つ | Phase 4 D8 |

層構造と依存方向は 0003 が既に記録しているため、重複する ADR は作らない。
