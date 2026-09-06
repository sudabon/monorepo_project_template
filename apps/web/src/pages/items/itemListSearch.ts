import { z } from 'zod';

export const ITEM_PAGE_SIZE = 20;

export const itemListSearchSchema = z.object({
  q: z.string().catch(''),
  page: z.coerce.number().int().min(1).catch(1),
});

export type ItemListSearch = z.infer<typeof itemListSearchSchema>;

export const defaultItemListSearch: ItemListSearch = { q: '', page: 1 };
