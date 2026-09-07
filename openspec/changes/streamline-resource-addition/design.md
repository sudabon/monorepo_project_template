## Context

背景と 3 つの障害は proposal.md - Why を参照。要求は specs/api-contract を参照。

前提として、このリポジトリには既に「契約から生成し、生成物をコミットし、`make gen-check` でズレを落とす」という仕組みがある（`GENERATED_GO_PATHS` / `GENERATED_WEB_PATHS`）。新しい仕組みを足すより、この既存の経路に乗せられるかで手段を選ぶ。

## Goals / Non-Goals

**Goals**

- リソースを 1 つ追加するときに書くコードを、フィールド名の列と domain 型への詰め替えだけにする
- 契約の値制約がずれた状態を CI で落とす
- 生成物のどれが正なのかを 1 つに決める

**Non-Goals**

- サーバ検証とクライアント検証を 1 つの実装に統合すること。用途が違う（サーバは信頼境界、クライアントは即時応答）
- 検索クエリの契約追加。これは別途ユーザ承認が要る公開契約の変更で、このチェンジの範囲外
- 新しい依存の追加

## Decisions

### D1: 制約の値だけを契約から生成し、両言語がそれを参照する

**決定**: `api/openapi.yaml` から入力制約（`minLength` / `maxLength` / `pattern`）を抽出し、`make gen` で 2 つの生成物を出す。

- `apps/api/internal/domain/constraints.gen.go`
- `packages/api-client/src/generated/constraints.ts`

`domain.ItemInput.Validate()` と `itemInputSchema.ts` は、数値と禁止文字をこの生成物から取る。エラーメッセージと検証の書き方は各言語に残す。生成物は `GENERATED_GO_PATHS` / `GENERATED_WEB_PATHS` に加え、既存の `make gen-check` にそのまま乗せる。

**理由**: 「契約だけ変えて実装を直し忘れる」を、新しい検査ではなく**既にある検査**で落とせる。突き合わせテストを別に書くより仕組みが 1 つ少ない。

**代替案**: (a) 突き合わせテストを書く → Go 側は kin-openapi があるので書けるが、TS 側は YAML パーサの依存が増える。(b) zod を契約から丸ごと生成する → 生成器の依存が増え、エラーメッセージを日本語にできない。(c) 現状維持 → リソースが増えるほど破れが増える。

**抽出器の置き場所**: `apps/api` に小さな generator を置き、`go tool` で実行する。`kin-openapi` は既に direct require で、YAML の解釈は生成コードと同じライブラリになる。Node 側に依存を足さない。

### D2: `readInput` を資源非依存の `decodeFields` と資源固有の詰め替えに分ける

**決定**: `apps/api/internal/handler/decode.go` に次を置く。

```go
// order は契約に現れるフィールド順。エラーもこの順で返す。
func decodeFields(c echo.Context, order []string) (map[string]*string, domain.ValidationErrors, error)
```

`items.go` に残るのは、フィールド名の列と `domain.ItemInput` への詰め替えだけにする。現在の並べ替えは二重ループ（O(n²)）なので、`order` のインデックス比較に置き換える。

**理由**: Content-Type 検査・サイズ制限・多重 JSON 検出・null と欠損の区別は、どのリソースでも同じ。ここを共有しないと、2 本目のリソースでこの 60 行が丸ごと複製される。

**代替案**: リフレクションで汎用バインダを書く → 生成型ごとの詰め替えが暗黙になり、契約とのズレがコンパイル時に出なくなる。採らない。

### D3: `api-client` がキャッシュ無効化キーを公開する

**決定**: `createItemQueries` / `createItemMutations` と同じモジュールから `itemKeys` を export し、`{ all: ['items'], list: (params) => [...], detail: (id) => [...] }` の形にする。SPA は `queryClient.invalidateQueries({ queryKey: itemKeys.all })` を使う。

**理由**: キーの組み立ては生成型を知っている側の責務。現在は無効化側だけが文字列で繋がっており、キー形状を変えても型エラーにならない。

### D4: `strict-server` は外す

**決定**: `api/oapi-codegen.yaml` から `strict-server: true` を外す。

**理由**: `StrictServerInterface` は生成されているが 1 か所も使われていない。かつ、採用すると生成コードの既定バインドがボディを読むため、**null と欠損の区別や型エラーのフィールド単位収集ができなくなる**。契約が要求する ValidationError の粒度（どのフィールドがなぜ不正か）を満たせない。使わないものを生成し続けると、生成物 874 行のどちらが正か読み手が判断できない。

**代替案**: strict を採用して `readInput` を捨てる → 上記のとおり契約の要求を満たせなくなる。両方生成したまま残す → 現状維持で、読み手の迷いが残る。

### D5: JSON スタックは `encoding/json/v2` に揃える

**決定**: `decodeFields` は `encoding/json/v2` で書く。BFF は既に v2（`jsonSerializer`）を使っており、API の `readInput` だけが v1 だった。

**理由**: 判断の基準は速度でも機能でもなく、**テンプレートが 2 通りの手本を示さない**こと。Go 1.27 の標準ライブラリに v2 が入っており、新規に書くコードの既定を v1 に戻す理由がない。

**確認が要る点**: 現在の `readInput` は `map[string]json.RawMessage` で「JSON null」と「フィールドの欠損」を区別し、これが 422 のフィールド単位エラーの前提になっている。v2 で同じ区別が書けることを実装時に確認する。書けない場合は両者を v1 に揃え、その理由を本 design に追記する（どちらに倒すにせよ 1 つにする、が決定の本体）。

**代替案**: v1 に揃える → BFF の `jsonSerializer` を書き換えることになり、変更範囲が広がる。現状維持 → 2 通りの手本が残る。

### D6: 画面間で共有するメッセージは専用モジュールに置く

**決定**: `itemLoadMessage` を `apps/web/src/pages/items/messages.ts` に移す。`ItemDetailPage` と `ItemEditPage` の双方がそこから import する。

**理由**: 現在は `ItemEditPage` が `ItemDetailPage` を import している。テンプレートの参照実装がページ間の横依存を示すと、案件側でそれが増える。

## Risks / Trade-offs

- **生成ステップが 1 つ増える** → `make gen` に載せ、`make gen-check` の既存経路で検査する。新しいコマンドは増やさない
- **generator 自体がテンプレートの保守対象になる** → 抽出するのは `minLength` / `maxLength` / `pattern` の 3 つに限り、汎用な OpenAPI → コード変換にはしない。汎用化が必要になったら `oapi-codegen` 側の機能を待つ
- **`strict-server` を外すと将来採用したくなったとき戻す手間がある** → 設定 1 行なので戻せる。判断の理由を ADR ではなく本 design と `apps/api/AGENTS.md` に残す
- **`decodeFields` の抽象が早すぎる可能性** → 2 本目のリソースを実際に足して確認する。タスクに含める

## Open Questions

なし。検索クエリの契約追加は Non-Goals に置いた別件で、このチェンジの判断には影響しない。
