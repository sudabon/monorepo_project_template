# 0004: BFF セッションストア

- 状態: 採用
- 日付: 2026-09-06

## 判断

セッション実体はサーバ側に置き、識別子だけを Cookie に入れる（design D1）。
保存先は backend と同じ PostgreSQL を流用する。専用の Redis 等は初期値にしない。

テンプレートの BFF は `sessions` テーブルを goose で管理し、version テーブルは
`bff_goose_db_version` にして API の migration と衝突させない。

## 比較

| 候補 | 複数 ECS タスク | デプロイ後の存続 | 追加インフラ | TTL |
| --- | --- | --- | --- | --- |
| プロセス内メモリ | 満たさない | 満たさない | なし | 実装次第 |
| PostgreSQL（API と共用） | 満たす | 満たす | なし（既存） | `expires_at` と取得時の延長 |
| Redis 等の専用ストア | 満たす | 満たす（永続化設定次第） | あり | ネイティブ TTL |

プロセス内メモリは D1 の要件（複数タスク・デプロイ）を満たさない。Redis は
セッション向きだが、Compose・CI・運用にサービスを 1 つ足す。PostgreSQL は
既に API のローカルと CI で動いており、依存を増やさない。

## 判断基準（案件で見直すとき）

- セッション数が DB の負荷として無視できない、または TTL 切れの掃除を DB に
  寄せたくない → Redis 等の専用ストアを検討する
- API と BFF で DB 障害を共有したくない → 別インスタンスまたは専用ストア
- テンプレートの初期値は「既存 PostgreSQL で足りる」側に置く

単体のストア契約はメモリ実装でも検証する。プロセスの本番経路は PostgreSQL だけを使う。

参考: [ADR 0003](0003-api-database-and-architecture.md)、[docs/bff.md](../bff.md)。
