import { expect, test, type Page } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

async function installConversationServer(page: Page, propose: boolean) {
  const decisions: string[] = [];
  await page.routeWebSocket('ws://127.0.0.1:18080/v1/realtime/voice', socket => {
    let seq = 1; let authenticated = false;
    const send = (event: object) => socket.send(JSON.stringify({ seq: seq++, sessionId: 'web-session', ...event }));
    socket.onMessage(raw => {
      const message = JSON.parse(String(raw));
      if (message.type === 'session.authenticate') { expect(message.authorization).toMatch(/^Bearer /); authenticated = true; socket.send(JSON.stringify({ type: 'session.authenticated' })); return; }
      expect(authenticated).toBe(true);
      if (message.type === 'session.start') { expect(message.tenantId).toBe('tenant-home'); expect(message.inventoryId).toBe('inventory-household'); send({ type: 'session.started' }); }
      if (message.type === 'text.input') {
        if (propose) send({ type: 'action.plan.proposed', actionPlan: { planId: 'move-plan', confirmationSummary: 'Move the camping tent into Garage.', commands: [{ summary: 'Move Camping tent', parentTitle: 'Garage' }], risks: [] } });
        else { send({ type: 'assistant.response.completed', response: { sessionId: 'web-session', tenantId: 'tenant-home', inventoryId: 'inventory-household', displayResponse: 'The camping tent is in Garage.', artifacts: [] } }); send({ type: 'session.completed', followUpAvailable: true }); }
      }
      if (message.type === 'action.plan.approve' || message.type === 'action.plan.cancel') { decisions.push(message.type); send({ type: message.type === 'action.plan.approve' ? 'action.plan.executed' : 'action.plan.cancelled' }); }
    });
  });
  return decisions;
}

test.beforeEach(async ({ page }) => { resetWorkspaceApiState(page); await installAuthenticatedWorkspace(page); });
test('typed conversation keeps composer reachable and restores focus', async ({ page }, testInfo) => {
  await installConversationServer(page, false);
  await page.goto('/tenants/tenant-home/inventories/inventory-household');
  const opener = page.getByRole('button', { name: 'Ask Stuff Stash', exact: true }); await opener.click();
  const dialog = page.getByRole('dialog'); const composer = page.getByRole('textbox', { name: 'Message Stuff Stash' });
  await expect(composer).toBeFocused(); await composer.fill('Where is the tent?'); await composer.press('Enter');
  await expect(dialog.getByText('The camping tent is in Garage.', { exact: true }).first()).toBeVisible();
  await expect(composer).toBeEnabled();
  const bounds = await composer.boundingBox(); expect(bounds!.y + bounds!.height).toBeLessThanOrEqual(page.viewportSize()!.height);
  await page.screenshot({ path: testInfo.outputPath('inventory-conversation.png'), fullPage: true });
  await composer.press('Escape'); await expect(opener).toBeFocused(); await expect(dialog).not.toBeVisible();
});
for (const approve of [false, true]) {
  test(`conversation ${approve ? 'approves' : 'cancels'} only after explicit review`, async ({ page }) => {
    const decisions = await installConversationServer(page, true);
    await page.goto('/tenants/tenant-home/inventories/inventory-household');
    await page.getByRole('button', { name: 'Ask Stuff Stash', exact: true }).click();
    await page.getByRole('textbox', { name: 'Message Stuff Stash' }).fill('Move the tent into Garage'); await page.getByRole('button', { name: 'Send', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Review changes', exact: true })).toBeVisible(); expect(decisions).toEqual([]);
    await page.getByRole('button', { name: approve ? 'Approve changes' : 'Cancel changes', exact: true }).click();
    await expect.poll(() => decisions).toEqual([approve ? 'action.plan.approve' : 'action.plan.cancel']);
    await expect(page.getByRole('heading', { name: 'Review changes', exact: true })).not.toBeVisible();
  });
}
