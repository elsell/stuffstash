import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { InventoryInvitationScreen } from './InventoryInvitationScreen';

it('offers concise invitation review and an explicit start-over action without accepting', async () => {
  const harness = new MobileRenderHarness();
  let accepts = 0;
  let resets = 0;
  await harness.render(<InventoryInvitationScreen initialized invalidLink={false}
    reference={{ tenantId: 'tenant', inventoryId: 'inventory', invitationId: 'invitation', acceptanceToken: 'A'.repeat(43) }}
    previewQuery={{ async execute() { return { inventoryId: 'inventory', inventoryName: 'Home Inventory', relationship: 'viewer',
      status: 'pending', isExpired: false, expiresAt: '2030-01-01T00:00:00Z' }; } }}
    acceptCommand={{ async execute() { accepts++; throw new Error('Not requested'); } }}
    onAccepted={async () => {}} onDismiss={() => {}} onSwitchAccount={() => {}}
    onStartOver={async () => { resets++; }} />);
  await harness.settle();
  expect(harness.byText('You’re invited')).toBeDefined();
  expect(harness.byText('Home Inventory')).toBeDefined();
  await harness.press(harness.byLabel('Sign out and start over'));
  expect(resets).toBe(1);
  expect(accepts).toBe(0);
  await harness.unmount();
});

it('offers native dismissal and recovery commands without joining', async () => {
  const h = new MobileRenderHarness();
  let dismissals = 0;
  let previews = 0;
  const props = {
    initialized: true, invalidLink: false,
    reference: { tenantId: 'tenant', inventoryId: 'inventory', invitationId: 'invitation', acceptanceToken: 'A'.repeat(43) },
    previewQuery: { async execute() { previews++; throw new Error('offline'); } },
    acceptCommand: { async execute() { throw new Error('Must not accept'); } },
    onAccepted: async () => {}, onDismiss: () => { dismissals++; }, onSwitchAccount: () => {}
  };
  try {
    await h.render(<InventoryInvitationScreen {...props} />);
    await h.press(h.byLabel('Try again'));
    expect(previews).toBe(2);
    await h.press(h.byLabel('Done'));
    expect(dismissals).toBe(1);
  } finally { await h.unmount(); }
});
