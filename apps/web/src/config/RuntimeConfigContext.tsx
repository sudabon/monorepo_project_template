import { createContext, useContext } from 'react';
import type { RuntimeConfig } from './schema.ts';

const RuntimeConfigContext = createContext<RuntimeConfig | null>(null);

export const RuntimeConfigProvider = RuntimeConfigContext.Provider;

export function useRuntimeConfig(): RuntimeConfig {
  const config = useContext(RuntimeConfigContext);
  if (!config) {
    throw new Error('RuntimeConfigProvider is missing');
  }
  return config;
}
