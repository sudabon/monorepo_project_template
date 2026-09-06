import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const webRoot = join(dirname(fileURLToPath(import.meta.url)), '..');

const forbidden = [
  '@mui/',
  '@material-ui/',
  'antd',
  '@chakra-ui/',
  '@headlessui/',
  '@radix-ui/',
  '@mantine/',
  'react-bootstrap',
  'bootstrap',
  '@fluentui/',
  'primereact',
  '@nextui-org/',
  '@heroui/',
  'reactstrap',
  'semantic-ui-react',
  '@ariakit/',
  '@base-ui-components/',
  'shadcn',
];

function appSources(): string[] {
  return readdirSync(join(webRoot, 'src'), {
    recursive: true,
    encoding: 'utf8',
  }).filter(
    (entry) =>
      (entry.endsWith('.ts') || entry.endsWith('.tsx')) &&
      !entry.endsWith('.test.ts') &&
      !entry.endsWith('.test.tsx') &&
      entry !== 'routeTree.gen.ts',
  );
}

describe('template conventions', () => {
  it('does not declare UI component libraries as dependencies', () => {
    const manifest = JSON.parse(
      readFileSync(join(webRoot, 'package.json'), 'utf8'),
    ) as {
      dependencies?: Record<string, string>;
      devDependencies?: Record<string, string>;
    };
    const names = [
      ...Object.keys(manifest.dependencies ?? {}),
      ...Object.keys(manifest.devDependencies ?? {}),
    ];
    for (const name of names) {
      expect(
        forbidden.some(
          (pattern) => name === pattern || name.startsWith(pattern),
        ),
        `${name} is a UI component library`,
      ).toBe(false);
    }
  });

  it('does not use React.Suspense for loading', () => {
    // Walk src instead of listing files: an enumerated list silently stops
    // covering whatever is added next.
    const files = appSources();
    expect(files.length).toBeGreaterThan(20);
    for (const file of files) {
      const source = readFileSync(join(webRoot, 'src', file), 'utf8');
      expect(source, `${file} uses Suspense`).not.toMatch(/\bSuspense\b/);
    }
  });
});
