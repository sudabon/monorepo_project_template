# build-toolchain Specification

## Purpose
リポジトリの開発タスクをどう実行するかの契約を定める。開発者と CI が同じコマンドで同じ検証を実行できる状態を保証し、ツールチェーンのバージョン差異と依存更新の取り込みによる事故を防ぐ。

## Requirements

### Requirement: 単一のタスク入口

リポジトリのすべての開発タスクは `make` 経由で実行できなければならない (MUST)。CI が実行するコマンドは、すべてローカルから同一のターゲット名で実行できなければならない (MUST)。

提供するターゲットは最低限 `setup` / `gen` / `lint` / `fmt` / `test` / `build` / `check` とし、`check` は CI と同じ順序で他ターゲットを実行しなければならない (MUST)。

#### Scenario: clone 直後のセットアップと検証

- **WHEN** リポジトリを clone した直後の環境で `make setup && make check` を実行する
- **THEN** 追加の手動手順なしに両コマンドが終了コード 0 で完了する

#### Scenario: CI 手順のローカル再現

- **WHEN** CI ワークフローが実行するコマンド列を確認する
- **THEN** すべてのステップが `make <target>` の形式であり、同じターゲットをローカルで実行できる

#### Scenario: 未定義ターゲットの検出

- **WHEN** `setup` / `gen` / `lint` / `fmt` / `test` / `build` / `check` のいずれかを実行する
- **THEN** "No rule to make target" にならず、各ターゲットが定義済みとして実行される

### Requirement: 言語別のパス分離実行

CI は変更されたパスに応じて、実行する必要のないジョブをスキップしなければならない (MUST)。Go 側の検証、TypeScript 側の検証、API 契約の検証はそれぞれ独立したワークフローとして分離しなければならない (MUST)。複数言語を貫通して検証するブラウザ E2E はこの分離の対象外とし、貫通する側すべてのパスを対象にしなければならない (MUST)。

#### Scenario: Go のみを変更した場合

- **WHEN** Go のソースファイルだけを変更した Pull Request を作成する
- **THEN** Go 側のワークフローが実行され、TypeScript の単体検証とビルドを行うワークフローは実行されない

#### Scenario: Go のみの変更でもブラウザ E2E は実行される

- **WHEN** Go のソースファイルだけを変更した Pull Request を作成する
- **THEN** SPA から BFF と API まで貫通するブラウザ E2E のワークフローが実行される

#### Scenario: TypeScript のみを変更した場合

- **WHEN** TypeScript のソースファイルだけを変更した Pull Request を作成する
- **THEN** TypeScript 側のワークフローが実行され、Go 側のワークフローは実行されない

#### Scenario: 契約ファイルを変更した場合

- **WHEN** `api/` 配下のファイルを変更した Pull Request を作成する
- **THEN** 契約検証のワークフローが実行される

### Requirement: ツールチェーンバージョンの固定

Node.js のバージョンは `.node-version` と `package.json` の `engines.node` で完全一致で固定し、両者が一致していなければならない (MUST)。pnpm のバージョンは `packageManager` の `pnpm@<version>` で完全一致で固定しなければならない (MUST)。Go のバージョンは `go.mod` と CI の設定で一致していなければならない (MUST)。依存パッケージのバージョンは完全一致でピンし、`^` や `~` を使ってはならない (MUST NOT)。pnpm の `workspace:` プロトコルによるワークスペース内部参照は、このピン検査の対象外とする。lockfile はリポジトリにコミットしなければならない (MUST)。

#### Scenario: Node バージョンの一致

- **WHEN** `.node-version` と `package.json` の `engines.node` の値を照合する
- **THEN** 両者が同一の Node.js バージョンを指している

#### Scenario: pnpm バージョンの固定

- **WHEN** `package.json` の `packageManager` を確認する
- **THEN** `pnpm@<version>` の形式で pnpm 自身のバージョンが完全一致で指定されている

#### Scenario: Go バージョンの一致

- **WHEN** `go.mod` の `go` ディレクティブと CI の Go セットアップ設定を照合する
- **THEN** 両者が同一の Go バージョンを指している

#### Scenario: バージョンレンジ指定の排除

- **WHEN** `package.json` の依存宣言を確認する
- **THEN** レジストリ依存が完全一致のバージョンで指定されており、レンジ指定子を含まない。`workspace:` で始まる内部参照は許容される

### Requirement: 依存更新の取り込み方針

依存更新は自動化され、更新の影響度に応じて扱いを変えなければならない (MUST)。patch と minor は週次でグループ化し、CI 通過を条件に自動マージしてよい (MAY)。major は個別の Pull Request として起票し、自動マージしてはならない (MUST NOT)。対象は Go modules、npm、GitHub Actions のすべてを含まなければならない (MUST)。

#### Scenario: patch / minor の週次グループ化

- **WHEN** 依存更新ボットが patch および minor の更新を検出する
- **THEN** 週次でグループ化された Pull Request が作成され、CI が通過した場合のみ自動マージされる

#### Scenario: major は個別 PR

- **WHEN** 依存更新ボットが major の更新を検出する
- **THEN** 個別の Pull Request が作成され、自動マージされない

#### Scenario: 更新対象エコシステムの網羅

- **WHEN** 依存更新の設定を確認する
- **THEN** Go modules、npm、GitHub Actions のすべてが更新対象に含まれ、lockfile maintenance が有効になっている
