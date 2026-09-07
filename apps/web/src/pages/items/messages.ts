import { ApiError } from '@monorepo-project-template/api-client';

export function itemLoadMessage(error: unknown): string {
  if (error instanceof ApiError && error.status === 404) {
    return 'サンプルリソースが見つかりません';
  }
  return 'サンプルリソースを取得できませんでした';
}
