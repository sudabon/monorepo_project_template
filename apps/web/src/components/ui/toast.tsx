import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

export type ToastTone = 'error' | 'info';

export type Toast = {
  id: string;
  message: string;
  tone: ToastTone;
};

type ToastContextValue = {
  toasts: Toast[];
  pushToast: (toast: { message: string; tone?: ToastTone }) => void;
  dismissToast: (id: string) => void;
};

const ToastContext = createContext<ToastContextValue | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const dismissToast = useCallback((id: string) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const pushToast = useCallback(
    (toast: { message: string; tone?: ToastTone }) => {
      const id = crypto.randomUUID();
      setToasts((current) => [
        ...current,
        { id, message: toast.message, tone: toast.tone ?? 'info' },
      ]);
    },
    [],
  );

  const value = useMemo(
    () => ({ toasts, pushToast, dismissToast }),
    [toasts, pushToast, dismissToast],
  );

  return (
    <ToastContext.Provider value={value}>{children}</ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const value = useContext(ToastContext);
  if (!value) {
    throw new Error('ToastProvider is missing');
  }
  return value;
}

export function ToastViewport() {
  const { toasts, dismissToast } = useToast();
  return (
    <div
      className="fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-2"
      aria-live="polite"
      aria-relevant="additions"
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          role="status"
          className={
            toast.tone === 'error'
              ? 'rounded-md bg-destructive px-3 py-2 text-sm text-destructive-foreground'
              : 'rounded-md bg-foreground px-3 py-2 text-sm text-background'
          }
        >
          <p>{toast.message}</p>
          <button
            type="button"
            className="mt-1 text-xs underline"
            onClick={() => dismissToast(toast.id)}
          >
            閉じる
          </button>
        </div>
      ))}
    </div>
  );
}

// TODO(template): move focus to a toast only when the message is urgent and no dialog is open.
// TODO(template): announce toasts with aria-live="assertive" for auth/session-critical errors.
