import assert from 'node:assert/strict';
import { test } from 'node:test';
import { QueryClient } from '@tanstack/react-query';
import { ApiError, createItemQueries } from '../src/index.ts';

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
    await assert.rejects(cache.fetchQuery(items.get('missing')), (error) => {
      assert.ok(error instanceof ApiError);
      assert.equal(error.status, 401);
      assert.deepEqual(error.body, {
        code: 'unauthorized',
        message: 'Session expired.',
      });
      return true;
    });
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
