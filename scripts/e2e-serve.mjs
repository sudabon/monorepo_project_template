import { spawn } from 'node:child_process';

const children = [];
let shuttingDown = false;

function start(command, args, extraEnv = {}) {
  const child = spawn(command, args, {
    stdio: 'inherit',
    env: { ...process.env, ...extraEnv },
  });
  children.push(child);
  child.on('exit', (code) => {
    if (!shuttingDown && code) {
      shutdown(code);
    }
  });
  return child;
}

function shutdown(code = 0) {
  if (shuttingDown) {
    return;
  }
  shuttingDown = true;
  for (const child of children) {
    if (!child.killed) {
      child.kill('SIGTERM');
    }
  }
  process.exit(code);
}

async function waitFor(url, timeoutMs = 120_000) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    try {
      const response = await fetch(url);
      if (response.ok) {
        return;
      }
    } catch {
      // Process is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 300));
  }
  throw new Error(`Timed out waiting for ${url}`);
}

process.on('SIGINT', () => shutdown(0));
process.on('SIGTERM', () => shutdown(0));

start('bash', ['scripts/go-task.sh', 'run-api']);
await waitFor('http://127.0.0.1:8080/health/shallow');

start('bash', ['scripts/go-task.sh', 'run-bff'], {
  HTTP_ADDR: ':8081',
  BFF_COOKIE_SECURE: 'false',
  BFF_DEMO_USERNAME: 'demo',
  BFF_DEMO_PASSWORD: 'demo',
  BACKEND_URL: 'http://127.0.0.1:8080',
});
await waitFor('http://127.0.0.1:8081/health/shallow');

start('pnpm', ['--filter', '@monorepo-project-template/web', 'dev']);
await waitFor('http://127.0.0.1:5173');

await new Promise(() => {});
