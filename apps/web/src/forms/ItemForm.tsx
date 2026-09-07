import { ApiError } from '@monorepo-project-template/api-client';
import { zodResolver } from '@hookform/resolvers/zod';
import { useId } from 'react';
import { useForm } from 'react-hook-form';
import { Button } from '../components/ui/button.tsx';
import { Input } from '../components/ui/input.tsx';
import { Textarea } from '../components/ui/textarea.tsx';
import { itemInputSchema, type ItemInputValues } from './itemInputSchema.ts';
import { applyMappedErrors, mapServerErrors } from './mapServerErrors.ts';

const itemFields = new Set(['name', 'description']);

type Props = {
  submit: (values: ItemInputValues) => Promise<void>;
  defaultValues?: ItemInputValues;
};

export function ItemForm({
  submit,
  defaultValues = { name: '', description: '' },
}: Props) {
  const nameErrorId = useId();
  const descriptionErrorId = useId();
  const form = useForm<ItemInputValues>({
    resolver: zodResolver(itemInputSchema),
    defaultValues,
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
          describedBy={errors.name ? nameErrorId : undefined}
          {...register('name')}
        />
        {errors.name ? (
          <p id={nameErrorId} role="alert">
            {errors.name.message}
          </p>
        ) : null}
      </div>
      <div>
        <Textarea
          label="説明"
          invalid={Boolean(errors.description)}
          describedBy={errors.description ? descriptionErrorId : undefined}
          rows={4}
          {...register('description')}
        />
        {errors.description ? (
          <p id={descriptionErrorId} role="alert">
            {errors.description.message}
          </p>
        ) : null}
      </div>
      <Button type="submit" disabled={isSubmitting}>
        保存
      </Button>
    </form>
  );
}
