import { expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { localization, t } from '$lib/presentation/localization';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import CaseSummary from './CaseSummary.svelte';

it('presents complete expectations and locale lists without translating authored names', async () => {
  const target = document.createElement('div'); document.body.append(target);
  const value: CaseDefinition = { title: 'Camping', utterance: 'Find my tent', assets: [
    { id: 'tent', title: 'My tent', kind: 'item', description: '', parentId: '', tagNames: ['Summer', 'Camping'] },
    { id: 'box', title: 'Summer box', kind: 'container', description: '', parentId: '', tagNames: [] }
  ], expectations: { kind: 'answer', referencedAssets: ['tent', 'box'], locations: [], proposals: [], forbiddenOperations: ['archive', 'checkout'] } };
  const component = mount(CaseSummary, { target, props: { value } });
  try {
    expect(target.textContent).toContain(t('evaluation.mustMention', { items: localization.list(['My tent', 'Summer box']) }));
    expect(target.textContent).toContain(t('web.CaseSummary.tagsFull', { value: localization.list(['Summer', 'Camping']) }));
    expect(target.textContent).toContain(t('web.CaseSummary.forbiddenChangesFull', { value: localization.list([t('case.operation.archive'), t('case.operation.checkout')]) }));
  } finally { await unmount(component); target.remove(); }
});
