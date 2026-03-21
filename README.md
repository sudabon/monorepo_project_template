# monorepo_project_template

クリーンアーキテクチャに沿ったモノレポ Web アプリを実装するための**ベース／テンプレート**用リポジトリです。依存の向きとソース配置をテストで固定し、エージェント向けの規約をリポジトリ内にまとめています。

## 現状のリポジトリに含まれるもの

| 内容 | 説明 |
|------|------|
| **アーキテクチャテスト** | `backend/tests/arch/` に unittest ベースの検証があります。Python パッケージ名は **`todo_app`** を想定しています（フォーク時にリネームしてください）。 |
| **依存ルール** | `test_dependency_rule.py` … `domain` 層が `application` / `interfaces` / `infrastructure` へ内向き import しないことを AST で検査します。 |
| **ディレクトリ構造** | `test_source_structure.py` … `todo_app` 直下に **`domain` → `app` → `interface` → `infra`** の4層フォルダのみがあることを検証します（内側から外側の順）。 |
| **エージェント規約** | `.cursor/rules/`、`.claude/rules/`、`.codex/rules/` にアーキテクチャ・バックエンド／フロント規約・テスト・Git などを配置しています。 |

依存の方向の原則（`.cursor/rules/architecture.md` と同趣旨）は次のとおりです。

```
interface → app → domain
infra     → app → domain
```

- **domain** は他レイヤーに依存しない（標準ライブラリ中心）。
- **app** は **domain** のみ参照する。
- **infra** は **app** の抽象（リポジトリ等）を実装する。
- **interface** は **app** のユースケースを呼び出す。

## まだ含まれていないもの（目標構成）

アプリケーション本体（FastAPI の `src/`、フロントエンド、`pyproject.toml` / `package.json` など）は**未配置**です。導入後に目指すスタック・ディレクトリ・コマンドの全体像は **`AGENTS.md`** に記載しています（Python 3.13 / FastAPI / uv、React / Vite / pnpm など）。

テストを通すには、`AGENTS.md` の `backend/src/` 配下のような構成に合わせて **`todo_app` パッケージ**（またはテスト内パスと整合する名前）を配置し、上記4層フォルダを作成したうえで、依存ルールに従って実装してください。

なお、`test_dependency_rule.py` が検査する import パス上のレイヤー名（例: `application`, `interfaces`, `infrastructure`）と、`test_source_structure.py` が期待するフォルダ名（`app`, `interface`, `infra`）は一致していません。スキャフォールド時は **`AGENTS.md` とテストの両方**を読み、必要ならテスト側をプロジェクトのパッケージ名・モジュール名に合わせて更新してください。

## ドキュメントの読み方

- **人間・ツール共通のプロジェクト概要・コマンド例** → [`AGENTS.md`](./AGENTS.md)
- **依存方向・禁止事項の短い要約** → [`.cursor/rules/architecture.md`](./.cursor/rules/architecture.md)

---

この README はリポジトリの実ファイル構成に基づいています。ソースやパッケージ名を追加したら、`todo_app` 参照とテストの期待値を自分のプロジェクト名に合わせて更新してください。
