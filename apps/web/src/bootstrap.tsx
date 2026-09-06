import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { createAppClients } from './api/clients.ts';
import { sessionQueryKey } from './auth/session.ts';
import { ConfigErrorScreen } from './components/ConfigErrorScreen.tsx';
import { ErrorBoundary } from './components/ErrorBoundary.tsx';
import { RootErrorFallback } from './components/ErrorFallback.tsx';
import { AppErrorToasts } from './components/ui/AppErrorToasts.tsx';
import { ToastProvider } from './components/ui/toast.tsx';
import { loadRuntimeConfig } from './config/load.ts';
import { createAppQueryClient } from './query/createQueryClient.ts';
import { notifyAppError } from './query/errorBus.ts';
import { handleClassifiedError } from './query/handleError.ts';
import { createAppRouter } from './router.tsx';

export async function bootstrap(root: HTMLElement): Promise<void> {
  try {
    const config = await loadRuntimeConfig();
    const errorHandler: { current: (error: unknown) => void } = {
      current: () => {},
    };
    const queryClient = createAppQueryClient({
      onError: (error) => errorHandler.current(error),
    });
    const clients = createAppClients(config, queryClient);
    const router = createAppRouter({ queryClient, config, clients });
    errorHandler.current = (error) => {
      handleClassifiedError(error, {
        onAuthError: () => {
          queryClient.setQueryData(sessionQueryKey, { authenticated: false });
          void router.navigate({ to: '/login' });
        },
        onNotify: notifyAppError,
      });
    };
    createRoot(root).render(
      <StrictMode>
        <ErrorBoundary fallback={<RootErrorFallback />}>
          <QueryClientProvider client={queryClient}>
            <ToastProvider>
              <AppErrorToasts />
              <RouterProvider router={router} />
            </ToastProvider>
          </QueryClientProvider>
        </ErrorBoundary>
      </StrictMode>,
    );
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : '設定ファイルを読み込めませんでした';
    createRoot(root).render(<ConfigErrorScreen message={message} />);
  }
}
