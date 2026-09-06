import { z } from 'zod';

export const runtimeConfigSchema = z.object({
  apiBaseUrl: z.string().min(1),
});

export type RuntimeConfig = z.infer<typeof runtimeConfigSchema>;
