import { describe, expect, it } from 'vitest';
import { runtimeConfigSchema } from './schema.ts';

describe('runtimeConfigSchema', () => {
  it('accepts a non-empty apiBaseUrl', () => {
    expect(runtimeConfigSchema.parse({ apiBaseUrl: '/api' })).toEqual({
      apiBaseUrl: '/api',
    });
  });

  it('rejects missing apiBaseUrl', () => {
    const result = runtimeConfigSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it('rejects an empty apiBaseUrl', () => {
    const result = runtimeConfigSchema.safeParse({ apiBaseUrl: '' });
    expect(result.success).toBe(false);
  });

  it('rejects a non-string apiBaseUrl', () => {
    const result = runtimeConfigSchema.safeParse({ apiBaseUrl: 1 });
    expect(result.success).toBe(false);
  });
});
