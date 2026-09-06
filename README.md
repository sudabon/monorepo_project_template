# monorepo_project_template

BtoB SaaS 向けのモノレポテンプレート。GitHub の Template repository から複製して使う。

設計の背景は書かない。[docs/](docs/STACK.md) と [ADR](docs/adr/README.md) を参照する。

## 最初に実行するコマンド

```sh
make setup
make check
```

ローカルで API / BFF / SPA を動かす手順は [docs/RUNBOOK.md](docs/RUNBOOK.md)。

## 複製後のチェックリスト

次の識別子はテンプレート固有で、全文検索で一意に見つかる。すべて案件の値に置き換える。

| # | 対象 | テンプレートの値 | 置き換える場所の目安 |
| --- | --- | --- | --- |
| 1 | Go モジュールパス | `github.com/sudabon/monorepo_project_template` | 各 `go.mod`、`replace`、Go の import、`.go-arch-lint.yml` |
| 2 | npm パッケージ名 | `@monorepo-project-template/*`、ルート `monorepo-project-template` | `package.json`、`Makefile` の `--filter`、ソースの import、`pnpm-lock.yaml`（置き換え後に `pnpm install`） |
| 3 | サンプルリソース名 | OpenAPI の `items` / `Item` / `listItems`、テーブル `items`、`createItemQueries` | `api/openapi.yaml`、`apps/api`、`packages/api-client`、`apps/web`、`tests/e2e`。案件の最初のリソースに置き、`make gen` する |
| 4 | CI のシークレット名 | 現状 workflow に `secrets.*` 参照はない。サービスコンテナと Make の DB 認証はユーザ / パスワード / DB 名とも `template` | シークレットを足したら GitHub Secrets 名とこの表を案件名に合わせる。`.github/workflows/go.yml` の `POSTGRES_*`、`Makefile` の `DATABASE_URL`、`compose.yaml`、`scripts/go-task.sh` の `template` も変える |
| 5 | テンプレート専用ガードの解除 | `.openspec-e2e-kit.json` の `"templateRepo": true` | **キーごと削除する**（`false` でもガードは有効になるが、テンプレート由来の設定は残さない） |

合わせて次も直す。

- `compose.yaml` の `name: monorepo-project-template`
- OpenAPI の `info.title`（`Monorepo template API`）
- 複製先では `monorepo-project-template-instructions.md` を削除する（テンプレート作者向け）
- 識別子を置き換えたら次を実行する。モジュールパスの長さで import の折返しが変わり、`require` の並びも新しいパスの辞書順に直る。

```sh
go -C apps/api mod tidy
go -C apps/bff mod tidy
go -C packages/go-platform mod tidy
make fmt
```

## 適用後の検証

ドキュメント上の言及を除いて、テンプレート識別子が残っていないことを確認する。

```sh
rg -n --hidden \
  -g '!.git/**' \
  -g '!README.md' \
  -g '!docs/**' \
  -g '!openspec/**' \
  'github\.com/sudabon/monorepo_project_template|monorepo-project-template|"templateRepo"\s*:\s*true'
```

ヒットが 0 件であること。`templateRepo` を消さずにこの検索を実行すると `.openspec-e2e-kit.json` が検出される。

続けて `make setup && make check` が成功すること。

## ドキュメント

| 用途 | 場所 |
| --- | --- |
| 採用バージョンと選定理由 | [docs/STACK.md](docs/STACK.md) |
| メジャー更新手順 | [docs/UPGRADING.md](docs/UPGRADING.md) |
| 起動・障害・依存更新・DB | [docs/RUNBOOK.md](docs/RUNBOOK.md) |
| SPA の実行時設定と静的配信 | [docs/web.md](docs/web.md) |
| BFF の Cookie・同一オリジン前提 | [docs/bff.md](docs/bff.md) |
| 設計判断 | [docs/adr/](docs/adr/README.md) |
| Coding Agent 向け指示 | [AGENTS.md](AGENTS.md) |

GitHub の Settings → General → Template repository を有効にできる。`.env` や秘密鍵は `.gitignore` 済みで、workflow も `secrets.*` を参照していない。
