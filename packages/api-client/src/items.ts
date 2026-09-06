import { queryOptions } from '@tanstack/react-query';
import type { ClientOptions } from 'openapi-fetch';
import { createApiClient, responseData } from './client.ts';
import type { operations } from './generated/schema.d.ts';

type ListItemsParams = operations['listItems']['parameters']['query'];
type ItemId = operations['getItem']['parameters']['path']['id'];

export function createItemQueries(options: ClientOptions) {
  const client = createApiClient(options);

  return {
    list(params: ListItemsParams = {}) {
      return queryOptions({
        queryKey: ['items', 'list', params] as const,
        queryFn: async ({ signal }) =>
          responseData(
            await client.GET('/items', { params: { query: params }, signal }),
          ),
      });
    },
    get(id: ItemId) {
      return queryOptions({
        queryKey: ['items', 'detail', id] as const,
        queryFn: async ({ signal }) =>
          responseData(
            await client.GET('/items/{id}', {
              params: { path: { id } },
              signal,
            }),
          ),
      });
    },
  };
}
