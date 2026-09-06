import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { ClientOptions } from 'openapi-fetch';
import { createApiClient, responseData } from './client.ts';
import type { operations } from './generated/schema.d.ts';

type ListItemsParams = operations['listItems']['parameters']['query'];
type ItemId = operations['getItem']['parameters']['path']['id'];
type CreateItemBody =
  operations['createItem']['requestBody']['content']['application/json'];
type UpdateItemVariables = operations['updateItem']['parameters']['path'] & {
  body: operations['updateItem']['requestBody']['content']['application/json'];
};

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

export function createItemMutations(options: ClientOptions) {
  const client = createApiClient(options);

  return {
    create() {
      return mutationOptions({
        mutationKey: ['items', 'create'] as const,
        mutationFn: async (body: CreateItemBody) =>
          responseData(await client.POST('/items', { body })),
      });
    },
    update() {
      return mutationOptions({
        mutationKey: ['items', 'update'] as const,
        mutationFn: async ({ id, body }: UpdateItemVariables) =>
          responseData(
            await client.PUT('/items/{id}', { params: { path: { id } }, body }),
          ),
      });
    },
    delete() {
      return mutationOptions({
        mutationKey: ['items', 'delete'] as const,
        mutationFn: async (
          id: operations['deleteItem']['parameters']['path']['id'],
        ): Promise<void> => {
          responseData(
            await client.DELETE('/items/{id}', { params: { path: { id } } }),
          );
        },
      });
    },
  };
}
