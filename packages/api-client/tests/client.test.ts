import assert from 'node:assert/strict';
import { test } from 'node:test';
import { MutationObserver, QueryClient } from '@tanstack/react-query';
import {
  ApiError,
  createItemMutations,
  createItemQueries,
} from '../src/index.ts';

test('list serializes pagination and returns the response body', async () => {
  const cache = new QueryClient();
  const items = createItemQueries({
    baseUrl: 'https://example.test/api',
    fetch: async (request) => {
      assert.equal(request.method, 'GET');
      assert.equal(
        request.url,
        'https://example.test/api/items?page=2&pageSize=10',
      );
      return Response.json({ items: [], page: 2, pageSize: 10, total: 11 });
    },
  });
  try {
    assert.deepEqual(
      await cache.fetchQuery(items.list({ page: 2, pageSize: 10 })),
      {
        items: [],
        page: 2,
        pageSize: 10,
        total: 11,
      },
    );
  } finally {
    cache.clear();
  }
});

test('HTTP errors reach the caller with status and shared error body', async () => {
  const cache = new QueryClient();
  const items = createItemQueries({
    baseUrl: 'https://example.test/api',
    fetch: async () =>
      Response.json(
        { code: 'unauthorized', message: 'Session expired.' },
        { status: 401 },
      ),
  });
  try {
    await assert.rejects(
      cache.fetchQuery(items.get('missing')),
      (error: unknown) => {
        assert.ok(error instanceof ApiError);
        assert.equal(error.status, 401);
        assert.deepEqual(error.body, {
          code: 'unauthorized',
          message: 'Session expired.',
        });
        return true;
      },
    );
  } finally {
    cache.clear();
  }
});

test('query cancellation reaches the transport AbortSignal', async () => {
  const cache = new QueryClient();
  let transportSignal: AbortSignal | undefined;
  const items = createItemQueries({
    baseUrl: 'https://example.test/api',
    fetch: (request) => {
      transportSignal = request.signal;
      return new Promise((_resolve, reject) => {
        request.signal.addEventListener('abort', () =>
          reject(request.signal.reason),
        );
      });
    },
  });
  const options = items.list();
  const pending = cache.fetchQuery(options);
  const rejected = assert.rejects(pending);
  try {
    await cache.cancelQueries({ queryKey: options.queryKey });
    await rejected;
    assert.equal(transportSignal?.aborted, true);
  } finally {
    cache.clear();
  }
});

const savedItem = {
  id: '7d50d365-01b8-491e-bdda-71bd20774db9',
  name: 'Saved',
  description: '',
  createdAt: '2026-09-06T00:00:00Z',
  updatedAt: '2026-09-06T00:00:00Z',
};

test('create sends POST input only when the mutation executes and returns the Item', async () => {
  const cache = new QueryClient();
  let requests = 0;
  const mutations = createItemMutations({
    baseUrl: 'https://example.test/api',
    fetch: async (request) => {
      requests++;
      assert.equal(request.method, 'POST');
      assert.equal(request.url, 'https://example.test/api/items');
      assert.equal(request.headers.get('Content-Type'), 'application/json');
      assert.deepEqual(await request.json(), { name: 'Saved' });
      return Response.json(savedItem, { status: 201 });
    },
  });
  const observer = new MutationObserver(cache, mutations.create());
  try {
    assert.equal(requests, 0);
    assert.deepEqual(await observer.mutate({ name: 'Saved' }), savedItem);
    assert.equal(requests, 1);
  } finally {
    cache.clear();
  }
});

test('update sends PUT to the selected Item and returns its response', async () => {
  const cache = new QueryClient();
  const mutations = createItemMutations({
    baseUrl: 'https://example.test/api',
    fetch: async (request) => {
      assert.equal(request.method, 'PUT');
      assert.equal(
        request.url,
        'https://example.test/api/items/7d50d365-01b8-491e-bdda-71bd20774db9',
      );
      assert.deepEqual(await request.json(), { name: 'Saved' });
      return Response.json(savedItem);
    },
  });
  try {
    const observer = new MutationObserver(cache, mutations.update());
    assert.deepEqual(
      await observer.mutate({ id: savedItem.id, body: { name: 'Saved' } }),
      savedItem,
    );
  } finally {
    cache.clear();
  }
});

test('delete sends DELETE without a body and accepts an empty 204 response', async () => {
  const cache = new QueryClient();
  const mutations = createItemMutations({
    baseUrl: 'https://example.test/api',
    fetch: async (request) => {
      assert.equal(request.method, 'DELETE');
      assert.equal(
        request.url,
        'https://example.test/api/items/7d50d365-01b8-491e-bdda-71bd20774db9',
      );
      assert.equal(await request.text(), '');
      return new Response(null, { status: 204 });
    },
  });
  try {
    const observer = new MutationObserver(cache, mutations.delete());
    assert.equal(await observer.mutate(savedItem.id), undefined);
  } finally {
    cache.clear();
  }
});

test('mutation validation errors preserve every field error for the application', async () => {
  const cache = new QueryClient();
  const body = {
    code: 'validation_error',
    message: 'Some fields are invalid.',
    errors: [
      { field: 'name', message: 'Name must not be empty.' },
      { field: 'description', message: 'Description is too long.' },
    ],
  };
  const mutations = createItemMutations({
    baseUrl: 'https://example.test/api',
    fetch: async () => Response.json(body, { status: 422 }),
  });
  try {
    const observer = new MutationObserver(cache, mutations.create());
    await assert.rejects(observer.mutate({ name: '' }), (error: unknown) => {
      assert.ok(error instanceof ApiError);
      assert.equal(error.status, 422);
      assert.deepEqual(error.body, body);
      return true;
    });
  } finally {
    cache.clear();
  }
});
