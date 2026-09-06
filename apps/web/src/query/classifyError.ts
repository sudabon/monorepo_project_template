import { ApiError } from '@monorepo-project-template/api-client';

export type ClassifiedError =
  | { kind: 'auth' }
  | { kind: 'server'; message: string }
  | { kind: 'unknown'; message: string };

export function classifyError(error: unknown): ClassifiedError {
  if (error instanceof ApiError && error.status === 401) {
    return { kind: 'auth' };
  }
  if (error instanceof ApiError && error.status >= 500) {
    return { kind: 'server', message: error.message };
  }
  const message =
    error instanceof Error ? error.message : '予期しないエラーが発生しました';
  return { kind: 'unknown', message };
}
