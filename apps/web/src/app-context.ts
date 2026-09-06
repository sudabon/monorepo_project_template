import type { QueryClient } from '@tanstack/react-query';
import type { AppClients } from './api/clients.ts';
import type { RuntimeConfig } from './config/schema.ts';

export type RouterContext = {
  queryClient: QueryClient;
  config: RuntimeConfig;
  clients: AppClients;
};
