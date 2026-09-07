## Why

テンプレートの価値は最初の 1 本ではなく、**2 本目のリソースを足す速さ**で決まる。現状その動線に 3 つの障害がある。

1. `apps/api/internal/handler/items.go` の `readInput` は 60 行あり、Content-Type 検査・多重 JSON 検出・null と欠損の区別・型エラーと業務エラーのマージ・契約順への並べ替えを 1 関数で行っている。中身は正しいが `"name"` / `"description"` が 2 か所にリテラルで埋まっており、**リソースを増やすたびに全体をコピーする必要がある**
2. 「name 1〜100 文字、description 2000 文字以内、NUL 禁止」が `api/openapi.yaml`・`domain/item.go`・`itemInputSchema.ts` の 3 か所に手書きされている。原則「契約は 1 か所にしか置かない」に対する唯一の破れで、リソースが増えれば破れも増える
3. `queryClient.invalidateQueries({ queryKey: ['items'] })` が 3 ページで手書きされている。キーを組み立てるのは `packages/api-client` の責務なのに、無効化側だけが文字列で繋がっている。キー形状を変えても型エラーにならず、無効化が黙って効かなくなる

## What Changes

- `readInput` を、資源に依存しない `decodeFields`（JSON 解析・null 判定・型エラー収集・契約順の整列）と、資源固有の薄い変換に分ける。リソース追加時に書くのはフィールド名の列と domain 型への詰め替えだけにする
- 契約の値制約とサーバ／クライアント実装の検証が、ずれたまま通らないようにする。**検証ロジックそのものは各言語に残す**（Go の生成コードは検証を持たず、zod は UI の即時応答に要る）。共有するのは制約の値だけとし、手段は design で決める
- `packages/api-client` がキャッシュ無効化に使うキーを公開し、SPA から文字列リテラルを消す
- `itemLoadMessage` を `ItemDetailPage.tsx` から共有モジュールへ移す。現在は `ItemEditPage` がページからページを import しており、真似されると横依存が広がる
- `api/oapi-codegen.yaml` の `strict-server: true` の扱いを決める。現在 `StrictServerInterface` は生成されるが 1 か所も使われておらず、生成物のどちらが正か読み手が判断できない
- JSON の読み書きに使う標準ライブラリを 1 つに決める。現在 BFF は `encoding/json/v2`、API の `readInput` は `encoding/json`（v1）を使っており、テンプレートが 2 通りの手本を示している
- `apps/api/AGENTS.md` に「リソースを 1 つ追加するときに書くもの」を手順として明記する

## Capabilities

### New Capabilities

なし。

### Modified Capabilities

- `api-contract`: 契約の値制約と実装側検証の一致を新しい要求として加える。あわせて「生成物とアプリケーション層の分離」に、キャッシュ無効化キーもラッパが公開するという要求を加える

`readInput` の分割、`itemLoadMessage` の移動、`strict-server` の決着は実装の整理であり、観測可能な振る舞いは変わらない。

## Impact

- **変更**: `apps/api/internal/handler/items.go`、新規 `apps/api/internal/handler/decode.go`、`packages/api-client/src/items.ts` と `index.ts`、`apps/web/src/pages/items/*.tsx`、新規 `apps/web/src/pages/items/messages.ts`、`api/oapi-codegen.yaml`（決定次第）、`apps/api/AGENTS.md`
- **新規テスト**: 契約と実装の制約一致を検証するテスト（`apps/web` 側と Go 側の境界値テーブル）
- **依存追加**: なし。制約の抽出は既に direct require の `kin-openapi` で行う
- **`apps/bff`**: JSON スタックの決定が v1 側に倒れた場合のみ `internal/handler` の `jsonSerializer` を変更する（design D5）
- **生成物**: `strict-server` を外す判断をした場合、`apps/api/internal/generated/api.gen.go` が縮小する。`make gen-check` で差分を確認する
