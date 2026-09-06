import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useRouteContext } from '@tanstack/react-router';
import { useState } from 'react';
import { Button } from '../../components/ui/button.tsx';
import { Modal } from '../../components/ui/modal.tsx';
import { defaultItemListSearch } from './itemListSearch.ts';

type Props = {
  itemId: string;
};

export function ItemDetailPage({ itemId }: Props) {
  const { clients } = useRouteContext({ from: '/_authenticated' });
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const query = useQuery(clients.items.get(itemId));
  const mutation = useMutation(clients.mutations.delete());
  const [confirmOpen, setConfirmOpen] = useState(false);

  if (query.isPending) {
    return (
      <main className="p-6">
        <p>読み込み中</p>
      </main>
    );
  }

  if (!query.data) {
    return (
      <main className="p-6">
        <p>サンプルリソースが見つかりません</p>
      </main>
    );
  }

  const item = query.data;

  async function confirmDelete(): Promise<void> {
    await mutation.mutateAsync(itemId);
    await queryClient.invalidateQueries({ queryKey: ['items'] });
    setConfirmOpen(false);
    await navigate({ to: '/items', search: defaultItemListSearch });
  }

  return (
    <main className="p-6">
      <p className="mb-4">
        <Link to="/items" search={defaultItemListSearch}>
          一覧へ戻る
        </Link>
      </p>
      <h1 className="text-xl font-semibold">{item.name}</h1>
      <p className="mt-2 whitespace-pre-wrap">{item.description}</p>
      <div className="mt-6 flex gap-3">
        <Link
          to="/items/$itemId/edit"
          params={{ itemId: item.id }}
          className="inline-flex items-center justify-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground"
        >
          編集
        </Link>
        <Button
          type="button"
          disabled={mutation.isPending}
          onClick={() => setConfirmOpen(true)}
        >
          削除
        </Button>
      </div>
      <Modal
        open={confirmOpen}
        title="削除の確認"
        onClose={() => setConfirmOpen(false)}
      >
        <p>「{item.name}」を削除しますか？この操作は取り消せません。</p>
        <div className="mt-4 flex justify-end gap-2">
          <Button type="button" onClick={() => setConfirmOpen(false)}>
            キャンセル
          </Button>
          <Button
            type="button"
            disabled={mutation.isPending}
            onClick={() => {
              void confirmDelete();
            }}
          >
            削除する
          </Button>
        </div>
      </Modal>
    </main>
  );
}
