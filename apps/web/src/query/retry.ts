import { ApiError } from '@monorepo-project-template/api-client';

export function shouldRetryQuery(
  failureCount: number,
  error: unknown,
): boolean {
  if (isClientOrAuthError(error)) {
    return false;
  }
  return failureCount < 2;
}

export function isClientOrAuthError(error: unknown): boolean {
  return error instanceof ApiError && error.status >= 400 && error.status < 500;
}
