import { createFileRoute } from '@tanstack/react-router';
import { ItemDetailPage } from '../../../../pages/items/ItemDetailPage.tsx';

export const Route = createFileRoute('/_authenticated/items/$itemId/')({
  component: ItemDetailRoute,
});

function ItemDetailRoute() {
  const { itemId } = Route.useParams();
  return <ItemDetailPage itemId={itemId} />;
}
