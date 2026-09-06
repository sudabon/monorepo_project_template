import { classifyError } from './classifyError.ts';

export function handleClassifiedError(
  error: unknown,
  deps: {
    onAuthError: () => void;
    onNotify: (message: string) => void;
  },
): void {
  const classified = classifyError(error);
  switch (classified.kind) {
    case 'auth':
      deps.onAuthError();
      return;
    case 'validation':
    case 'server':
    case 'unknown':
      deps.onNotify(classified.message);
  }
}
