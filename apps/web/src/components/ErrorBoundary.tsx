import { Component, type ErrorInfo, type ReactNode } from 'react';
import { reportError } from '../monitoring/reportError.ts';

type Props = {
  children: ReactNode;
  fallback: ReactNode;
  onError?: (error: Error, info: ErrorInfo) => void;
};

type State = { error: Error | null };

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    reportError(error, { componentStack: info.componentStack ?? undefined });
    this.props.onError?.(error, info);
  }

  render(): ReactNode {
    if (this.state.error) {
      return this.props.fallback;
    }
    return this.props.children;
  }
}
