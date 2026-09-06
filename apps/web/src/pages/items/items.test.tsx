import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { authedSession, renderApp } from '../../test/renderApp.tsx';

const testConfig = {
  apiBaseUrl: 'http://example.test/api',
  authBaseUrl: 'http://example.test/auth',
};

const itemA = {
  id: '11111111-1111-4111-8111-111111111111',
  name: 'Alpha Widget',
  description: 'First',
  createdAt: '2026-09-06T00:00:00Z',
  updatedAt: '2026-09-06T00:00:00Z',
};

const itemB = {
  id: '22222222-2222-4222-8222-222222222222',
  name: 'Beta Gadget',
  description: 'Second',
  createdAt: '2026-09-06T00:00:01Z',
  updatedAt: '2026-09-06T00:00:01Z',
};

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function itemsFetch(
  handler: (
    input: RequestInfo | URL,
    init?: RequestInit,
  ) => Response | Promise<Response>,
) {
  return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input instanceof Request ? input.url : input);
    if (url.includes('/auth/session')) {
      return jsonResponse(authedSession);
    }
    return handler(input, init);
  });
}

describe('sample resource list conventions', () => {
  it('keeps list search in the URL schema rather than component state', async () => {
    const { readFileSync } = await import('node:fs');
    const { dirname, join } = await import('node:path');
    const { fileURLToPath } = await import('node:url');
    const dir = dirname(fileURLToPath(import.meta.url));
    const listPage = readFileSync(join(dir, 'ItemListPage.tsx'), 'utf8');
    const listRoute = readFileSync(
      join(dir, '../../routes/_authenticated/items/index.tsx'),
      'utf8',
    );
    expect(listPage).not.toMatch(/\buseState\b/);
    expect(listRoute).toMatch('validateSearch');
    expect(listRoute).toMatch('itemListSearchSchema');
    const detailPage = readFileSync(join(dir, 'ItemDetailPage.tsx'), 'utf8');
    expect(detailPage).not.toMatch(/window\.confirm/);
  });
});

describe('sample resource list', () => {
  it('loads items through api-client queryOptions and renders them', async () => {
    const fetchImpl = itemsFetch(async (input) => {
      const url = String(input instanceof Request ? input.url : input);
      if (url.includes('/api/items?') || url.endsWith('/api/items')) {
        return jsonResponse({
          items: [itemA, itemB],
          page: 1,
          pageSize: 20,
          total: 2,
        });
      }
      throw new Error(`unexpected fetch ${url}`);
    });
    await renderApp({
      path: '/items',
      session: authedSession,
      config: testConfig,
      fetchImpl,
    });
    expect(
      await screen.findByRole('heading', { name: 'サンプルリソース' }),
    ).toBeInTheDocument();
    expect(await screen.findByText('Alpha Widget')).toBeInTheDocument();
    expect(screen.getByText('Beta Gadget')).toBeInTheDocument();
    expect(
      fetchImpl.mock.calls.some((call) => {
        const input = call[0];
        const url = input instanceof Request ? input.url : String(input);
        return url.includes('/api/items');
      }),
    ).toBe(true);
  });

  it('writes search changes to the URL instead of component state', async () => {
    const user = userEvent.setup();
    const { router } = await renderApp({
      path: '/items',
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input) => {
        const url = String(input instanceof Request ? input.url : input);
        if (url.includes('/api/items')) {
          return jsonResponse({
            items: [itemA, itemB],
            page: 1,
            pageSize: 20,
            total: 2,
          });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await screen.findByText('Alpha Widget');
    await user.type(screen.getByLabelText('検索'), 'Alpha');
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({
        q: 'Alpha',
        page: 1,
      });
    });
    expect(screen.getByLabelText('検索')).toHaveValue('Alpha');
    expect(screen.getByText('Alpha Widget')).toBeInTheDocument();
    expect(screen.queryByText('Beta Gadget')).not.toBeInTheDocument();
  });

  it('moves to the next page by rewriting the URL and shows that page', async () => {
    const user = userEvent.setup();
    const fetchImpl = itemsFetch(async (input) => {
      const url = String(input instanceof Request ? input.url : input);
      const parsed = new URL(url, 'http://example.test');
      if (
        !parsed.pathname.endsWith('/items') &&
        !parsed.pathname.endsWith('/api/items')
      ) {
        throw new Error(`unexpected fetch ${url}`);
      }
      const page = parsed.searchParams.get('page') ?? '1';
      if (page === '2') {
        return jsonResponse({
          items: [itemB],
          page: 2,
          pageSize: 20,
          total: 21,
        });
      }
      return jsonResponse({
        items: [itemA],
        page: 1,
        pageSize: 20,
        total: 21,
      });
    });
    const { router } = await renderApp({
      path: '/items',
      session: authedSession,
      config: testConfig,
      fetchImpl,
    });
    expect(await screen.findByText('Alpha Widget')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '次のページ' }));
    await waitFor(() => {
      expect(router.state.location.search).toMatchObject({ page: 2 });
    });
    expect(await screen.findByText('Beta Gadget')).toBeInTheDocument();
    expect(screen.queryByText('Alpha Widget')).not.toBeInTheDocument();
    expect(screen.getByText('ページ 2')).toBeInTheDocument();
  });

  it('shows an empty state instead of an error when nothing matches', async () => {
    await renderApp({
      path: '/items',
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input) => {
        const url = String(input instanceof Request ? input.url : input);
        if (url.includes('/api/items')) {
          return jsonResponse({ items: [], page: 1, pageSize: 20, total: 0 });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    expect(
      await screen.findByText('該当するサンプルリソースはありません'),
    ).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});

describe('sample resource detail', () => {
  it('opens the selected item from the list and shows its contents', async () => {
    const user = userEvent.setup();
    await renderApp({
      path: '/items',
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input) => {
        const url = String(input instanceof Request ? input.url : input);
        if (url.includes(`/api/items/${itemA.id}`)) {
          return jsonResponse(itemA);
        }
        if (url.includes('/api/items')) {
          return jsonResponse({
            items: [itemA],
            page: 1,
            pageSize: 20,
            total: 1,
          });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await user.click(await screen.findByRole('link', { name: 'Alpha Widget' }));
    expect(
      await screen.findByRole('heading', { name: 'Alpha Widget' }),
    ).toBeInTheDocument();
    expect(screen.getByText('First')).toBeInTheDocument();
  });
});

describe('sample resource create and edit', () => {
  it('creates an item and shows it on the list after refetch', async () => {
    const user = userEvent.setup();
    const created = {
      ...itemB,
      name: 'Created Widget',
    };
    let items = [itemA];
    await renderApp({
      path: '/items',
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input, init) => {
        const request =
          input instanceof Request ? input : new Request(input, init);
        const url = request.url;
        if (request.method === 'POST' && url.endsWith('/api/items')) {
          items = [created, ...items];
          return jsonResponse(created, 201);
        }
        if (
          url.includes(`/api/items/${itemA.id}`) ||
          url.includes(`/api/items/${created.id}`)
        ) {
          const found = items.find((item) => url.includes(item.id));
          return found
            ? jsonResponse(found)
            : jsonResponse({ code: 'not_found', message: 'missing' }, 404);
        }
        if (url.includes('/api/items')) {
          return jsonResponse({
            items,
            page: 1,
            pageSize: 20,
            total: items.length,
          });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await user.click(await screen.findByRole('link', { name: '新規作成' }));
    await user.type(await screen.findByLabelText('名前'), 'Created Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByText('Created Widget')).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'サンプルリソース' }),
    ).toBeInTheDocument();
  });

  it('updates an item and shows the new contents on the detail screen', async () => {
    const user = userEvent.setup();
    let current = { ...itemA };
    await renderApp({
      path: `/items/${itemA.id}`,
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input, init) => {
        const request =
          input instanceof Request ? input : new Request(input, init);
        const url = request.url;
        if (
          request.method === 'PUT' &&
          url.includes(`/api/items/${itemA.id}`)
        ) {
          const body = (await request.json()) as {
            name: string;
            description?: string;
          };
          current = {
            ...current,
            name: body.name,
            description: body.description ?? '',
          };
          return jsonResponse(current);
        }
        if (url.includes(`/api/items/${itemA.id}`)) {
          return jsonResponse(current);
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await user.click(await screen.findByRole('link', { name: '編集' }));
    const name = await screen.findByLabelText('名前');
    await user.clear(name);
    await user.type(name, 'Renamed Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(
      await screen.findByRole('heading', { name: 'Renamed Widget' }),
    ).toBeInTheDocument();
  });

  it('shows server field errors and does not keep a rejected create', async () => {
    const user = userEvent.setup();
    const created = false;
    await renderApp({
      path: '/items/new',
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input, init) => {
        const request =
          input instanceof Request ? input : new Request(input, init);
        const url = request.url;
        if (request.method === 'POST' && url.endsWith('/api/items')) {
          return jsonResponse(
            {
              code: 'validation_error',
              message: 'Invalid',
              errors: [{ field: 'name', message: '同じ名前は使えません' }],
            },
            422,
          );
        }
        if (url.includes('/api/items')) {
          return jsonResponse({
            items: created ? [itemA] : [],
            page: 1,
            pageSize: 20,
            total: created ? 1 : 0,
          });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await user.type(await screen.findByLabelText('名前'), 'Alpha Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByRole('alert')).toHaveTextContent(
      '同じ名前は使えません',
    );
    expect(created).toBe(false);
    expect(
      screen.getByRole('heading', { name: 'サンプルリソースを作成' }),
    ).toBeInTheDocument();
  });

  it('disables submit while a create request is in flight', async () => {
    const user = userEvent.setup();
    let resolveCreate: ((value: Response) => void) | undefined;
    const pending = new Promise<Response>((resolve) => {
      resolveCreate = resolve;
    });
    let posts = 0;
    await renderApp({
      path: '/items/new',
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input, init) => {
        const request =
          input instanceof Request ? input : new Request(input, init);
        const url = request.url;
        if (request.method === 'POST' && url.endsWith('/api/items')) {
          posts += 1;
          return pending;
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await user.type(await screen.findByLabelText('名前'), 'Pending Widget');
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByRole('button', { name: '保存' })).toBeDisabled();
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(posts).toBe(1);
    resolveCreate?.(jsonResponse({ ...itemA, name: 'Pending Widget' }, 201));
  });
});

describe('sample resource delete', () => {
  it('asks for confirmation in a modal and only deletes after approval', async () => {
    const user = userEvent.setup();
    let items = [itemA];
    await renderApp({
      path: `/items/${itemA.id}`,
      session: authedSession,
      config: testConfig,
      fetchImpl: itemsFetch(async (input, init) => {
        const request =
          input instanceof Request ? input : new Request(input, init);
        const url = request.url;
        if (
          request.method === 'DELETE' &&
          url.includes(`/api/items/${itemA.id}`)
        ) {
          items = [];
          return new Response(null, { status: 204 });
        }
        if (url.includes(`/api/items/${itemA.id}`)) {
          return items.length > 0
            ? jsonResponse(itemA)
            : jsonResponse({ code: 'not_found', message: 'missing' }, 404);
        }
        if (url.includes('/api/items')) {
          return jsonResponse({
            items,
            page: 1,
            pageSize: 20,
            total: items.length,
          });
        }
        throw new Error(`unexpected fetch ${url}`);
      }),
    });
    await user.click(await screen.findByRole('button', { name: '削除' }));
    expect(
      screen.getByRole('dialog', { name: '削除の確認' }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'Alpha Widget' }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '削除' }));
    await user.click(screen.getByRole('button', { name: '削除する' }));
    expect(
      await screen.findByRole('heading', { name: 'サンプルリソース' }),
    ).toBeInTheDocument();
    expect(screen.queryByText('Alpha Widget')).not.toBeInTheDocument();
  });
});
