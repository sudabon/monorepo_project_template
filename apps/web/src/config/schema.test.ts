import { describe, expect, it } from 'vitest';
import { runtimeConfigSchema } from './schema.ts';

describe('runtimeConfigSchema', () => {
  it('accepts non-empty base URLs', () => {
    expect(
      runtimeConfigSchema.parse({ apiBaseUrl: '/api', authBaseUrl: '/auth' }),
    ).toEqual({ apiBaseUrl: '/api', authBaseUrl: '/auth' });
  });

  it('rejects a config that sets only apiBaseUrl', () => {
    // Otherwise moving apiBaseUrl to another origin would leave sign-in
    // pointing at the SPA's own origin.
    const result = runtimeConfigSchema.safeParse({
      apiBaseUrl: 'https://bff.example/api',
    });
    expect(result.success).toBe(false);
  });

  it('rejects missing apiBaseUrl', () => {
    const result = runtimeConfigSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it('rejects an empty apiBaseUrl', () => {
    const result = runtimeConfigSchema.safeParse({
      apiBaseUrl: '',
      authBaseUrl: '/auth',
    });
    expect(result.success).toBe(false);
  });

  it('rejects a non-string apiBaseUrl', () => {
    const result = runtimeConfigSchema.safeParse({
      apiBaseUrl: 1,
      authBaseUrl: '/auth',
    });
    expect(result.success).toBe(false);
  });
});
