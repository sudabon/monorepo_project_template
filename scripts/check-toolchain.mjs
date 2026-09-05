import assert from 'node:assert/strict';
import { globSync, readFileSync } from 'node:fs';

const exact = /^\d+\.\d+\.\d+$/;
const root = JSON.parse(readFileSync('package.json', 'utf8'));
const nodeVersion = readFileSync('.node-version', 'utf8').trim();

assert.match(
  nodeVersion,
  exact,
  '.node-version must pin an exact stable version',
);
assert.equal(
  root.engines?.node,
  nodeVersion,
  'engines.node must match .node-version',
);
assert.equal(process.versions.node, nodeVersion, `Use Node.js ${nodeVersion}`);
assert.match(
  root.packageManager ?? '',
  /^pnpm@\d+\.\d+\.\d+$/,
  'Pin pnpm in packageManager',
);

for (const path of globSync([
  'package.json',
  'apps/*/package.json',
  'packages/*/package.json',
])) {
  const manifest = JSON.parse(readFileSync(path, 'utf8'));
  for (const field of [
    'dependencies',
    'devDependencies',
    'optionalDependencies',
    'peerDependencies',
  ]) {
    for (const [name, version] of Object.entries(manifest[field] ?? {})) {
      assert.match(
        version.replace(/^workspace:/, ''),
        exact,
        `${path}: pin ${name} exactly`,
      );
    }
  }
}

console.log(
  `Toolchain pins OK: Node.js ${nodeVersion}, ${root.packageManager}`,
);
