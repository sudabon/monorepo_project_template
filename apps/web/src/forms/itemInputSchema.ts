import { z } from 'zod';

export const itemInputSchema = z.object({
  name: z
    .string()
    .min(1, '名前を入力してください')
    .refine((value) => !value.includes('\u0000'), {
      message: '名前に NUL 文字は使えません',
    }),
  description: z.string().refine((value) => !value.includes('\u0000'), {
    message: '説明に NUL 文字は使えません',
  }),
});

export type ItemInputValues = z.infer<typeof itemInputSchema>;
