import { z } from 'zod';

const userSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
});

export const sessionSchema = z.discriminatedUnion('authenticated', [
  z.object({
    authenticated: z.literal(false),
  }),
  z.object({
    authenticated: z.literal(true),
    user: userSchema,
    csrfToken: z.string().min(1),
  }),
]);

export type Session = z.infer<typeof sessionSchema>;

export const sessionQueryKey = ['session'] as const;

export async function fetchSession(
  authBaseUrl: string,
  fetchImpl: typeof fetch = fetch,
): Promise<Session> {
  const response = await fetchImpl(`${authBaseUrl}/session`, {
    credentials: 'include',
  });
  if (!response.ok) {
    throw new Error('セッションを確認できませんでした');
  }
  return sessionSchema.parse(await response.json());
}
