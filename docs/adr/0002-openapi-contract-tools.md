# 0002: OpenAPI 契約の lint と生成

- 状態: 採用
- 日付: 2026-09-06

## 判断

契約は `api/openapi.yaml` の OpenAPI 3.0.3 を手書きし、Spectral 6.16.3 の
`spectral:oas` で検査する。既存の pnpm workspace から実行でき、推奨ルールと
案件固有のルールを同じ設定に置けるため採用する。Redocly も候補だが、今回は
ドキュメントサイトや bundle を必要としないため追加しない。

無効化するルールは `info-contact` のみ。テンプレートに実在しない問い合わせ先を
記載しないためで、案件側で所有者が決まったら有効にする。その他の推奨ルールを
維持し、warning も `--fail-severity warn` で失敗扱いにする。

Go は `go.mod` の tool directive で固定した oapi-codegen 2.8.0 を `go tool` から
呼び、Echo v4 の通常・strict サーバインタフェースと型を生成する。strict の
リクエスト・レスポンス型は Phase 2 の実装漏れをコンパイルで検出するために使う。
ランタイムの入力検証と共通エラーへの変換は Phase 2 の責務。

TypeScript は openapi-typescript 7.13.0 の生成型を openapi-fetch 0.17.0 に渡す。
openapi-typescript の peer dependency が TypeScript 5.x のため、TypeScript は
5.9.3 を固定する。生成物は両言語とも Git 管理し、手書きコードとは分離する。
`--default-non-nullable false` を指定して、default 値を持つ省略可能なリクエスト項目が
TypeScript だけ必須になることを防ぐ。省略可能性は OpenAPI の required に従う。

## 根拠とトレードオフ

生成物の量は増えるが、仕様変更の影響をレビューできる。生成器・設定・依存の変更も
契約 CI の対象にし、ローカルと CI で同じ Make ターゲットから生成する。
生成物の欠落を見落とさないよう、差分検査では追跡済みファイルと未追跡ファイルの
両方を確認する。

参考: [Spectral rulesets](https://docs.stoplight.io/docs/spectral/4b182c85c0ce0-rulesets)、
[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/tree/v2.8.0)、
[openapi-typescript](https://openapi-ts.dev/introduction)、
[openapi-fetch](https://openapi-ts.dev/openapi-fetch/)。
