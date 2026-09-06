import { createFileRoute } from '@tanstack/react-router';
import { ItemEditPage } from '../../../../pages/items/ItemEditPage.tsx';

export const Route = createFileRoute('/_authenticated/items/$itemId/edit')({
  component: ItemEditRoute,
});

function ItemEditRoute() {
  const { itemId } = Route.useParams();
  return <ItemEditPage itemId={itemId} />;
}
