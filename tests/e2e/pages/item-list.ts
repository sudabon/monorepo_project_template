import type { Page } from '@playwright/test';

export class ItemListPage {
  constructor(private readonly page: Page) {}

  heading() {
    return this.page.getByRole('heading', {
      name: 'サンプルリソース',
      exact: true,
    });
  }

  search() {
    return this.page.getByLabel('検索');
  }

  createLink() {
    return this.page.getByRole('link', { name: '新規作成' });
  }

  itemLink(name: string) {
    return this.page.getByRole('link', { name });
  }

  nextPage() {
    return this.page.getByRole('button', { name: '次のページ' });
  }

  pageNumber(page: number) {
    return this.page.getByText(`ページ ${page}`, { exact: true });
  }

  emptyState() {
    return this.page.getByText('該当するサンプルリソースはありません');
  }

  async goto() {
    await this.page.goto('/items');
  }
}
