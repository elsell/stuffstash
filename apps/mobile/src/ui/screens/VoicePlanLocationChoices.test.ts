import { expect, it } from 'vitest';
import { voicePlanLocationChoices } from './VoicePlanLocationChoices';

it('keeps the selected opaque destination and only offers earlier proposed parents', () => {
  const commands = [
    { id: 'shelf', kind: 'create_asset', title: 'Shelf', summary: 'Create shelf' },
    { id: 'item', kind: 'create_asset', title: 'Drill', summary: 'Create drill', parentCommandId: 'shelf' },
    { id: 'later', kind: 'create_asset', title: 'Later', summary: 'Create later' }
  ];
  const initial = voicePlanLocationChoices(commands, 'item', {}, '');
  expect(initial.current).toEqual({ kind: 'command', id: 'shelf', label: 'Shelf' });
  expect(initial.proposed.map(option => option.id)).toEqual(['shelf']);
  const draft = voicePlanLocationChoices(commands, 'item', { shelf: { title: 'Garage shelf' }, item: { parent: { kind: 'asset', id: 'bin', label: 'Garage / Bin' } } }, 'garage');
  expect(draft.current).toEqual({ kind: 'asset', id: 'bin', label: 'Garage / Bin' });
  expect(draft.proposed).toEqual([{ kind: 'command', id: 'shelf', label: 'Garage shelf' }]);
  expect(voicePlanLocationChoices(commands, 'item', {}, 'missing').proposed).toEqual([]);
  expect(voicePlanLocationChoices(commands, 'absent', {}, '').valid).toBe(false);
});
