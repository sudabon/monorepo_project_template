import {
  createItemMutations,
  createItemQueries,
} from '@monorepo-project-template/api-client';
import type { QueryClient } from '@tanstack/react-query';
import type { RuntimeConfig } from '../config/schema.ts';
import type { Session } from '../auth/session.ts';
import { sessionQueryKey } from '../auth/session.ts';

export function createAuthedFetch(queryClient: QueryClient): typeof fetch {
  return (input, init) => {
    const headers = new Headers(
      init?.headers ?? (input instanceof Request ? input.headers : undefined),
    );
    const method = (
      init?.method ?? (input instanceof Request ? input.method : 'GET')
    ).toUpperCase();
    if (
      method === 'POST' ||
      method === 'PUT' ||
      method === 'PATCH' ||
      method === 'DELETE'
    ) {
      const session = queryClient.getQueryData<Session>(sessionQueryKey);
      if (session?.authenticated) {
        headers.set('X-CSRF-Token', session.csrfToken);
      }
    }
    return fetch(input, { ...init, headers, credentials: 'include' });
  };
}

export function createAppClients(
  config: RuntimeConfig,
  queryClient: QueryClient,
) {
  const fetchImpl = createAuthedFetch(queryClient);
  const options = {
    baseUrl: config.apiBaseUrl,
    credentials: 'include' as const,
    fetch: fetchImpl,
  };
  return {
    items: createItemQueries(options),
    mutations: createItemMutations(options),
  };
}

export type AppClients = ReturnType<typeof createAppClients>;
