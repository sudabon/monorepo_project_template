import {
  createRootRouteWithContext,
  Link,
  Outlet,
} from '@tanstack/react-router';
import type { RouterContext } from '../app-context.ts';
import { RuntimeConfigProvider } from '../config/RuntimeConfigContext.tsx';
import { RootErrorFallback } from '../components/ErrorFallback.tsx';
import { ToastViewport } from '../components/ui/toast.tsx';

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
  errorComponent: RootErrorFallback,
});

function RootLayout() {
  const { config } = Route.useRouteContext();
  return (
    <RuntimeConfigProvider value={config}>
      <header className="border-b border-border px-4 py-3">
        <p className="font-semibold">
          <Link to="/" className="text-foreground">
            Template
          </Link>
        </p>
      </header>
      <Outlet />
      <ToastViewport />
    </RuntimeConfigProvider>
  );
}
