import { useId, type InputHTMLAttributes } from 'react';
import { cn } from '../../lib/cn.ts';

type Props = InputHTMLAttributes<HTMLInputElement> & {
  invalid?: boolean;
  label?: string;
  describedBy?: string;
};

export function Input({
  className,
  invalid,
  id,
  label,
  describedBy,
  ...props
}: Props) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const input = (
    <input
      id={inputId}
      className={cn(
        'block w-full rounded-md border border-border px-3 py-2 text-sm',
        'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring',
        invalid && 'border-destructive',
        className,
      )}
      {...props}
      aria-invalid={invalid || undefined}
      aria-describedby={describedBy}
    />
  );
  if (!label) {
    return input;
  }
  return (
    <div className="flex flex-col gap-1 text-sm">
      <label htmlFor={inputId}>{label}</label>
      {input}
    </div>
  );
}
