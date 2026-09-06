import { useRouteContext } from '@tanstack/react-router';

export function HomePage() {
  const { session } = useRouteContext({ from: '/_authenticated' });
  return (
    <main className="p-6">
      <h1 className="text-xl font-semibold">ホーム</h1>
      {session.authenticated ? (
        <p className="mt-2 text-muted-foreground">{session.user.name}</p>
      ) : null}
    </main>
  );
}
