import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { navigationOptions, resetNavigation } from '../../test-support/navigation';
import { t } from '../../presentation/localization';
import { ExpirationFiltersScreen } from '../expiration/ExpirationFiltersScreen';
import { AddDestinationSelectionScreen } from './AddDestinationSelectionScreen';

it.each([
  ['mobile.ExpirationFiltersScreen.chooseType', 'mobile.ExpirationFiltersScreen.typesTitle', 'mobile.ExpirationFiltersScreen.searchTypes'],
  ['mobile.ExpirationFiltersScreen.chooseTags', 'mobile.ExpirationFiltersScreen.tags', 'mobile.ExpirationFiltersScreen.searchTags'],
  ['mobile.ExpirationFiltersScreen.chooseLocation', 'mobile.ExpirationFiltersScreen.locationsTitle', 'mobile.ExpirationFiltersScreen.searchLocations'],
] as const)('localizes the filter title and native search prompt for %s', async (open, title, search) => {
  resetNavigation(); const h = new MobileRenderHarness();
  try {
    await h.render(<ExpirationFiltersScreen initial={{ mode: 'all' }} choices={{ types: [], tags: [], locations: [] }} onApply={() => {}} onCancel={() => {}} />);
    await h.press(h.byLabel(t(open)));
    const options = Object.assign({}, ...navigationOptions());
    expect(options.title).toBe(t(title));
    expect(options.headerSearchBarOptions.placeholder).toBe(t(search));
  } finally { await h.unmount(); resetNavigation(); }
});

it('localizes only the destination root fallback and preserves user text', async () => {
  const h = new MobileRenderHarness();
  const base = { query: '', matches: [], disabled: false, loading: false, failed: false, creating: false, canCreate: false, onQuery() {}, onRetry() {}, onSelect() {}, onCreate() {}, onClose() {} };
  try {
    await h.render(<AddDestinationSelectionScreen {...base} />);
    expect(h.byText(t('mobile.AddDestinationSelectionScreen.current', { value: t('mobile.AddDestinationSelectionScreen.inventoryRoot') }))).toBeDefined();
    await h.render(<AddDestinationSelectionScreen {...base} unresolvedSelection="My garage" />);
    expect(h.byText(t('mobile.AddDestinationSelectionScreen.current', { value: 'My garage' }))).toBeDefined();
  } finally { await h.unmount(); resetNavigation(); }
});
