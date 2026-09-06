import { ApiError } from '@monorepo-project-template/api-client';
import { describe, expect, it } from 'vitest';
import { shouldRetryQuery } from './retry.ts';

function apiError(status: number): ApiError {
  return new ApiError(new Response(null, { status }), {
    code: 'error',
    message: 'failed',
  });
}

describe('shouldRetryQuery', () => {
  it('does not retry authentication errors', () => {
    expect(shouldRetryQuery(0, apiError(401))).toBe(false);
  });

  it('does not retry other client errors', () => {
    expect(shouldRetryQuery(0, apiError(400))).toBe(false);
    expect(shouldRetryQuery(0, apiError(404))).toBe(false);
    expect(shouldRetryQuery(0, apiError(422))).toBe(false);
  });

  it('retries server errors until the limit', () => {
    expect(shouldRetryQuery(0, apiError(500))).toBe(true);
    expect(shouldRetryQuery(1, apiError(500))).toBe(true);
    expect(shouldRetryQuery(2, apiError(500))).toBe(false);
  });
});
