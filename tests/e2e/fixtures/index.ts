import { expect, test as base } from '@playwright/test';

type Fixtures = {
  authenticatedUser: undefined;
  sampleResourcesMultiPage: { prefix: string };
};

export const test = base.extend<Fixtures>({
  authenticatedUser: [
    async ({ page }, use) => {
      await page.goto('/login');
      await page.getByLabel('ユーザ名').fill('demo');
      await page.getByLabel('パスワード').fill('demo');
      await page.getByRole('button', { name: 'ログイン' }).click();
      await expect(page.getByRole('heading', { name: 'ホーム' })).toBeVisible();
      await use();
    },
    { auto: true },
  ],
  sampleResourcesMultiPage: async ({ page }, use) => {
    const prefix = `Pager ${Date.now()}`;
    const sessionResponse = await page.request.get('/auth/session');
    expect(sessionResponse.ok()).toBeTruthy();
    const session = (await sessionResponse.json()) as { csrfToken: string };
    for (let index = 1; index <= 21; index += 1) {
      const response = await page.request.post('/api/items', {
        data: { name: `${prefix} ${String(index).padStart(2, '0')}` },
        headers: { 'X-CSRF-Token': session.csrfToken },
      });
      expect(response.ok(), await response.text()).toBeTruthy();
    }
    await use({ prefix });
  },
});

export { expect } from '@playwright/test';
