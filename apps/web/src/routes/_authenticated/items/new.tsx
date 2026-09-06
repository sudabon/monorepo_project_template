import { createFileRoute } from '@tanstack/react-router';
import { ItemCreatePage } from '../../../pages/items/ItemCreatePage.tsx';

export const Route = createFileRoute('/_authenticated/items/new')({
  component: ItemCreatePage,
});
