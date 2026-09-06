import type { FieldPath, FieldValues, UseFormSetError } from 'react-hook-form';

export type FieldErrorInput = {
  field: string;
  message: string;
};

export type MappedServerErrors = {
  fieldErrors: Record<string, string>;
  formErrors: string[];
};

export function mapServerErrors(
  errors: ReadonlyArray<FieldErrorInput>,
  fieldNames: ReadonlySet<string>,
): MappedServerErrors {
  const fieldErrors: Record<string, string> = {};
  const formErrors: string[] = [];
  for (const error of errors) {
    if (fieldNames.has(error.field)) {
      const existing = fieldErrors[error.field];
      fieldErrors[error.field] = existing
        ? `${existing} ${error.message}`
        : error.message;
    } else {
      formErrors.push(
        error.field ? `${error.field}: ${error.message}` : error.message,
      );
    }
  }
  return { fieldErrors, formErrors };
}

export function applyMappedErrors<TFieldValues extends FieldValues>(
  setError: UseFormSetError<TFieldValues>,
  mapped: MappedServerErrors,
): void {
  for (const [field, message] of Object.entries(mapped.fieldErrors)) {
    setError(field as FieldPath<TFieldValues>, { type: 'server', message });
  }
  if (mapped.formErrors.length > 0) {
    setError('root.server' as FieldPath<TFieldValues>, {
      type: 'server',
      message: mapped.formErrors.join('\n'),
    });
  }
}
