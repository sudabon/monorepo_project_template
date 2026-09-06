import { QueryClient } from '@tanstack/react-query';
import createClient from 'openapi-fetch';
import type { components, paths } from '../src/generated/schema.d.ts';
import { createItemQueries } from '@monorepo-project-template/api-client';

const client = createClient<paths>({ baseUrl: 'https://example.test/api' });
client.GET('/items');
client.POST('/items', { body: { name: 'Example' } });
client.PUT('/items/{id}', {
  params: { path: { id: '7d50d365-01b8-491e-bdda-71bd20774db9' } },
  body: { name: 'Updated' },
});
// @ts-expect-error Only contract paths are accepted.
client.GET('/unknown');
// @ts-expect-error The collection has no DELETE operation.
client.DELETE('/items');
// @ts-expect-error A name is required by the generated request body.
client.POST('/items', { body: {} });

const items = createItemQueries({ baseUrl: 'https://example.test/api' });
const cache = new QueryClient();
const page: Promise<components['schemas']['ItemPage']> = cache.fetchQuery(
  items.list({ page: 2, pageSize: 10 }),
);
const item: Promise<components['schemas']['Item']> = cache.fetchQuery(
  items.get('7d50d365-01b8-491e-bdda-71bd20774db9'),
);
// @ts-expect-error Pagination comes from generated query parameters.
items.list({ page: '2' });
// @ts-expect-error Identifiers are strings in the contract.
items.get(123);
void page;
void item;
