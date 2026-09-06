import { render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ErrorBoundary } from './ErrorBoundary.tsx';
import { RouteErrorFallback } from './ErrorFallback.tsx';
import { setErrorReporter } from '../monitoring/reportError.ts';

function Boom(): never {
  throw new Error('page crashed');
}

describe('ErrorBoundary', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
    setErrorReporter(() => {});
  });

  it('keeps surrounding chrome usable when a screen throws', () => {
    render(
      <div>
        <nav>
          <a href="/login">ログインへ</a>
        </nav>
        <ErrorBoundary fallback={<RouteErrorFallback />}>
          <Boom />
        </ErrorBoundary>
      </div>,
    );
    expect(
      screen.getByRole('link', { name: 'ログインへ' }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('heading', {
        name: 'この画面の表示中にエラーが発生しました',
      }),
    ).toBeInTheDocument();
  });

  it('calls the monitoring hook when the top-level boundary catches', () => {
    const reporter = vi.fn();
    setErrorReporter(reporter);
    render(
      <ErrorBoundary fallback={<p>アプリケーションでエラーが発生しました</p>}>
        <Boom />
      </ErrorBoundary>,
    );
    expect(reporter).toHaveBeenCalled();
    expect(reporter.mock.calls[0]?.[0]).toBeInstanceOf(Error);
  });
});
