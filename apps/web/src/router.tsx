import {
  createBrowserHistory,
  createRouter,
  type RouterHistory,
} from '@tanstack/react-router';
import type { QueryClient } from '@tanstack/react-query';
import type { AppClients } from './api/clients.ts';
import type { RuntimeConfig } from './config/schema.ts';
import { RouteErrorFallback } from './components/ErrorFallback.tsx';
import { routeTree } from './routeTree.gen.ts';

export function createAppRouter(options: {
  queryClient: QueryClient;
  config: RuntimeConfig;
  clients: AppClients;
  history?: RouterHistory;
}) {
  return createRouter({
    routeTree,
    history: options.history ?? createBrowserHistory(),
    context: {
      queryClient: options.queryClient,
      config: options.config,
      clients: options.clients,
    },
    defaultPreload: false,
    scrollRestoration: !options.history,
    defaultErrorComponent: RouteErrorFallback,
  });
}

export type AppRouter = ReturnType<typeof createAppRouter>;

declare module '@tanstack/react-router' {
  interface Register {
    router: AppRouter;
  }
}
