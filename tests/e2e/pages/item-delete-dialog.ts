import type { Page } from '@playwright/test';

export class ItemDeleteDialog {
  constructor(private readonly page: Page) {}

  dialog() {
    return this.page.getByRole('dialog', { name: '削除の確認' });
  }

  cancel() {
    return this.dialog().getByRole('button', { name: 'キャンセル' });
  }

  confirm() {
    return this.dialog().getByRole('button', { name: '削除する' });
  }
}
