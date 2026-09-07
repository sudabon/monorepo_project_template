import { useId, type TextareaHTMLAttributes } from 'react';
import { cn } from '../../lib/cn.ts';

type Props = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  invalid?: boolean;
  label?: string;
  describedBy?: string;
};

export function Textarea({
  className,
  invalid,
  id,
  label,
  describedBy,
  ...props
}: Props) {
  const generatedId = useId();
  const textareaId = id ?? generatedId;
  const textarea = (
    <textarea
      id={textareaId}
      className={cn(
        'block min-h-24 w-full rounded-md border border-border px-3 py-2 text-sm',
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
    return textarea;
  }
  return (
    <div className="flex flex-col gap-1 text-sm">
      <label htmlFor={textareaId}>{label}</label>
      {textarea}
    </div>
  );
}
