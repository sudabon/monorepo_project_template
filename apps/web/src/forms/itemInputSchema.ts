import { constraints } from '@monorepo-project-template/api-client';
import { z } from 'zod';

const name = constraints.ItemInput.name;
const description = constraints.ItemInput.description;

export const itemInputSchema = z.object({
  name: z
    .string()
    .min(name.minLength, '名前を入力してください')
    .max(name.maxLength, `名前は ${name.maxLength} 文字以内で入力してください`)
    .refine((value) => new RegExp(name.pattern).test(value), {
      message: '名前に NUL 文字は使えません',
    }),
  description: z
    .string()
    .max(
      description.maxLength,
      `説明は ${description.maxLength} 文字以内で入力してください`,
    )
    .refine((value) => new RegExp(description.pattern).test(value), {
      message: '説明に NUL 文字は使えません',
    }),
});

export type ItemInputValues = z.infer<typeof itemInputSchema>;
