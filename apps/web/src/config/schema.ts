import { z } from 'zod';

// Both bases are required. Defaulting authBaseUrl would let a deployment move
// apiBaseUrl to another origin and silently keep signing in against the SPA's.
export const runtimeConfigSchema = z.object({
  apiBaseUrl: z.string().min(1),
  authBaseUrl: z.string().min(1),
});

export type RuntimeConfig = z.infer<typeof runtimeConfigSchema>;
