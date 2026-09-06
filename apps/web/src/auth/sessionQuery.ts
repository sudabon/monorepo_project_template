import { queryOptions } from '@tanstack/react-query';
import { fetchSession, sessionQueryKey } from './session.ts';

export function sessionQueryOptions(authBaseUrl: string) {
  return queryOptions({
    queryKey: sessionQueryKey,
    queryFn: () => fetchSession(authBaseUrl),
  });
}
