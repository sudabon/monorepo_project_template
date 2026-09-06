import { ApiError } from '@monorepo-project-template/api-client';
import { describe, expect, it, vi } from 'vitest';
import { classifyError } from './classifyError.ts';
import { handleClassifiedError } from './handleError.ts';

function apiError(status: number, message = 'failed'): ApiError {
  return new ApiError(new Response(null, { status }), {
    code: 'error',
    message,
  });
}

describe('classifyError', () => {
  it('classifies 401 as auth', () => {
    expect(classifyError(apiError(401))).toEqual({ kind: 'auth' });
  });

  it('classifies 5xx as server', () => {
    expect(classifyError(apiError(500, 'boom'))).toEqual({
      kind: 'server',
      message: 'boom',
    });
  });

  it('does not swallow unknown errors', () => {
    expect(classifyError(new Error('network down'))).toEqual({
      kind: 'unknown',
      message: 'network down',
    });
  });
});

describe('handleClassifiedError', () => {
  it('routes 401 to onAuthError', () => {
    const onAuthError = vi.fn();
    const onNotify = vi.fn();
    handleClassifiedError(apiError(401), { onAuthError, onNotify });
    expect(onAuthError).toHaveBeenCalledOnce();
    expect(onNotify).not.toHaveBeenCalled();
  });

  it('notifies 5xx and unknown errors', () => {
    const onAuthError = vi.fn();
    const onNotify = vi.fn();
    handleClassifiedError(apiError(500, 'server'), { onAuthError, onNotify });
    handleClassifiedError(new Error('other'), { onAuthError, onNotify });
    expect(onAuthError).not.toHaveBeenCalled();
    expect(onNotify).toHaveBeenCalledWith('server');
    expect(onNotify).toHaveBeenCalledWith('other');
  });
});
