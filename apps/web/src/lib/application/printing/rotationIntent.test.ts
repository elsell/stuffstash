import { expect, it } from 'vitest';
import { pairingReturnPath, rotationIntent } from './rotationIntent';
it('keeps an exact public rotation target through sign-in and refuses partial or duplicate targets', () => {
    const target = { tenantId: 't', inventoryId: 'i', connectorId: 'c' };
    const path = pairingReturnPath('pair', target);
    expect(rotationIntent(new URL(path, 'https://stash.example').searchParams)).toEqual(target);
    expect(rotationIntent(new URLSearchParams('tenantId=t'))).toBeNull();
    expect(rotationIntent(new URLSearchParams('tenantId=t&tenantId=other&inventoryId=i&connectorId=c'))).toBeNull();
    expect(rotationIntent(new URLSearchParams())).toBeUndefined();
});
