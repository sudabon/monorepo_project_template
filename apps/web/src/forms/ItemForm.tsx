import { ApiError } from '@monorepo-project-template/api-client';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { Button } from '../components/ui/button.tsx';
import { Input } from '../components/ui/input.tsx';
import { itemInputSchema, type ItemInputValues } from './itemInputSchema.ts';
import { applyMappedErrors, mapServerErrors } from './mapServerErrors.ts';

const itemFields = new Set(['name', 'description']);

type Props = {
  submit: (values: ItemInputValues) => Promise<void>;
};

export function ItemForm({ submit }: Props) {
  const form = useForm<ItemInputValues>({
    resolver: zodResolver(itemInputSchema),
    defaultValues: { name: '', description: '' },
  });
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = form;

  return (
    <form
      className="flex max-w-md flex-col gap-3"
      onSubmit={handleSubmit(async (values) => {
        try {
          await submit(values);
        } catch (error) {
          if (
            error instanceof ApiError &&
            error.body &&
            'errors' in error.body
          ) {
            applyMappedErrors(
              setError,
              mapServerErrors(error.body.errors, itemFields),
            );
            return;
          }
          throw error;
        }
      })}
      noValidate
    >
      {errors.root?.server ? (
        <p role="alert">{errors.root.server.message}</p>
      ) : null}
      <div>
        <Input
          label="名前"
          invalid={Boolean(errors.name)}
          {...register('name')}
        />
        {errors.name ? <p role="alert">{errors.name.message}</p> : null}
      </div>
      <div>
        <Input
          label="説明"
          invalid={Boolean(errors.description)}
          {...register('description')}
        />
        {errors.description ? (
          <p role="alert">{errors.description.message}</p>
        ) : null}
      </div>
      <Button type="submit" disabled={isSubmitting}>
        保存
      </Button>
    </form>
  );
}
