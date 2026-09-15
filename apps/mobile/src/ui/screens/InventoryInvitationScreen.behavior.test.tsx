import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setWindowFontScaleForTest } from '../../test-support/react-native';
import { InventoryInvitationScreen } from './InventoryInvitationScreen';
import { InventoryInvitationEmailMismatchError, type InventoryInvitationPreview,
  type InventoryInvitationAcceptance } from '../../application/invitations/InventoryInvitationRepository';

const referenceA = { tenantId: 'tenant-a', inventoryId: 'inventory-a', invitationId: 'invitation-a', acceptanceToken: 'A'.repeat(43) };
const referenceB = { ...referenceA, inventoryId: 'inventory-b', invitationId: 'invitation-b' };
const acceptance: InventoryInvitationAcceptance = { ...referenceA, principalId: 'principal', relationship: 'viewer', status: 'accepted' };
type Props = Parameters<typeof InventoryInvitationScreen>[0];
function props(overrides: Partial<Props> = {}): Props {
  return { initialized: true, invalidLink: false, reference: referenceA,
    previewQuery: { execute: async () => preview('Kitchen inventory') },
    acceptCommand: { execute: async () => acceptance },
    onAccepted: async () => {}, onDismiss: () => {}, onSwitchAccount: () => {}, ...overrides };
}
function preview(inventoryName: string, overrides: Partial<InventoryInvitationPreview> = {}): InventoryInvitationPreview {
  return { inventoryId: 'inventory-a', inventoryName, relationship: 'viewer', status: 'pending',
    isExpired: false, expiresAt: '2030-01-02T15:04:05Z', ...overrides };
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(done => { resolve = done; });
  return { promise, resolve };
}

it('requires explicit acceptance before opening', async () => {
  const h = new MobileRenderHarness(); const accepted: unknown[] = []; const opened: string[] = [];
  try {
    await h.render(<InventoryInvitationScreen {...props({
      acceptCommand: { execute: async reference => { accepted.push(reference); return acceptance; } },
      onAccepted: async id => { opened.push(id); }
    })} />);
    expect(accepted).toEqual([]); expect(opened).toEqual([]);
    await h.press(h.byLabel('Join inventory'));
    expect(accepted).toEqual([referenceA]); expect(opened).toEqual([]);
    await h.press(h.byLabel('Open inventory'));
    expect(opened).toEqual(['inventory-a']);
  } finally { await h.unmount(); }
});

it('ignores a late preview after a newer invitation becomes current', async () => {
  const h = new MobileRenderHarness(); const old = deferred<InventoryInvitationPreview>();
  const current = deferred<InventoryInvitationPreview>();
  const previewQuery: Props['previewQuery'] = { execute: ref => ref.invitationId === referenceA.invitationId ? old.promise : current.promise };
  try {
    await h.render(<InventoryInvitationScreen {...props({ previewQuery })} />);
    await h.render(<InventoryInvitationScreen {...props({ previewQuery, reference: referenceB })} />);
    await h.run(() => current.resolve(preview('Garage inventory', { inventoryId: 'inventory-b' })));
    await h.run(() => old.resolve(preview('Old inventory')));
    expect(h.allText()).toContain('Garage inventory'); expect(h.allText()).not.toContain('Old inventory');
  } finally { await h.unmount(); }
});

it('ignores a late acceptance after the invitation changes', async () => {
  const h = new MobileRenderHarness(); const old = deferred<InventoryInvitationAcceptance>();
  const previewQuery: Props['previewQuery'] = { execute: async ref => preview(ref.inventoryId) };
  try {
    await h.render(<InventoryInvitationScreen {...props({ previewQuery, acceptCommand: { execute: () => old.promise } })} />);
    await h.press(h.byLabel('Join inventory'));
    await h.render(<InventoryInvitationScreen {...props({ previewQuery, reference: referenceB })} />);
    await h.run(() => old.resolve(acceptance));
    expect(h.allText()).toContain('inventory-b'); expect(h.allText()).not.toContain('You’re in');
  } finally { await h.unmount(); }
});

it('offers account switching for a mismatched identity', async () => {
  const h = new MobileRenderHarness(); let switches = 0;
  try {
    await h.render(<InventoryInvitationScreen {...props({ onSwitchAccount: () => { switches++; },
      previewQuery: { execute: async () => { throw new InventoryInvitationEmailMismatchError(); } } })} />);
    await h.press(h.byLabel('Switch account')); expect(switches).toBe(1);
    expect(h.byLabel('Join inventory')).toBeUndefined();
  } finally { await h.unmount(); }
});

it.each(['expired', 'revoked', 'cancelled'] as const)('does not allow joining a %s invitation', async status => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<InventoryInvitationScreen {...props({ previewQuery: { execute: async () => preview('Kitchen', { status, isExpired: status === 'expired' }) } })} />);
    expect(h.allText()).toContain(`Invitation ${status}`);
    expect(h.byLabel('Join inventory')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('opens an already accepted invitation without joining again', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<InventoryInvitationScreen {...props({ previewQuery: { execute: async () => preview('Kitchen', { status: 'accepted' }) } })} />);
    expect(h.allText()).toContain('You’re in');
    expect(h.byLabel('Open inventory')).toBeDefined(); expect(h.byLabel('Join inventory')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('preserves the adaptive scroll layout without capping text scaling', async () => {
  const h = new MobileRenderHarness(); setWindowFontScaleForTest(3);
  try {
    await h.render(<InventoryInvitationScreen {...props()} />);
    expect(h.byType('ScrollView')?.props.contentContainerStyle.justifyContent).toBe('flex-start');
    expect(h.allByType('Text').every(node => node.props.maxFontSizeMultiplier === undefined)).toBe(true);
  } finally { await h.unmount(); setWindowFontScaleForTest(1); }
});
