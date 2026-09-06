import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import { sessionQueryOptions } from '../auth/sessionQuery.ts';
import { RouteErrorFallback } from '../components/ErrorFallback.tsx';

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(
      sessionQueryOptions(context.config.authBaseUrl),
    );
    if (!session.authenticated) {
      throw redirect({ to: '/login' });
    }
    return { session };
  },
  component: AuthenticatedLayout,
  errorComponent: RouteErrorFallback,
});

function AuthenticatedLayout() {
  return <Outlet />;
}
