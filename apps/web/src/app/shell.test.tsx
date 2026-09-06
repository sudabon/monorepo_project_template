import { ApiError } from '@monorepo-project-template/api-client';
import { screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bootstrap } from '../bootstrap.tsx';
import { createAppClients } from '../api/clients.ts';
import { authedSession, renderApp } from '../test/renderApp.tsx';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('runtime config bootstrap', () => {
  beforeEach(() => {
    document.body.innerHTML = '<div id="root"></div>';
  });

  it('shows an error instead of a blank screen when config fetch fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes('/config.json')) {
          return new Response('', { status: 404 });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    );
    const root = document.getElementById('root');
    if (!root) {
      throw new Error('missing root');
    }
    await bootstrap(root);
    expect(
      await screen.findByRole('heading', {
        name: '設定を読み込めませんでした',
      }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'ログイン' }),
    ).not.toBeInTheDocument();
  });

  it('does not start the API client when schema validation fails', async () => {
    const fetchImpl = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/config.json')) {
        return jsonResponse({ apiBaseUrl: '' });
      }
      throw new Error(`client used before valid config: ${url}`);
    });
    vi.stubGlobal('fetch', fetchImpl);
    const root = document.getElementById('root');
    if (!root) {
      throw new Error('missing root');
    }
    await bootstrap(root);
    expect(
      await screen.findByRole('heading', {
        name: '設定を読み込めませんでした',
      }),
    ).toBeInTheDocument();
    expect(fetchImpl.mock.calls.map((call) => String(call[0]))).toEqual([
      '/config.json',
    ]);
  });
});

describe('config-driven API destination', () => {
  it('sends traffic to the apiBaseUrl from config', async () => {
    const urls: string[] = [];
    const config = { apiBaseUrl: 'https://staging.example/api' };
    const queryClient = (
      await import('../query/createQueryClient.ts')
    ).createAppQueryClient();
    const clients = createAppClients(config, queryClient);
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      urls.push(String(input instanceof Request ? input.url : input));
      return jsonResponse({ items: [], page: 1, pageSize: 20, total: 0 });
    });
    await queryClient.fetchQuery(clients.items.list());
    expect(urls[0]).toContain('https://staging.example/api/items');
    queryClient.clear();
  });
});

describe('auth routing', () => {
  it('opens the login screen for an unauthenticated visitor', async () => {
    await renderApp({
      path: '/',
      fetchImpl: vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes('/auth/session')) {
          return jsonResponse({ authenticated: false });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    expect(
      await screen.findByRole('heading', { name: 'ログイン' }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'ホーム' }),
    ).not.toBeInTheDocument();
  });

  it('never flashes protected content when opening a guarded URL unauthenticated', async () => {
    let resolveSession: ((value: Response) => void) | undefined;
    const sessionGate = new Promise<Response>((resolve) => {
      resolveSession = resolve;
    });
    await renderApp({
      path: '/',
      waitForLoad: false,
      fetchImpl: vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes('/auth/session')) {
          return sessionGate;
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    expect(
      screen.queryByRole('heading', { name: 'ホーム' }),
    ).not.toBeInTheDocument();
    resolveSession?.(jsonResponse({ authenticated: false }));
    expect(
      await screen.findByRole('heading', { name: 'ログイン' }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'ホーム' }),
    ).not.toBeInTheDocument();
  });

  it('shows protected content for an authenticated session', async () => {
    await renderApp({ path: '/', session: authedSession });
    expect(
      await screen.findByRole('heading', { name: 'ホーム' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Demo')).toBeInTheDocument();
  });

  it('reuses the session query cache across ensureQueryData calls', async () => {
    const fetchImpl = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/auth/session')) {
        return jsonResponse(authedSession);
      }
      throw new Error(`unexpected fetch ${url}`);
    });
    const { queryClient } = await renderApp({ path: '/', fetchImpl });
    await screen.findByRole('heading', { name: 'ホーム' });
    const { sessionQueryOptions } = await import('../auth/sessionQuery.ts');
    await queryClient.ensureQueryData(sessionQueryOptions());
    await queryClient.ensureQueryData(sessionQueryOptions());
    const sessionCalls = fetchImpl.mock.calls.filter((call) =>
      String(call[0]).includes('/auth/session'),
    );
    expect(sessionCalls).toHaveLength(1);
  });
});

describe('global query errors', () => {
  it('navigates to login on 401', async () => {
    const { queryClient } = await renderApp({
      path: '/',
      session: authedSession,
    });
    await screen.findByRole('heading', { name: 'ホーム' });
    await queryClient
      .fetchQuery({
        queryKey: ['protected-resource'],
        queryFn: async () => {
          throw new ApiError(new Response(null, { status: 401 }), {
            code: 'unauthenticated',
            message: 'Session expired.',
          });
        },
      })
      .catch(() => undefined);
    expect(
      await screen.findByRole('heading', { name: 'ログイン' }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'ホーム' }),
    ).not.toBeInTheDocument();
  });

  it('shows a toast for 5xx errors', async () => {
    const { queryClient } = await renderApp({
      path: '/',
      session: authedSession,
    });
    await screen.findByRole('heading', { name: 'ホーム' });
    await queryClient
      .fetchQuery({
        queryKey: ['server-fail'],
        queryFn: async () => {
          throw new ApiError(new Response(null, { status: 500 }), {
            code: 'internal_error',
            message: 'An internal error occurred.',
          });
        },
      })
      .catch(() => undefined);
    expect(await screen.findByRole('status')).toHaveTextContent(
      'An internal error occurred.',
    );
    expect(screen.getByRole('heading', { name: 'ホーム' })).toBeInTheDocument();
  });

  it('displays unclassified errors instead of swallowing them', async () => {
    const { queryClient } = await renderApp({
      path: '/',
      session: authedSession,
    });
    await screen.findByRole('heading', { name: 'ホーム' });
    await queryClient
      .fetchQuery({
        queryKey: ['unknown-fail'],
        queryFn: async () => {
          throw new Error('quota exceeded');
        },
      })
      .catch(() => undefined);
    expect(await screen.findByRole('status')).toHaveTextContent(
      'quota exceeded',
    );
  });
});

describe('request failures', () => {
  it('shows an error when session lookup fails instead of a blank screen', async () => {
    await renderApp({
      path: '/',
      fetchImpl: vi.fn(async () => {
        throw new Error('failed to fetch');
      }),
    });
    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(document.body).not.toHaveTextContent(/^[\s]*$/);
  });
});
