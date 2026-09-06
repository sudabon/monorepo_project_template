# 未完了の妥協点（`TODO(template)`）

テンプレートが意図的に手を抜いた箇所の一覧。コード側のマーカーは `// TODO(template):`。

扱いは 2 種類しかない。

- **案件で必ず解消**: 放置すると本番で壊れるか、危険な状態のまま出る
- **テンプレートに残す**: 案件の要件次第。不要なら消さずに残してよい

一覧を更新したら、コード側のマーカーと同時に直す。片方だけ消さない。

## 案件で必ず解消

| 場所 | 内容 | 期限 | 放置すると |
| --- | --- | --- | --- |
| `apps/bff/internal/identity/static.go` | デモユーザ 1 名の `Static` を IdP かユーザストアに差し替える | **本番公開前** | `BFF_DEMO_USERNAME` / `BFF_DEMO_PASSWORD` を知る誰でもログインできる。[README](../README.md) のチェックリスト 6 |
| `apps/bff/internal/session/postgres.go` | 期限切れセッション行の定期削除 | 運用開始前 | `sessions` が単調に増える。`Get` は期限切れを不在として扱うので認証は壊れないが、テーブルが肥大する |

## テンプレートに残す

| 場所 | 内容 | 判断の材料 |
| --- | --- | --- |
| `apps/web/src/pages/items/ItemListPage.tsx` | 契約に検索クエリが無いため、一覧の検索は表示中ページ内の絞り込み | 全件検索が要るなら `api/openapi.yaml` の `listItems` に `q` を足してサーバ側検索にする。公開契約の変更なので承認が要る |
| `apps/web/src/components/ui/modal.tsx` | 背景スクロールの固定 | モーダルの背後がスクロールしても業務上困らないなら残してよい |
| `apps/web/src/components/ui/modal.tsx` | opener が閉じる前に unmount した場合のフォーカス復帰 | 一覧から開いて行が消えるような画面を作るなら埋める |
| `apps/web/src/components/ui/toast.tsx` | 緊急時のみトーストへフォーカスを移す | スクリーンリーダー対応の要件がある案件で埋める |
| `apps/web/src/components/ui/toast.tsx` | 認証・セッション断のトーストを `aria-live="assertive"` にする | 同上 |

UI 部品を自前で持つ判断は [ADR 0010](adr/0010-copy-owned-ui-components.md)。
アクセシビリティの実装が自前になることは、その判断の対価として想定済み。

## 棚卸し

```sh
rg -n 'TODO\(template\)' --glob '!docs/**'
```

この一覧に無いマーカーが出たら、追加するか解消する。マーカーを増やすときは
理由と「放置すると何が起きるか」を書く。書けないものは妥協ではなく、ただの未実装。
