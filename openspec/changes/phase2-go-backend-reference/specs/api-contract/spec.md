## ADDED Requirements

### Requirement: サンプルリソースの NUL 文字禁止

OpenAPI 契約は Item と ItemInput の name / description に NUL（U+0000）を含められないことを pattern と説明で明記しなければならない (MUST)。作成・更新で NUL を含む入力は、既存の ValidationError 形式と HTTP 422 で拒否することを明記しなければならない (MUST)。

#### Scenario: 入出力の文字列制約

- **WHEN** Item または ItemInput の name / description を契約に従って検証する
- **THEN** NUL を含む文字列は不正となり、文字数制約を満たす通常の改行や文字列としての `\u0000` は許可される

#### Scenario: NUL 入力へのエラー形式

- **WHEN** NUL を含む作成・更新リクエストへの契約上の応答を確認する
- **THEN** HTTP 422 と ValidationError の code / message / errors により、違反フィールドと理由を取得できる
