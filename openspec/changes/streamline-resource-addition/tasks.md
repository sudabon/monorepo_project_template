## 1. 契約から制約を生成する

- [x] 1.1 `api/openapi.yaml` の `ItemInput` から `minLength` / `maxLength` / `pattern` を抽出する generator を `apps/api` に追加し、`go tool` で実行できることを確認する（design D1）
- [x] 1.2 generator が `apps/api/internal/domain/constraints.gen.go` と `packages/api-client/src/generated/constraints.ts` を出力し、2 回連続実行しても差分が出ない（冪等）ことを確認する
- [x] 1.3 `make gen-go` / `make gen-web` に生成を組み込み、`GENERATED_GO_PATHS` / `GENERATED_WEB_PATHS` に出力先を追加する
- [x] 1.4 `domain.ItemInput.Validate()` の 1 / 100 / 2000 と NUL を生成物参照に置き換え、既存の `item_test.go` が変更なしで通ることを確認する
- [x] 1.5 `itemInputSchema.ts` の 100 / 2000 と NUL を生成物参照に置き換え、既存の `itemInputSchema.test.ts` が変更なしで通ることを確認する
- [x] 1.6 契約の `maxLength` だけを変更して `make gen` を実行せずに `make gen-check` を走らせると失敗することを確認し、確認後に契約を戻す
- [x] 1.7 契約の `maxLength` を変更して `make gen` を実行すると、Go と TS の両方の境界値が同時に動くことを確認し、確認後に戻す

## 2. リソース追加の動線

- [x] 2.1 `apps/api/internal/handler/decode.go` に `decodeFields(c, order)` を切り出し、既存の `items_test.go` と `contract_test.go` を変更せずに通すことを確認する（design D2）
- [x] 2.2 契約順への並べ替えを二重ループから `order` のインデックス比較に置き換え、複数フィールドが同時に不正なときのエラー順が契約どおりであることをテストで確認する
- [x] 2.3 `handler/items.go` に残る資源固有のコードが、フィールド名の列と `domain.ItemInput` への詰め替えだけになっていることを確認する
- [x] 2.4 `decodeFields` を `encoding/json/v2` で書き、JSON null とフィールド欠損の区別が v2 で表現できることを確認する。表現できない場合は BFF を含めて v1 に揃え、理由を design D5 に追記する
- [x] 2.5 `apps/{api,bff}` で使われている JSON パッケージが 1 つになったことを grep で確認する
- [x] 2.6 `api/oapi-codegen.yaml` から `strict-server: true` を外し、`make gen` 後に `StrictServerInterface` が生成されないこと、`make build-go` と `make test-go` が通ることを確認する（design D4）

## 3. SPA 側の結び直し

- [x] 3.1 `packages/api-client` から `itemKeys` を export し、`createItemQueries` の `queryKey` がそれを使うように変更する（design D3）
- [x] 3.2 `ItemCreatePage` / `ItemEditPage` / `ItemDetailPage` の `queryKey: ['items']` を `itemKeys.all` に置き換え、文字列リテラルが 0 件になったことを確認する
- [x] 3.3 `itemKeys.all` の形を一時的に変更すると SPA 側が型検査で失敗することを確認し、確認後に戻す
- [x] 3.4 `itemLoadMessage` を `apps/web/src/pages/items/messages.ts` に移し、ページ間の import が 0 件になったことを確認する（design D6）

## 4. 手順の明文化と検証

- [x] 4.1 `apps/api/AGENTS.md` に「リソースを 1 つ追加するときに書くもの」を、契約・handler・usecase・repository・migration の順で手順として明記する
- [x] 4.2 `apps/web/AGENTS.md` に、キャッシュ無効化キーは `api-client` から取ること、ページ間で import しないことを明記する
- [x] 4.3 `packages/api-client/AGENTS.md` に、キーの公開はラッパの責務であることを明記する
- [x] 4.4 2 つ目のリソースを実際に追加して、`decodeFields` の抽象が過不足ないこと、書くコードがフィールド名の列と詰め替えだけで済むことを確認する。確認後に追加分を取り除き、判明した過不足を design に追記する（design Risks）
- [x] 4.5 `make check` と `git diff --check` が通ることを確認する
