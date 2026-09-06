import { describe, expect, it, vi } from 'vitest';
import { createAppQueryClient } from '../query/createQueryClient.ts';
import { authedSession } from '../test/renderApp.tsx';
import { sessionQueryKey } from '../auth/session.ts';
import { createAuthedFetch } from './clients.ts';

describe('createAuthedFetch', () => {
  it('adds the CSRF header when the transport passes a POST Request', async () => {
    const queryClient = createAppQueryClient();
    queryClient.setQueryData(sessionQueryKey, authedSession);
    let csrf = '';
    globalThis.fetch = vi.fn(async (_input, init) => {
      csrf = new Headers(init?.headers).get('X-CSRF-Token') ?? '';
      return new Response(null, { status: 204 });
    });
    const authedFetch = createAuthedFetch(queryClient);
    await authedFetch(
      new Request('http://example.test/api/items', { method: 'POST' }),
    );
    expect(csrf).toBe('csrf-token');
    queryClient.clear();
  });
});
