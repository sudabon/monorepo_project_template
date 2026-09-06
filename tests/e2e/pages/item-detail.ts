import type { Page } from '@playwright/test';

export class ItemDetailPage {
  constructor(private readonly page: Page) {}

  heading(name: string) {
    return this.page.getByRole('heading', { name });
  }

  editLink() {
    return this.page.getByRole('link', { name: '編集' });
  }

  deleteButton() {
    return this.page.getByRole('button', { name: '削除' });
  }

  backToList() {
    return this.page.getByRole('link', { name: '一覧へ戻る' });
  }
}
