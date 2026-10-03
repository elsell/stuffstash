import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { localization, t } from '../../presentation/localization';
import { AssetBreadcrumbTrail } from './AssetCard';
import { AssetTagSelectionField } from './AssetTagSelectionField';

it('speaks breadcrumb names in order without changing authored path punctuation', async () => {
  const h = new MobileRenderHarness();
  const segments = [{ id: 'garage', title: 'Garage / workshop', isImmediateParent: false }, { id: 'box', title: 'Camp bin', isImmediateParent: true }];
  try {
    await h.render(<AssetBreadcrumbTrail segments={segments} onSegmentPress={() => {}} />);
    expect(h.byLabel(t('mobile.AssetCard.location', { value: localization.list(segments.map(s => s.title)) }))).toBeDefined();
    expect(h.allText()).toContain('Garage / workshop');
  } finally { await h.unmount(); }
});

it('formats selected tag names and retains the empty selection guidance', async () => {
  const h = new MobileRenderHarness();
  const tags = [{ id: 'a', label: 'Camping' }, { id: 'b', label: 'Summer' }];
  try {
    await h.render(<AssetTagSelectionField scope="inventory" tags={tags} selectedIds={['a', 'b']} disabled={false} onChange={() => {}} />);
    expect(h.allText()).toContain(localization.list(['Camping', 'Summer']));
    await h.render(<AssetTagSelectionField scope="inventory" tags={tags} selectedIds={[]} disabled={false} onChange={() => {}} />);
    expect(h.allText()).toContain(t('mobile.AssetTagSelectionField.noneSelected'));
  } finally { await h.unmount(); }
});
