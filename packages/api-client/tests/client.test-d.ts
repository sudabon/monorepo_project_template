import { MutationObserver, QueryClient } from '@tanstack/react-query';
import createClient from 'openapi-fetch';
import type { components, paths } from '../src/generated/schema.d.ts';
import {
  createItemMutations,
  createItemQueries,
} from '@monorepo-project-template/api-client';

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

const mutations = createItemMutations({ baseUrl: 'https://example.test/api' });
const create = new MutationObserver(cache, mutations.create());
const update = new MutationObserver(cache, mutations.update());
const remove = new MutationObserver(cache, mutations.delete());
const created: Promise<components['schemas']['Item']> = create.mutate({
  name: 'Example',
});
const updated: Promise<components['schemas']['Item']> = update.mutate({
  id: '7d50d365-01b8-491e-bdda-71bd20774db9',
  body: { name: 'Updated' },
});
const deleted: Promise<void> = remove.mutate(
  '7d50d365-01b8-491e-bdda-71bd20774db9',
);
// @ts-expect-error Create requires the name from the generated request body.
create.mutate({});
// @ts-expect-error Update identifiers are strings.
update.mutate({ id: 123, body: { name: 'Updated' } });
// @ts-expect-error Update requires the generated body shape.
update.mutate({ id: 'id', body: { name: 123 } });
// @ts-expect-error Delete identifiers are strings.
remove.mutate(123);
// @ts-expect-error Create returns one Item, not an untyped result or an array.
const wrongCreate: Promise<components['schemas']['Item'][]> = create.mutate({
  name: 'Example',
});
// @ts-expect-error Update returns one Item, not an untyped result or an array.
const wrongUpdate: Promise<components['schemas']['Item'][]> = update.mutate({
  id: 'id',
  body: { name: 'Updated' },
});
// @ts-expect-error A 204 delete has no Item response body.
const wrongDelete: Promise<components['schemas']['Item']> = remove.mutate('id');
void created;
void updated;
void deleted;
void wrongCreate;
void wrongUpdate;
void wrongDelete;
