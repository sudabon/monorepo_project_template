import { zodResolver } from '@hookform/resolvers/zod';
import { useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import { sessionQueryKey, sessionSchema } from '../auth/session.ts';
import { Button } from '../components/ui/button.tsx';
import { Input } from '../components/ui/input.tsx';
import { loginSchema, type LoginValues } from '../forms/loginSchema.ts';

export function LoginPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: '', password: '' },
  });
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = form;

  return (
    <main className="mx-auto max-w-sm p-6">
      <h1 className="text-xl font-semibold">ログイン</h1>
      <form
        className="mt-6 flex flex-col gap-3"
        onSubmit={handleSubmit(async (values) => {
          const response = await fetch('/auth/login', {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(values),
          });
          if (!response.ok) {
            setError('root.server', {
              type: 'server',
              message: 'ログインに失敗しました',
            });
            return;
          }
          const session = sessionSchema.parse(await response.json());
          queryClient.setQueryData(sessionQueryKey, session);
          await navigate({ to: '/' });
        })}
        noValidate
      >
        {errors.root?.server ? (
          <p role="alert">{errors.root.server.message}</p>
        ) : null}
        <div>
          <Input
            label="ユーザ名"
            autoComplete="username"
            invalid={Boolean(errors.username)}
            {...register('username')}
          />
          {errors.username ? (
            <p role="alert">{errors.username.message}</p>
          ) : null}
        </div>
        <div>
          <Input
            label="パスワード"
            type="password"
            autoComplete="current-password"
            invalid={Boolean(errors.password)}
            {...register('password')}
          />
          {errors.password ? (
            <p role="alert">{errors.password.message}</p>
          ) : null}
        </div>
        <Button type="submit" disabled={isSubmitting}>
          ログイン
        </Button>
      </form>
    </main>
  );
}
