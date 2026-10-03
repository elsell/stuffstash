import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetCard } from './AssetCard';
import { localization, t } from '../../presentation/localization';

it('renders matched fields as one cataloged caption with locale list formatting', async () => {
  const h = new MobileRenderHarness();
  const labels = [t('search.match.title'), t('search.match.tag')];
  const asset = { id: 'item', title: 'My tent', kindLabel: 'Item', description: '', locationTrailLabel: '', parentLocationTrail: [], updatedAtLabel: '', photoLabel: '', imagePlaceholderLabel: 'Item', searchMatchLabels: labels };
  try {
    await h.render(<AssetCard asset={asset} onPress={() => {}} onParentLocationPress={() => {}} />);
    expect(h.allText()).toContain(t('search.matchedFields', { fields: localization.list(labels) }));
    expect(h.allText()).toContain('My tent');
  } finally { await h.unmount(); }
});
