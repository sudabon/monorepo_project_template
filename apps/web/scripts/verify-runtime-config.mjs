import {
  createReadStream,
  existsSync,
  readFileSync,
  writeFileSync,
} from 'node:fs';
import { createServer } from 'node:http';
import { extname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const webRoot = join(fileURLToPath(new URL('.', import.meta.url)), '..');
const distDir = join(webRoot, 'dist');

if (!existsSync(join(distDir, 'config.json'))) {
  throw new Error(
    'apps/web/dist/config.json is missing. Run the web build first.',
  );
}

const types = new Map([
  ['.html', 'text/html; charset=utf-8'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.css', 'text/css; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.svg', 'image/svg+xml'],
]);

function serve(root, port) {
  return new Promise((resolve, reject) => {
    const server = createServer((req, res) => {
      const url = new URL(req.url ?? '/', `http://127.0.0.1:${port}`);
      const relative = url.pathname === '/' ? '/index.html' : url.pathname;
      const filePath = join(root, relative);
      if (!filePath.startsWith(root) || !existsSync(filePath)) {
        res.statusCode = 404;
        res.end('not found');
        return;
      }
      res.setHeader(
        'Content-Type',
        types.get(extname(filePath)) ?? 'application/octet-stream',
      );
      createReadStream(filePath).pipe(res);
    });
    server.on('error', reject);
    server.listen(port, '127.0.0.1', () => resolve(server));
  });
}

const original = JSON.parse(readFileSync(join(distDir, 'config.json'), 'utf8'));
const swapped = {
  apiBaseUrl: 'https://other.example/api',
  authBaseUrl: 'https://other.example/auth',
};
writeFileSync(join(distDir, 'config.json'), `${JSON.stringify(swapped)}\n`);

const server = await serve(distDir, 4179);
try {
  const response = await fetch('http://127.0.0.1:4179/config.json', {
    cache: 'no-store',
  });
  const body = await response.json();
  if (
    body.apiBaseUrl !== swapped.apiBaseUrl ||
    body.authBaseUrl !== swapped.authBaseUrl
  ) {
    throw new Error(`expected swapped base URLs, got ${JSON.stringify(body)}`);
  }
} finally {
  writeFileSync(join(distDir, 'config.json'), `${JSON.stringify(original)}\n`);
  server.close();
}

console.log(
  'Runtime config swap OK: the same dist serves different base URLs from config.json.',
);
