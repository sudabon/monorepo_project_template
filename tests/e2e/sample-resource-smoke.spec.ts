import { expect, test } from './fixtures/index.ts';
import { ItemDeleteDialog } from './pages/item-delete-dialog.ts';
import { ItemDetailPage } from './pages/item-detail.ts';
import { ItemFormPage } from './pages/item-form.ts';
import { ItemListPage } from './pages/item-list.ts';

test('参照実装が端から端まで壊れていない', {
  tag: [
    '@phase5-fullstack-vertical-slice',
    '@TP-001',
    '@TP-002',
    '@TP-003',
    '@TP-004',
    '@TP-005',
    '@TP-006',
    '@TP-007',
  ],
}, async ({ page, sampleResourcesMultiPage }) => {
  const list = new ItemListPage(page);
  const detail = new ItemDetailPage(page);
  const form = new ItemFormPage(page);
  const del = new ItemDeleteDialog(page);
  const createdName = `E2E Widget ${Date.now()}`;
  const updatedName = `${createdName} updated`;

  await test.step('TP-001: 作成した項目が一覧に現れる', async () => {
    await list.goto();
    await expect(list.heading()).toBeVisible();
    await list.createLink().click();
    await expect(form.createHeading()).toBeVisible();
    await form.name().fill(createdName);
    await form.save().click();
    await expect(list.heading()).toBeVisible();
    await expect(list.itemLink(createdName)).toBeVisible();
  });

  await test.step('TP-002: 一覧から詳細の内容が表示される', async () => {
    await list.itemLink(createdName).click();
    await expect(detail.heading(createdName)).toBeVisible();
  });

  await test.step('TP-003: 編集した内容が詳細に反映される', async () => {
    await detail.editLink().click();
    await expect(form.editHeading()).toBeVisible();
    await form.name().fill(updatedName);
    await form.save().click();
    await expect(detail.heading(updatedName)).toBeVisible();
  });

  await test.step('TP-004: 削除確認を取り消すと一覧に残る', async () => {
    await detail.deleteButton().click();
    await expect(del.dialog()).toBeVisible();
    await del.cancel().click();
    await expect(del.dialog()).toBeHidden();
    await expect(detail.heading(updatedName)).toBeVisible();
    await detail.backToList().click();
    await expect(list.itemLink(updatedName)).toBeVisible();
  });

  await test.step('TP-005: 削除を承認すると一覧から消える', async () => {
    await list.itemLink(updatedName).click();
    await detail.deleteButton().click();
    await del.confirm().click();
    await expect(list.heading()).toBeVisible();
    await expect(list.itemLink(updatedName)).toHaveCount(0);
  });

  await test.step('TP-006: 検索条件がリロード後も保持される', async () => {
    await list.search().fill(sampleResourcesMultiPage.prefix);
    await expect(list.search()).toHaveValue(sampleResourcesMultiPage.prefix);
    await page.reload();
    await expect(list.search()).toHaveValue(sampleResourcesMultiPage.prefix);
    await expect(
      list.itemLink(`${sampleResourcesMultiPage.prefix} 21`),
    ).toBeVisible();
  });

  await test.step('TP-007: ページ位置がリロード後も保持される', async () => {
    await list.search().fill('');
    await list.nextPage().click();
    await expect(list.pageNumber(2)).toBeVisible();
    await page.reload();
    await expect(list.pageNumber(2)).toBeVisible();
  });
});
