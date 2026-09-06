import { useEffect } from 'react';
import { subscribeAppError } from '../../query/errorBus.ts';
import { useToast } from './toast.tsx';

export function AppErrorToasts() {
  const { pushToast } = useToast();
  useEffect(() => {
    return subscribeAppError((message) => {
      pushToast({ message, tone: 'error' });
    });
  }, [pushToast]);
  return null;
}
