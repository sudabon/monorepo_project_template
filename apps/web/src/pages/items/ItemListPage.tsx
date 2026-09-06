import { useQuery } from '@tanstack/react-query';
import { Link, useRouteContext } from '@tanstack/react-router';
import { Button } from '../../components/ui/button.tsx';
import { Input } from '../../components/ui/input.tsx';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from '../../components/ui/table.tsx';
import { ITEM_PAGE_SIZE, type ItemListSearch } from './itemListSearch.ts';

type Props = {
  search: ItemListSearch;
  onQueryChange: (q: string) => void;
  onPageChange: (page: number) => void;
};

export function ItemListPage({ search, onQueryChange, onPageChange }: Props) {
  const { clients } = useRouteContext({ from: '/_authenticated' });
  const query = useQuery(
    clients.items.list({ page: search.page, pageSize: ITEM_PAGE_SIZE }),
  );
  const items = query.data?.items ?? [];
  // TODO(template): 契約に検索クエリが無いため、表示中ページを名前で絞り込む。
  // サーバ側検索が必要なら OpenAPI の listItems に q を追加する。
  const visible = search.q
    ? items.filter((item) =>
        item.name.toLowerCase().includes(search.q.toLowerCase()),
      )
    : items;
  const total = query.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / ITEM_PAGE_SIZE));

  return (
    <main className="p-6">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-semibold">サンプルリソース</h1>
        <Link
          to="/items/new"
          className="inline-flex items-center justify-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground"
        >
          新規作成
        </Link>
      </div>
      <div className="mb-4 max-w-sm">
        <Input
          label="検索"
          value={search.q}
          onChange={(event) => onQueryChange(event.target.value)}
        />
      </div>
      {query.isPending ? <p>読み込み中</p> : null}
      {query.isSuccess && visible.length === 0 ? (
        <p>該当するサンプルリソースはありません</p>
      ) : null}
      {visible.length > 0 ? (
        <Table>
          <TableHead>
            <TableRow>
              <TableHeaderCell>名前</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {visible.map((item) => (
              <TableRow key={item.id}>
                <TableCell>
                  <Link
                    to="/items/$itemId"
                    params={{ itemId: item.id }}
                    className="underline"
                  >
                    {item.name}
                  </Link>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : null}
      <div className="mt-4 flex items-center gap-3">
        <p>ページ {search.page}</p>
        <Button
          disabled={search.page <= 1}
          onClick={() => onPageChange(search.page - 1)}
        >
          前のページ
        </Button>
        <Button
          disabled={search.page >= totalPages}
          onClick={() => onPageChange(search.page + 1)}
        >
          次のページ
        </Button>
      </div>
    </main>
  );
}
