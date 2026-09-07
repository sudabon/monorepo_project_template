## Context

`packages/go-platform` は ADR 0006 で作られ、Echo に依存しない 3 パッケージ（`logging` / `server` / `database`）だけを持つ。今回移す対象のうち health ルートと `traceRequest` は Echo の型（`echo.MiddlewareFunc`、`echo.Context`）に触れるため、この境界を跨ぐかどうかが最初の判断になる。

背景は proposal.md - Why を参照。

## Goals / Non-Goals

**Goals**

- 同一実装の複製をなくし、片側だけが直る事故を構造的に防ぐ
- 共有先の粒度を、案件側が触らずに済む範囲に留める

**Non-Goals**

- 業務ロジックの共有。`usecase` / `domain` / `repository` は各サービスに閉じたままとする
- go-platform を汎用ライブラリにすること。**このテンプレートの 2 サービスが共有する配線だけ**を置く
- エラーコード語彙の統一。サービスごとに異なるのが正しい

## Decisions

### D1: Echo 依存は `go-platform/echox` サブパッケージに閉じる

**決定**: `traceRequest` と health ルートは `packages/go-platform/echox/` に置く。`logging` / `server` / `database` は Echo 非依存のまま維持する。

**理由**: Echo はスタック表の確定事項であり、両サービスが既に使っている。一方で、go-platform 全体が Echo を要求すると「DB プールだけ使いたい」サービスが Echo を引き込むことになる。パッケージを分ければ、import しないものはバイナリに入らない。

**代替案**: (a) `net/http` の `http.Handler` middleware として書き、Echo 非依存にする → Echo の `HTTPErrorHandler` と `c.Path()` に触れないためログの内容が落ちる。(b) 各サービスに残す → 現状維持であり、既に事故が 2 件起きている。

### D2: `traceRequest` は BFF 側の実装を正とする

**決定**: 統一実装は BFF 側に揃える。すなわち `c.Request().Clone(ctx)` でリクエストを複製し、**リクエストヘッダにも** `X-Request-ID` を書く。API 側の `WithContext` だけの実装は捨てる。

**理由**: BFF はプロキシとして下流にヘッダを渡す必要があり、そのためにヘッダを書いている。API は現在下流を持たないが、将来別サービスを呼ぶときに同じ形であるほうがよい。逆に API 側の `debug.Stack()` を含む panic 整形は BFF に無い改善なので、こちらは API 側を採る。**片側ずつの改善を両取りする。**

### D3: エラー応答は封筒だけ共有し、対応表は共有しない

**決定**: `echox.WriteError(c, status, code, message)` として JSON の書き出し・`Committed` 判定・5xx のログだけを共有する。ステータスからコード語彙への変換は各サービスの `errors.go` に残す。

**理由**: API は `domain.ValidationErrors` を 422 に写す必要があり、BFF は `csrf_rejected` / `payload_too_large` を持つ。共通化すると、どちらのサービスにも属さないコードが共有側に溜まる。**共有してよいのは「どのサービスでも同じ」ものだけ**という線を最初に引いておく。

### D4: `testdb` は `testing` を import するので独立パッケージに置く

**決定**: `packages/go-platform/testdb/` として独立させ、`Open(t *testing.T, newProvider func(*sql.DB) (*goose.Provider, error)) *sql.DB` の形にする。

**理由**: 本番コードが import しない限りバイナリに入らない。`migrations` の embed FS はアプリ固有なので、provider を関数で受け取る。統合の際にコンテキストは `t.Context()` に揃える（API 側の `context.Background()` は取り残された古い書き方）。

### D5: UUID は標準ライブラリを既定とする

**決定**: リクエスト ID の生成は `go-platform` 内に閉じ、そこでは Go 1.27 標準の `uuid` を使う。`github.com/google/uuid` は `apps/api/internal/handler` の生成型変換だけに残す。

**理由**: 生成コードが `google/uuid.UUID` を返す以上、API から完全には外せない。しかし「新しく書くコードはどちらを使うのか」が現状 2 通りに読めるため、既定を 1 つに決める。`docs/STACK.md` の google/uuid 行も用途を限定した記述に直す。

あわせて `toItem` の `uuid.MustParse` を `uuid.Parse` + エラー返却に変える。DB の列型が uuid なので実際には失敗しないが、**テンプレートがリクエスト経路で `Must*` を使う手本を示すべきではない**。

### D6: arch-lint は `anyVendorDeps` をやめて `canUse` を明示する

**決定**: `handler` / `main` / `migrations` / `testdb` の `anyVendorDeps: true` を、BFF の `proxy` と同じ `canUse:` の列挙に置き換える。

**理由**: 現状これらのコンポーネントは外部依存が無検査で、「依存方向を lint で保証する」という売り文句の範囲が実際には狭い。依存を足すたびに設定更新が要るのはコストだが、それが検査の目的である。

## Risks / Trade-offs

- **go-platform が肥大化し、案件側が頻繁に触る場所になる** → 置いてよいのは「両サービスに同じ形で存在する配線」だけ、という基準を `packages/go-platform/AGENTS.md` に明記する。業務ロジックと片側だけの都合は入れない
- **共有化で個別最適の余地が減る** → D3 のとおり、サービスごとに違う部分（エラー語彙）は最初から共有対象外に置く
- **移動により arch-lint と go.mod の整合が崩れる** → 各ステップで `make lint-go` と `make test-go` を実行し、緑を保ったまま進める
- **`canUse` の明示で依存追加時の摩擦が増える** → 意図した摩擦。追加理由を PR に書く運用と噛み合う

## Migration Plan

すべて既存テストが回帰検査になる。振る舞いを変えないため、段階ごとに `make check` を通しながら進める。

1. `go-platform` に新パッケージを追加する（この時点では誰も使わない）
2. BFF を新パッケージに切り替える
3. API を新パッケージに切り替える
4. 旧実装を削除し、arch-lint を締める

各段階で `make test-go` が緑であること。ロールバックは段階単位で `git revert` できる。
