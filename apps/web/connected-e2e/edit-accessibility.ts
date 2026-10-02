import { writeFile } from 'node:fs/promises';
import { expect, type Page, type TestInfo } from '@playwright/test';

/** Observe production focus handling after real sign-in; never synthesize focus. */
export async function verifyEditAccessibility(page: Page, assetPath: string, title: string, info: TestInfo) {
  const viewport = page.viewportSize();
  await page.setViewportSize({ width: 320, height: 800 });
  await page.goto(`${assetPath}/edit`);
  const dialog = page.getByRole('dialog', { name: 'Edit asset' });
  await expect(dialog).toBeVisible();
  await expect(dialog.getByLabel('Name', { exact: true })).toBeFocused();
  await page.keyboard.press('ControlOrMeta+A');
  await page.keyboard.type(`${title} unsaved`);
  await expect(dialog.getByLabel('Name', { exact: true })).toHaveValue(`${title} unsaved`);
  const observations: { key: string; insideDialog: boolean; label: string }[] = [];
  const cancel = dialog.getByRole('link', { name: 'Cancel', exact: true });
  // A bounded reverse traversal exercises the actual focus trap, including wrap.
  let reachedCancel = false;
  for (let index = 0; index < 20; index++) {
    await page.keyboard.press('Shift+Tab');
    const state = await dialog.evaluate(element => {
      const focused = document.activeElement;
      return { insideDialog: element.contains(focused), label: focused?.getAttribute('aria-label') ?? focused?.textContent?.trim() ?? '' };
    });
    observations.push({ key: 'Shift+Tab', ...state });
    expect(state.insideDialog).toBe(true);
    if (await cancel.evaluate(element => element === document.activeElement)) { reachedCancel = true; break; }
  }
  expect(reachedCancel).toBe(true);
  for (const control of [cancel, dialog.getByRole('button', { name: 'Save', exact: true })]) {
    await control.scrollIntoViewIfNeeded();
    const box = await control.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.x).toBeGreaterThanOrEqual(0);
    expect(box!.x + box!.width).toBeLessThanOrEqual(320);
  }
  const geometry = await dialog.evaluate(element => ({ width: element.clientWidth, contentWidth: element.scrollWidth }));
  expect(geometry.contentWidth).toBeLessThanOrEqual(geometry.width + 1);
  await page.screenshot({ path: info.outputPath('connected-edit-narrow.png') });
  await expect(cancel).toBeFocused();
  await page.keyboard.press('Enter');
  await expect(dialog).toBeHidden();
  await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
  await expect(page).toHaveURL(new RegExp(`${assetPath}$`));
  await page.reload();
  await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
  await writeFile(info.outputPath('connected-edit-keyboard.json'), JSON.stringify({ viewport: { width: 320, height: 800 }, observations, geometry, cancelledWithoutSaving: true }, null, 2));
  if (viewport) await page.setViewportSize(viewport);
}
