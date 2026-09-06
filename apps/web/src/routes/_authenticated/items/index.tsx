import { createFileRoute } from '@tanstack/react-router';
import { ItemListPage } from '../../../pages/items/ItemListPage.tsx';
import { itemListSearchSchema } from '../../../pages/items/itemListSearch.ts';

export const Route = createFileRoute('/_authenticated/items/')({
  validateSearch: itemListSearchSchema,
  component: ItemListRoute,
});

function ItemListRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  return (
    <ItemListPage
      search={search}
      onQueryChange={(q) => {
        void navigate({ search: { q, page: 1 }, replace: true });
      }}
      onPageChange={(page) => {
        void navigate({ search: (prev) => ({ ...prev, page }) });
      }}
    />
  );
}
