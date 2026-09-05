import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { test } from 'node:test';

const script = new URL('../../scripts/check-toolchain.mjs', import.meta.url)
  .pathname;

function fixture(t) {
  const cwd = mkdtempSync(join(tmpdir(), 'toolchain-pins-'));
  t.after(() => rmSync(cwd, { recursive: true, force: true }));
  const manifest = {
    engines: { node: process.versions.node },
    packageManager: 'pnpm@11.25.0',
  };
  writeFileSync(join(cwd, '.node-version'), `${process.versions.node}\n`);
  return {
    cwd,
    manifest,
    run() {
      writeFileSync(join(cwd, 'package.json'), JSON.stringify(manifest));
      return spawnSync(process.execPath, [script], { cwd, encoding: 'utf8' });
    },
  };
}

test('accepts matching exact runtime pins', (t) => {
  assert.equal(fixture(t).run().status, 0);
});

test('rejects drift between .node-version and engines.node', (t) => {
  const repo = fixture(t);
  repo.manifest.engines.node = '0.0.1';
  assert.notEqual(repo.run().status, 0);
});

test('rejects a Node runtime different from the configured pin', (t) => {
  const repo = fixture(t);
  repo.manifest.engines.node = '0.0.1';
  writeFileSync(join(repo.cwd, '.node-version'), '0.0.1\n');
  assert.notEqual(repo.run().status, 0);
});

test('rejects packageManager ranges', (t) => {
  const repo = fixture(t);
  repo.manifest.packageManager = 'pnpm@^11.25.0';
  assert.notEqual(repo.run().status, 0);
});

test('rejects dependency ranges in a workspace package', (t) => {
  const repo = fixture(t);
  mkdirSync(join(repo.cwd, 'apps/web'), { recursive: true });
  writeFileSync(
    join(repo.cwd, 'apps/web/package.json'),
    JSON.stringify({ dependencies: { example: '^1.2.3' } }),
  );
  assert.notEqual(repo.run().status, 0);
});
