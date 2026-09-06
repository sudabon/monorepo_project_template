import { ApiError } from '@monorepo-project-template/api-client';

export type ClassifiedError =
  | { kind: 'auth' }
  | { kind: 'validation'; message: string }
  | { kind: 'server'; message: string }
  | { kind: 'unknown'; message: string };

export function classifyError(error: unknown): ClassifiedError {
  if (error instanceof ApiError && error.status === 401) {
    return { kind: 'auth' };
  }
  // Still notified by default: a validation error only stays quiet when a form
  // declares that it renders the field errors itself.
  if (error instanceof ApiError && error.status === 422) {
    return { kind: 'validation', message: error.message };
  }
  if (error instanceof ApiError && error.status >= 500) {
    return { kind: 'server', message: error.message };
  }
  const message =
    error instanceof Error ? error.message : '予期しないエラーが発生しました';
  return { kind: 'unknown', message };
}
