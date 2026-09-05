import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { existsSync, globSync, readFileSync } from 'node:fs';

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

const pinnedPnpm = root.packageManager.slice('pnpm@'.length);
const pnpmBin = process.env.PNPM || 'pnpm';
const pnpmVersion = spawnSync(pnpmBin, ['--version'], {
  encoding: 'utf8',
  stdio: ['ignore', 'pipe', 'pipe'],
});
if (pnpmVersion.error || pnpmVersion.status !== 0) {
  const detail = pnpmVersion.error?.message ?? pnpmVersion.stderr.trim();
  assert.fail(
    `pnpm ${pinnedPnpm} is not runnable (${pnpmBin})${detail ? `: ${detail}` : ''}`,
  );
}
const runningPnpm = pnpmVersion.stdout.trim().split('+')[0];
assert.equal(
  runningPnpm,
  pinnedPnpm,
  `Use pnpm ${pinnedPnpm} (running ${runningPnpm})`,
);

function workspacePackageGlobs() {
  const yaml = readFileSync('pnpm-workspace.yaml', 'utf8');
  const globs = [];
  let inPackages = false;
  for (const raw of yaml.split('\n')) {
    const line = raw.replace(/#.*$/, '');
    if (/^packages:\s*$/.test(line)) {
      inPackages = true;
      continue;
    }
    if (!inPackages) continue;
    const item = /^\s+-\s+(.+?)\s*$/.exec(line);
    if (item) {
      globs.push(item[1].replace(/^['"]|['"]$/g, ''));
      continue;
    }
    if (line.trim() !== '') break;
  }
  assert.ok(globs.length > 0, 'pnpm-workspace.yaml: packages list is required');
  return globs.map((glob) =>
    glob.endsWith('package.json')
      ? glob
      : `${glob.replace(/\/$/, '')}/package.json`,
  );
}

const manifests = ['package.json'];
if (existsSync('pnpm-workspace.yaml')) {
  manifests.push(...workspacePackageGlobs());
}

for (const path of globSync(manifests)) {
  const manifest = JSON.parse(readFileSync(path, 'utf8'));
  for (const field of [
    'dependencies',
    'devDependencies',
    'optionalDependencies',
    'peerDependencies',
  ]) {
    for (const [name, version] of Object.entries(manifest[field] ?? {})) {
      if (version.startsWith('workspace:')) continue;
      assert.match(version, exact, `${path}: pin ${name} exactly`);
    }
  }
}

console.log(
  `Toolchain pins OK: Node.js ${nodeVersion}, ${root.packageManager}`,
);
