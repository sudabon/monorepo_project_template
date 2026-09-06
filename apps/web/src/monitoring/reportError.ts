export type ErrorReporter = (
  error: unknown,
  context?: { componentStack?: string },
) => void;

let reporter: ErrorReporter = () => {
  // Empty by default. Wire Sentry / Datadog in docs/RUNBOOK.md.
};

export function setErrorReporter(next: ErrorReporter): void {
  reporter = next;
}

export function reportError(
  error: unknown,
  context?: { componentStack?: string },
): void {
  reporter(error, context);
}
