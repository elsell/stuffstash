import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { SearchHeader, type SearchHeaderProps } from './BrowseHeader';
import { lightPalette } from '../theme/tokens';

it('offers expiration navigation only inside Filters, with explicit date modes', async () => {
  const h = new MobileRenderHarness();
  const modes: string[] = [];
  const noop = () => {};
  const props: SearchHeaderProps = {
    isLoading: false, lifecycleState: 'active', checkoutState: 'any',
    filterDraft: { scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [] },
    inventoryContext: 'Home', palette: lightPalette, query: '', resultCount: 0, scope: 'all',
    selectedSurface: 'list', selectedTagIds: [], filtersExpanded: false, sort: 'updated_desc', submittedQuery: '',
    onChangeSurface: noop, onApplyFilters: noop, onChangeDraftLifecycleState: noop,
    onChangeDraftCheckoutState: noop, onChangeDraftScope: noop, onChangeDraftTagIds: noop,
    onChangeSort: noop, onClearFilters: noop, onRemoveFilter: noop, onToggleFilters: noop,
    onExpiration: mode => modes.push(mode)
  };
  try {
    await h.render(<SearchHeader {...props} />);
    expect(h.allText().join(' ')).not.toContain('Expiration');
    await h.render(<SearchHeader {...props} filtersExpanded />);
    for (const label of ['Review expiring soon items', 'Review expired items', 'Review all expiration dates']) {
      await h.press(h.byLabel(label));
    }
    expect(modes).toEqual(['soon', 'expired', 'all']);
  } finally { await h.unmount(); }
});
