# 0003: API の DB、マイグレーション、依存方向検査

- 状態: 採用
- 日付: 2026-09-06

## 判断

PostgreSQL と pgx v5.10.0 の database/sql ドライバを採用する。SQL を直接書き、
domain は標準ライブラリのみ、repository が SQL と DB エラー変換を所有する。
参考リポジトリの handler / usecase / domain / repository の責務を apps/api/internal
に配置し、起動時の組み立ては cmd/api に置く。

マイグレーションは goose v3.28.0。SQL の Up / Down とトランザクションを扱え、
単一バイナリで動く。必要な PostgreSQL ドライバだけを含む cmd/migrate をビルドし、
SQL を埋め込む。同じ goose Provider を統合テストでも使う。
golang-migrate も要件を満たすが、今回は 1 ファイルに Up / Down を記述できることと
Provider API によるテスト単位の分離を優先した。Atlas の宣言的差分生成は今回不要。
本番 API の起動で自動移行せず、デプロイ前に移行コマンドを別途実行する。

依存方向検査は go-arch-lint v1.18.0 を Go tool directive で固定し、make lint-go
に含める。宣言的な component と依存許可リスト、未知の package の検出、違反箇所の
診断を利用する。go list の自作スクリプトは実装・保守が増え、depguard は import の
禁止には向くが層全体の許可グラフは go-arch-lint のほうが明示的になる。
テストによる import 検査は対象列挙の漏れを避けるため採用しない。
ツールの推移依存は多いが API バイナリにはリンクされない。
deepScan は無効とし import 方向と外部ライブラリの許可を検査する。deepScan は
DI でインタフェースに渡す具象 adapter を逆向き依存と判定するため、今回の
composition root と生成 Echo 登録に適合しない。domain から外側への import を
実際に追加して失敗することを確認する。

go-arch-lint の graph 機能のテスト依存が古い genproto の一体型モジュールを参照する。
go mod tidy 時の分割版 RPC モジュールとの衝突を避けるため、親 genproto を
分割後の `v0.0.0-20260831171406-18b4a7587f8a` に揃える。

ローカルの DB は Compose、CI は PostgreSQL サービスコンテナ。
make test-go (make test からも呼ばれる) は TEST_DATABASE_URL がなければ Compose を
起動し、設定済みならその DB を利用する。各テストは UUID 付きの独立 schema を作成して
終了時に削除するため、テスト間でテーブルの truncate や共有データの削除はしない。
ただし専用のテスト DB を指定すること。testcontainers は Docker API 用の Go 依存と
環境差を増やすため採用しない。DB が利用不能ならテストは失敗し、黙って skip しない。

参考: [goose](https://github.com/pressly/goose)、
[go-arch-lint](https://github.com/fe3dback/go-arch-lint)、
[pgx](https://github.com/jackc/pgx)、
[参照リポジトリ](https://github.com/sudabon/monorepo_go_project_template)。
