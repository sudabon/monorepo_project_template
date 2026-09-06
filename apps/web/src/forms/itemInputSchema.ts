import { z } from 'zod';

// Mirrors ItemInput in api/openapi.yaml: name 1-100, description up to 2000,
// and no NUL in either. Keep this in step with the contract.
export const itemInputSchema = z.object({
  name: z
    .string()
    .min(1, '名前を入力してください')
    .max(100, '名前は 100 文字以内で入力してください')
    .refine((value) => !value.includes('\u0000'), {
      message: '名前に NUL 文字は使えません',
    }),
  description: z
    .string()
    .max(2000, '説明は 2000 文字以内で入力してください')
    .refine((value) => !value.includes('\u0000'), {
      message: '説明に NUL 文字は使えません',
    }),
});

export type ItemInputValues = z.infer<typeof itemInputSchema>;
