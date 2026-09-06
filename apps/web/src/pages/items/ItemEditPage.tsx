import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useRouteContext } from '@tanstack/react-router';
import { ItemForm } from '../../forms/ItemForm.tsx';
import { itemLoadMessage } from './ItemDetailPage.tsx';
import type { ItemInputValues } from '../../forms/itemInputSchema.ts';

type Props = {
  itemId: string;
};

export function ItemEditPage({ itemId }: Props) {
  const { clients } = useRouteContext({ from: '/_authenticated' });
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const query = useQuery(clients.items.get(itemId));
  const mutation = useMutation({
    ...clients.mutations.update(),
    meta: { formHandlesValidation: true },
  });

  async function submit(values: ItemInputValues): Promise<void> {
    await mutation.mutateAsync({ id: itemId, body: values });
    await queryClient.invalidateQueries({ queryKey: ['items'] });
    await navigate({ to: '/items/$itemId', params: { itemId } });
  }

  if (query.isPending) {
    return (
      <main className="p-6">
        <p>読み込み中</p>
      </main>
    );
  }

  if (query.isError || !query.data) {
    return (
      <main className="p-6">
        <p role="alert">{itemLoadMessage(query.error)}</p>
      </main>
    );
  }

  return (
    <main className="p-6">
      <p className="mb-4">
        <Link to="/items/$itemId" params={{ itemId }}>
          詳細へ戻る
        </Link>
      </p>
      <h1 className="text-xl font-semibold">サンプルリソースを編集</h1>
      <div className="mt-4">
        <ItemForm
          key={query.data.id}
          defaultValues={{
            name: query.data.name,
            description: query.data.description,
          }}
          submit={submit}
        />
      </div>
    </main>
  );
}
