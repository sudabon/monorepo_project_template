import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query';
import { classifyError } from './classifyError.ts';
import { shouldRetryQuery } from './retry.ts';

declare module '@tanstack/react-query' {
  interface Register {
    mutationMeta: {
      // Set by a mutation whose form maps contract field errors onto inputs.
      // Without it a 422 would be shown twice: on the field and in a toast.
      formHandlesValidation?: boolean;
    };
  }
}

export function createAppQueryClient(options?: {
  onError?: (error: unknown) => void;
}): QueryClient {
  const onError = options?.onError;
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (error) => {
        onError?.(error);
      },
    }),
    mutationCache: new MutationCache({
      onError: (error, _variables, _context, mutation) => {
        if (
          mutation.meta?.formHandlesValidation &&
          classifyError(error).kind === 'validation'
        ) {
          return;
        }
        onError?.(error);
      },
    }),
    defaultOptions: {
      queries: {
        // Session and list data stay usable across short navigations so
        // beforeLoad / screens do not refetch the same resource immediately.
        staleTime: 30_000,
        // Auth and other 4xx responses will not change on retry. Network and
        // 5xx failures get a couple of attempts before surfacing.
        retry: shouldRetryQuery,
        // B2B forms should not refetch on tab focus and overwrite in-progress work.
        refetchOnWindowFocus: false,
      },
    },
  });
}
