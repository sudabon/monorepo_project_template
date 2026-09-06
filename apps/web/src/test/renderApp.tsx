import { QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, RouterProvider } from '@tanstack/react-router';
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { createAppClients } from '../api/clients.ts';
import { sessionQueryKey, type Session } from '../auth/session.ts';
import { ErrorBoundary } from '../components/ErrorBoundary.tsx';
import { RootErrorFallback } from '../components/ErrorFallback.tsx';
import { AppErrorToasts } from '../components/ui/AppErrorToasts.tsx';
import { ToastProvider } from '../components/ui/toast.tsx';
import type { RuntimeConfig } from '../config/schema.ts';
import { createAppQueryClient } from '../query/createQueryClient.ts';
import { notifyAppError } from '../query/errorBus.ts';
import { handleClassifiedError } from '../query/handleError.ts';
import { createAppRouter } from '../router.tsx';

export const defaultConfig: RuntimeConfig = { apiBaseUrl: '/api' };

export const authedSession: Session = {
  authenticated: true,
  user: { id: 'user-1', name: 'Demo' },
  csrfToken: 'csrf-token',
};

export async function renderApp(options?: {
  path?: string;
  session?: Session;
  config?: RuntimeConfig;
  fetchImpl?: typeof fetch;
  waitForLoad?: boolean;
}): Promise<{
  queryClient: ReturnType<typeof createAppQueryClient>;
  router: ReturnType<typeof createAppRouter>;
}> {
  if (options?.fetchImpl) {
    globalThis.fetch = options.fetchImpl;
  }
  const config = options?.config ?? defaultConfig;
  const queryClient = createAppQueryClient({
    onError: (error) => {
      handleClassifiedError(error, {
        onAuthError: () => {
          queryClient.setQueryData(sessionQueryKey, { authenticated: false });
          void router.navigate({ to: '/login' });
        },
        onNotify: notifyAppError,
      });
    },
  });
  queryClient.setDefaultOptions({
    queries: { retry: false, staleTime: 30_000 },
  });
  if (options?.session) {
    queryClient.setQueryData(sessionQueryKey, options.session);
  }
  const clients = createAppClients(config, queryClient);
  const router = createAppRouter({
    queryClient,
    config,
    clients,
    history: createMemoryHistory({
      initialEntries: [options?.path ?? '/'],
    }),
  });
  if (options?.waitForLoad !== false) {
    await router.load();
  }
  render(
    (
      <ErrorBoundary fallback={<RootErrorFallback />}>
        <QueryClientProvider client={queryClient}>
          <ToastProvider>
            <AppErrorToasts />
            <RouterProvider router={router} />
          </ToastProvider>
        </QueryClientProvider>
      </ErrorBoundary>
    ) as ReactElement,
  );
  return { queryClient, router };
}
