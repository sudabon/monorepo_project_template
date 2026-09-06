# 共通ルール

このファイルは案内だけを書く。具体的な実装規約は作業対象の階層側が正。

## どの指示を読むか

| 作業対象 | 読むファイル |
| --- | --- |
| SPA・フロントエンド | [apps/web/AGENTS.md](apps/web/AGENTS.md) |
| BFF・ブラウザ向け認証 | [apps/bff/AGENTS.md](apps/bff/AGENTS.md) |
| API | [apps/api/AGENTS.md](apps/api/AGENTS.md) |
| API クライアント（TypeScript） | [packages/api-client/AGENTS.md](packages/api-client/AGENTS.md) |
| Go 共有基盤（ログ・shutdown・DB プール） | [packages/go-platform/AGENTS.md](packages/go-platform/AGENTS.md) |
| 契約（`api/openapi.yaml`）・E2E | 変更が及ぶ側の上記を読む。契約は両言語の生成元 |

作業対象のファイルだけを読む。他階層の `AGENTS.md` は開かない。

## 共通

- 検証の入口は `make check`。部分実行するときも Makefile のターゲットを使い、言語ツールを直接呼び出して順序を変えない。
- 生成物は手編集しない。契約やルート定義を直してから生成し直す。生成先は各階層の `AGENTS.md` に書く。
- 秘密情報をコミットしない。コミットメッセージは変更の理由を先に書く。
