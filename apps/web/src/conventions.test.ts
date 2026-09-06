import { readFileSync } from 'node:fs';
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
    const files = [
      'src/bootstrap.tsx',
      'src/main.tsx',
      'src/router.tsx',
      'src/pages/HomePage.tsx',
      'src/pages/LoginPage.tsx',
      'src/routes/__root.tsx',
      'src/routes/login.tsx',
      'src/routes/_authenticated.tsx',
      'src/routes/_authenticated/index.tsx',
    ];
    for (const file of files) {
      const source = readFileSync(join(webRoot, file), 'utf8');
      expect(source).not.toMatch(/\bSuspense\b/);
    }
  });
});
