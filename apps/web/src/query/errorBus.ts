type ErrorListener = (message: string) => void;

const listeners = new Set<ErrorListener>();

export function subscribeAppError(listener: ErrorListener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function notifyAppError(message: string): void {
  for (const listener of listeners) {
    listener(message);
  }
}
