import { describe, expect, it, vi } from 'vitest';
import { ConfigLoadError, loadRuntimeConfig } from './load.ts';

describe('loadRuntimeConfig', () => {
  it('returns parsed config from /config.json', async () => {
    const fetchImpl = vi.fn(async () =>
      Response.json({ apiBaseUrl: 'https://bff.example/api' }),
    );
    await expect(loadRuntimeConfig(fetchImpl)).resolves.toEqual({
      apiBaseUrl: 'https://bff.example/api',
    });
    expect(fetchImpl).toHaveBeenCalledWith('/config.json', {
      cache: 'no-store',
    });
  });

  it('fails when the file cannot be fetched', async () => {
    const fetchImpl = vi.fn(async () => new Response('', { status: 404 }));
    await expect(loadRuntimeConfig(fetchImpl)).rejects.toBeInstanceOf(
      ConfigLoadError,
    );
  });

  it('fails when JSON is invalid', async () => {
    const fetchImpl = vi.fn(
      async () => new Response('not-json', { status: 200 }),
    );
    await expect(loadRuntimeConfig(fetchImpl)).rejects.toMatchObject({
      message: '設定ファイルの形式が不正です',
    });
  });

  it('fails when the schema does not match', async () => {
    const fetchImpl = vi.fn(async () => Response.json({ apiBaseUrl: '' }));
    await expect(loadRuntimeConfig(fetchImpl)).rejects.toMatchObject({
      message: '設定ファイルの形式が不正です',
    });
  });
});
