import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const script = fileURLToPath(
  new URL('../../scripts/check-test-plan.sh', import.meta.url),
);

function fixture(t, config) {
  const cwd = mkdtempSync(join(tmpdir(), 'test-plan-'));
  t.after(() => rmSync(cwd, { recursive: true, force: true }));
  function git(...args) {
    const result = spawnSync('git', args, { cwd, encoding: 'utf8' });
    assert.equal(result.status, 0, result.stderr);
    return result.stdout.trim();
  }
  function write(path, content) {
    const fullPath = join(cwd, path);
    mkdirSync(dirname(fullPath), { recursive: true });
    writeFileSync(fullPath, content);
  }
  function commit() {
    git('add', '.');
    git(
      '-c',
      'user.name=Toolchain Test',
      '-c',
      'user.email=test@example.invalid',
      '-c',
      'commit.gpgsign=false',
      'commit',
      '-qm',
      'fixture',
    );
  }
  git('init', '-q', '-b', 'main');
  write('README.md', 'fixture\n');
  if (config !== undefined)
    write('.openspec-e2e-kit.json', JSON.stringify(config));
  commit();
  git('update-ref', 'refs/remotes/origin/main', 'HEAD');
  return {
    cwd,
    git,
    write,
    commit,
    run(base) {
      return spawnSync('bash', [script, ...(base ? [base] : [])], {
        cwd,
        encoding: 'utf8',
      });
    },
  };
}

test('templateRepo=true skips before resolving the base revision', (t) => {
  const repo = fixture(t, { version: '0.1.0', templateRepo: true });
  const result = repo.run('missing-base');
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /templateRepo.*skip/);
});

for (const [name, config] of [
  ['false', { templateRepo: false }],
  ['missing field', {}],
  ['missing config', undefined],
  ['string true', { templateRepo: 'true' }],
]) {
  test(`${name} keeps the guard active for a plan without tagged tests`, (t) => {
    const repo = fixture(t, config);
    repo.write('openspec/changes/example/test-plan.md', 'TP-001\n');
    repo.commit();
    const result = repo.run();
    assert.equal(result.status, 1, result.stderr);
    assert.match(result.stdout, /@example/);
  });
}

test('a change without a plan does not require E2E tests', (t) => {
  const repo = fixture(t, {});
  repo.write('openspec/changes/backend/proposal.md', 'Backend only\n');
  repo.commit();
  const result = repo.run();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  assert.match(result.stdout, /backend.*skip/);
});

test('a planned change with a tagged test passes', (t) => {
  const repo = fixture(t, {});
  repo.write('openspec/changes/example/test-plan.md', 'TP-001\n');
  repo.write(
    'tests/e2e/example.spec.ts',
    "test('example @example', () => {});\n",
  );
  repo.commit();
  const result = repo.run();
  assert.equal(result.status, 0, result.stdout + result.stderr);
});

test('a shorter change id does not match a longer tag', (t) => {
  const repo = fixture(t, {});
  repo.write('openspec/changes/phase1/test-plan.md', 'TP-001\n');
  repo.write(
    'tests/e2e/example.spec.ts',
    "test('example @phase10', () => {});\n",
  );
  repo.commit();
  const result = repo.run();
  assert.equal(result.status, 1, result.stderr);
  assert.match(result.stdout, /@phase1/);
});

test('a change without a plan does not mask another change with missing tests', (t) => {
  const repo = fixture(t, {});
  repo.write('openspec/changes/backend/proposal.md', 'Backend only\n');
  repo.write('openspec/changes/frontend/test-plan.md', 'TP-001\n');
  repo.commit();
  const result = repo.run();
  assert.equal(result.status, 1, result.stderr);
  assert.match(result.stdout, /@frontend/);
});

for (const path of [
  'README.md',
  'openspec/changes/archive/2026-09-05-example/proposal.md',
]) {
  test(`irrelevant changes (${path}) exit successfully`, (t) => {
    const repo = fixture(t, {});
    repo.write(path, 'Changed\n');
    repo.commit();
    const result = repo.run();
    assert.equal(result.status, 0, result.stdout + result.stderr);
  });
}

test('invalid JSON fails instead of disabling the guard', (t) => {
  const repo = fixture(t, {});
  repo.write('.openspec-e2e-kit.json', '{');
  repo.write('openspec/changes/example/test-plan.md', 'TP-001\n');
  repo.write(
    'tests/e2e/example.spec.ts',
    "test('example @example', () => {});\n",
  );
  repo.commit();
  assert.notEqual(repo.run().status, 0);
});

test('an unresolved diff base fails instead of skipping the guard', (t) => {
  const repo = fixture(t, {});
  assert.notEqual(repo.run('missing-base').status, 0);
});

test('fetching full history restores the PR base and the intended three-dot diff', (t) => {
  const repo = fixture(t, {});
  repo.git('checkout', '-qb', 'feature');
  repo.write('openspec/changes/example/test-plan.md', 'TP-001\n');
  repo.commit();
  const clone = mkdtempSync(join(tmpdir(), 'test-plan-shallow-'));
  t.after(() => rmSync(clone, { recursive: true, force: true }));
  function git(...args) {
    const result = spawnSync('git', args, { cwd: clone, encoding: 'utf8' });
    assert.equal(result.status, 0, result.stderr);
    return result.stdout.trim();
  }
  git(
    'clone',
    '--quiet',
    '--depth=1',
    '--branch=feature',
    `file://${repo.cwd}`,
    '.',
  );
  assert.equal(git('rev-parse', '--is-shallow-repository'), 'true');
  git(
    'fetch',
    '--quiet',
    '--unshallow',
    'origin',
    '+refs/heads/*:refs/remotes/origin/*',
  );
  assert.equal(git('rev-parse', '--is-shallow-repository'), 'false');
  assert.equal(
    git('diff', '--name-only', 'origin/main...HEAD'),
    'openspec/changes/example/test-plan.md',
  );
  const result = spawnSync('bash', [script], { cwd: clone, encoding: 'utf8' });
  assert.equal(result.status, 1, result.stderr);
  assert.match(result.stdout, /@example/);
});
