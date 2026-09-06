## Context

Phase 0 で `make` を入口とするタスク定義と 3 本の CI ワークフローが用意されている。`make gen` は空ターゲット、`contract.yml` は起動するだけで中身がない状態。本フェーズはその 2 つに実体を入れる。

制約は次の 3 点。

- 契約は 1 か所にしか置かない。Go と TypeScript の境界は OpenAPI 仕様ファイルのみ
- 生成物はコミットする。ビルド時生成に頼らず、リポジトリを見ただけで型が読める状態を保つ
- 依存は少ないほど良い。生成ツールを増やすほど、メジャー更新のたびに生成物が揺れる

動機は proposal.md - Why を参照。要求は specs/api-contract/spec.md を参照。

## Goals / Non-Goals

**Goals:**

- 仕様ファイルを唯一の入力として、Go の Echo サーバインタフェースと TypeScript の型付きクライアントを生成する
- 生成を冪等にし、仕様と生成物のずれを CI で機械的に落とす
- 生成物とアプリケーションコードの間に、手書きの薄い層を 1 枚だけ置く

**Non-Goals:**

- モックサーバ、ドキュメントサイトの生成（必要になった案件側で足す）
- 契約からのバリデーション実装の自動生成（Go 側のバリデーションは Phase 2 で扱う）
- 複数バージョンの API 併存（`/v2` 等）。テンプレートの範囲外
- gRPC / GraphQL など他の契約形式

## Decisions

### D1: 仕様は手書き、実装からは生成しない

`api/openapi.yaml` を人が書く。Go のコメントアノテーションから仕様を生成する方式（swaggo 等）は採らない。実装の都合が契約に漏れ、クライアント側の変更が「Go を直さないと始まらない」構造になるため。仕様を先に書き、両言語をそこから生成する。

- 代替案: Go から生成 → 契約が実装に引きずられる。却下
- 代替案: TypeSpec 等の中間 DSL から OpenAPI を生成 → 依存とビルド段数が 1 つ増える。OpenAPI を直接書けば足りる。却下

### D2: Go は oapi-codegen、TypeScript は openapi-typescript + openapi-fetch

Go 側は Echo 用のサーバインタフェースと型を生成する。TypeScript 側は型定義と、型を利用する薄い fetch ラッパに分ける。ランタイムを持つ重いクライアント生成器（openapi-generator 等）は使わない。生成物の読みやすさと、生成器のメジャー更新時の壊れにくさを優先する。

- 代替案: openapi-generator の typescript-fetch → 生成物が大きく、実行時ライブラリへの依存が増える。却下

### D3: 生成物は「編集しない領域」として隔離し、手書きラッパを 1 枚被せる

`packages/api-client` を「生成物ディレクトリ」と「手書きラッパ」に物理的に分ける。ラッパは TanStack Query の `queryOptions` を返す関数だけを公開し、アプリケーションは生成物を直接 import しない。

- これにより、生成器を差し替えても影響範囲がラッパ層に閉じる
- ラッパは薄く保つ。リトライ方針・エラーハンドリングは Phase 4 の QueryClient 側で決め、ここには書かない
- 生成物ディレクトリには編集禁止である旨をファイル先頭コメントで明示する

HTTP 応答の受け渡し: openapi-fetch は非 2xx でも resolve するため、通信層で
`ApiError` に status と契約の body を保持して throw し、Query に失敗を伝える。
ここでは再試行、401 の遷移、トースト、フィールドへの表示、キャッシュ無効化を
判断しない。これらは Phase 4 の QueryClient・フォーム側で扱う。

### D4: エラー形式は「共通エラー」と「バリデーションエラー」の 2 層

すべての操作が参照する共通エラースキーマを 1 つ定義し、バリデーションエラーはその拡張として、フィールド識別子とエラー内容の組を配列で返す形にする。Phase 3 の 401、Phase 4 のフォームエラーマッピング、Phase 5 のサーバ側バリデーション表示がすべてこの形式に依存するため、ここでの形が後段を決める。

- フィールド識別子は、クライアントのフォーム項目名と機械的に対応付けられる表現にする（ネストしたフィールドを指せること）

実装時確認: 共通部分は `{ code, message }`、422 は `allOf` で共通部分を参照し、
`errors: Array<{ field, message }>` を追加する。`field` は JSON のプロパティ名を
ドットで連結し、配列は `details.0.name` とする。プロパティ名自体にはドットを許さない。
Phase 4 の react-hook-form と同じパス表現なので、次の変換で機械的に対応できる。

```text
group errors by field, preserving every message
for each (field, messages):
  if field is a known registered form path:
    setError(field, { type: "server", message: messages.join("\n") })
  else:
    append messages to form-wide errors
display form-wide errors without discarding unknown fields
```

`name`、`metadata.owner.name`、`details.0.name` のいずれも同じ処理になる。
未知のフィールドは Phase 4 D7 に従ってフォーム全体に表示する。今回の Item の
フィールドはフラットだが、エラー形式はネスト・配列を追加しても変更不要と確認した。

### D5: 差分検出は `make gen && git diff --exit-code` を CI で実行する

生成し忘れを検出する仕組みはこれ 1 つに集約する。生成ツールのバージョンが CI とローカルで異なると偽陽性になるため、生成ツールのバージョンもピンして固定する（Go 側は tools 用の依存管理、TS 側は devDependencies）。

### D6: 生成物は整形ツールの対象から外す

`biome.json` の `files.includes` は `apps/**` と `packages/**` を含むため、そのままだと TypeScript の生成物も整形・lint の対象になる。生成物が整形結果と一致しなければ `make fmt-check-web` が落ち、整形すれば `make gen-check` の冪等性と衝突する。生成物のパスを `files.includes` から除外し、整形の責務を手書きコードに限定する。

- 生成物の可読性は生成器の出力品質に委ねる。整形して差分を作るより、生成器の設定で調整する

### D7: 契約 lint のルールセットは最小から始める

Spectral の推奨ルールセットを基点にし、テンプレートとして邪魔になるルールだけを個別に無効化する。案件側でルールを足す前提なので、初期から厳しくしすぎない。

## Risks / Trade-offs

- **生成ツールのバージョン差で CI が偽陽性になる** → 生成ツールのバージョンを完全一致でピンし、CI とローカルで同じ経路（`make gen`）から呼ぶ。D5 の前提
- **生成物のコミットで PR の差分が大きくなる** → レビュー時のノイズは受け入れる。`.gitattributes` で生成物を `linguist-generated` として扱い、差分表示を折りたためるようにする
- **エラー形式を後から変えると全フェーズに波及する** → D4 の形を Phase 3〜5 の利用側と突き合わせてから確定させる。本フェーズの実装時に、フォームエラーへのマッピングが機械的に書けるかを机上で確認する
- **oapi-codegen が Echo のメジャー更新に追随しない可能性** → 生成物がコミットされているため、追随できない期間も生成物の手動調整で凌げる。ただし手編集は D3 に反するため、その場合は生成器の差し替えを ADR に残して判断する
- **仕様と Go 実装の乖離（生成インタフェースを実装せずに握りつぶす）** → Phase 2 で生成インタフェースの実装漏れがコンパイルエラーになる構成にする
