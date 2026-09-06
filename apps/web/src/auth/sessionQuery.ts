import { queryOptions } from '@tanstack/react-query';
import { fetchSession, sessionQueryKey } from './session.ts';

export function sessionQueryOptions() {
  return queryOptions({
    queryKey: sessionQueryKey,
    queryFn: () => fetchSession(),
  });
}
