import createClient from 'openapi-fetch';
import type { ClientOptions } from 'openapi-fetch';
import type { components, paths } from './generated/schema.d.ts';

type ErrorBody =
  | components['schemas']['Error']
  | components['schemas']['ValidationError'];

// Preserve transport information so the application's QueryClient can classify failures.
export class ApiError extends Error {
  readonly status: number;
  readonly body: ErrorBody | undefined;

  constructor(response: Response, body: ErrorBody | undefined) {
    super(body?.message ?? `HTTP ${response.status}`);
    this.name = 'ApiError';
    this.status = response.status;
    this.body = body;
  }
}

export function createApiClient(options: ClientOptions) {
  return createClient<paths>(options);
}

export function responseData<T>(result: {
  data?: T;
  error?: ErrorBody;
  response: Response;
}): T {
  if (!result.response.ok) {
    throw new ApiError(result.response, result.error);
  }
  // The generated success response is the source of truth for the payload type.
  return result.data as T;
}
