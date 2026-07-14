import { describe, expect, it } from 'vitest';
import { voicePlanCommandEdits } from './VoicePlanEdits';

describe('voicePlanCommandEdits', () => {
  it('sends only changed create fields with explicit root placement', () => {
    expect(voicePlanCommandEdits({
      'cmd-one': { title: '  Holiday towels  ', parent: { kind: 'root', label: 'Inventory root' } }
    })).toEqual([{
      commandId: 'cmd-one',
      title: 'Holiday towels',
      parent: { kind: 'root' }
    }]);
  });

  it('preserves a reviewed existing parent as an opaque asset reference', () => {
    expect(voicePlanCommandEdits({
      'cmd-one': { parent: { kind: 'asset', id: 'asset-bin', label: 'Holiday bin' } }
    })).toEqual([{ commandId: 'cmd-one', parent: { kind: 'asset', id: 'asset-bin' } }]);
  });
});
