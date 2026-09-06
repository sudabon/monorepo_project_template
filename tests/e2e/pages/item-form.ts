import type { Page } from '@playwright/test';

export class ItemFormPage {
  constructor(private readonly page: Page) {}

  createHeading() {
    return this.page.getByRole('heading', { name: 'サンプルリソースを作成' });
  }

  editHeading() {
    return this.page.getByRole('heading', { name: 'サンプルリソースを編集' });
  }

  name() {
    return this.page.getByLabel('名前');
  }

  description() {
    return this.page.getByLabel('説明');
  }

  save() {
    return this.page.getByRole('button', { name: '保存' });
  }
}
