import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { useInvitationRouteActions } from './useInvitationRouteActions';

it.each(['openInventory', 'startOver'] as const)('keeps %s navigation in its original invitation session', async action => {
  const h = new MobileRenderHarness();
  const reference = { tenantId: 'tenant', inventoryId: 'one', invitationId: 'one', acceptanceToken: 'secret' };
  let actions!: ReturnType<typeof useInvitationRouteActions>;
  let finish!: () => void;
  let cleared = 0; let returned = 0;
  const work = () => new Promise<void>(resolve => { finish = resolve; });
  function Probe({ selected = reference }) {
    actions = useInvitationRouteActions({ reference: selected, clear: () => { cleared++; }, goHome: () => { returned++; }, selectInventory: work, startOver: work });
    return null;
  }
  try {
    for (const departure of ['replace', 'return', 'unmount', 'none']) {
      await h.render(<Probe />);
      let completion!: Promise<unknown>;
      await h.run(() => { completion = actions[action]('one'); });
      if (departure === 'replace') await h.render(<Probe selected={{ ...reference, invitationId: 'two' }} />);
      if (departure === 'unmount') await h.render(<></>);
      if (departure === 'return') { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
      await h.run(async () => { finish(); await completion; });
      expect(cleared).toBe(departure === 'none' ? 1 : 0);
      expect(returned).toBe(cleared);
    }
  } finally { await h.unmount(); setScreenFocused(true); }
});

it('propagates operation errors without clearing the invitation or navigating', async () => {
  const h = new MobileRenderHarness(); let actions!: ReturnType<typeof useInvitationRouteActions>;
  const failure = new Error('offline'); let navigation = 0;
  function Probe() {
    actions = useInvitationRouteActions({ clear: () => { navigation++; }, goHome: () => { navigation++; }, selectInventory: async () => { throw failure; }, startOver: async () => { throw failure; } });
    return null;
  }
  try {
    await h.render(<Probe />);
    await expect(actions.openInventory('one')).rejects.toBe(failure);
    await expect(actions.startOver()).rejects.toBe(failure);
    expect(navigation).toBe(0);
  } finally { await h.unmount(); }
});
