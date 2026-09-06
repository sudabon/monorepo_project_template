import { ApiError } from '@monorepo-project-template/api-client';
import type { QueryClient } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';
import { createAppQueryClient } from './createQueryClient.ts';

function apiError(status: number): ApiError {
  return new ApiError(new Response(null, { status }), {
    code: 'error',
    message: 'failed',
  });
}

async function countAttempts(
  client: QueryClient,
  error: unknown,
): Promise<number> {
  let attempts = 0;
  try {
    await client.fetchQuery({
      queryKey: ['retry-count', statusOf(error), crypto.randomUUID()],
      queryFn: async () => {
        attempts += 1;
        throw error;
      },
    });
  } catch {
    // expected
  }
  return attempts;
}

function statusOf(error: unknown): number {
  return error instanceof ApiError ? error.status : 0;
}

describe('createAppQueryClient retry', () => {
  it('does not retry 401 or 400', async () => {
    const client = createAppQueryClient();
    expect(await countAttempts(client, apiError(401))).toBe(1);
    expect(await countAttempts(client, apiError(400))).toBe(1);
    client.clear();
  });

  it('retries 5xx', async () => {
    const client = createAppQueryClient();
    expect(await countAttempts(client, apiError(503))).toBe(3);
    client.clear();
  });

  it('invokes onError for query failures', async () => {
    const onError = vi.fn();
    const client = createAppQueryClient({ onError });
    client.setDefaultOptions({ queries: { retry: false } });
    await expect(
      client.fetchQuery({
        queryKey: ['on-error'],
        queryFn: async () => {
          throw apiError(500);
        },
      }),
    ).rejects.toBeInstanceOf(ApiError);
    expect(onError).toHaveBeenCalledOnce();
    client.clear();
  });
});
