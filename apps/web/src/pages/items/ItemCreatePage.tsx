import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useRouteContext } from '@tanstack/react-router';
import { ItemForm } from '../../forms/ItemForm.tsx';
import type { ItemInputValues } from '../../forms/itemInputSchema.ts';
import { defaultItemListSearch } from './itemListSearch.ts';

export function ItemCreatePage() {
  const { clients } = useRouteContext({ from: '/_authenticated' });
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const mutation = useMutation({
    ...clients.mutations.create(),
    meta: { formHandlesValidation: true },
  });

  async function submit(values: ItemInputValues): Promise<void> {
    await mutation.mutateAsync(values);
    await queryClient.invalidateQueries({ queryKey: ['items'] });
    await navigate({ to: '/items', search: defaultItemListSearch });
  }

  return (
    <main className="p-6">
      <p className="mb-4">
        <Link to="/items" search={defaultItemListSearch}>
          一覧へ戻る
        </Link>
      </p>
      <h1 className="text-xl font-semibold">サンプルリソースを作成</h1>
      <div className="mt-4">
        <ItemForm submit={submit} />
      </div>
    </main>
  );
}
