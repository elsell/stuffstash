import { describe, expect, it } from 'vitest';
import {
  assetDetailBadges,
  assetDetailLocationContext,
  assetDetailMetadataRows,
  assetDetailSectionOrder,
  assetDetailSectionsPresentation,
  visibleAssetDescription
} from './AssetDetailPresentation';

describe('AssetDetailPresentation', () => {
  it('keeps the core asset workspace out of browse-card chrome', () => {
    expect(assetDetailSectionsPresentation({
      canContainAssets: false,
      hasPhotoStatus: false,
      hasWorkspaceStatus: false,
      photoUploadCount: 0
    })).toEqual([
      { role: 'identity', chrome: 'page' },
      { role: 'metadata', chrome: 'metadata_context' },
      { role: 'maintenance_actions', chrome: 'utility_toolbar' }
    ]);

    expect(assetDetailSectionsPresentation({
      canContainAssets: false,
      hasPhotoStatus: true,
      hasWorkspaceStatus: false,
      photoUploadCount: 1
    })).toEqual([
      { role: 'identity', chrome: 'page' },
      { role: 'status', chrome: 'status_panel' },
      { role: 'metadata', chrome: 'metadata_context' },
      { role: 'maintenance_actions', chrome: 'utility_toolbar' }
    ]);
  });

  it('puts contained assets before generic maintenance for container workspaces', () => {
    expect(assetDetailSectionOrder({
      canContainAssets: true,
      hasPhotoStatus: false,
      hasWorkspaceStatus: false,
      photoUploadCount: 0
    })).toEqual([
      { role: 'identity', chrome: 'page' },
      { role: 'metadata', chrome: 'metadata_context' },
      { role: 'contained_assets', chrome: 'contained_workspace' },
      { role: 'maintenance_actions', chrome: 'utility_toolbar' }
    ]);
  });

  it('keeps the legacy section presentation export aligned with section order', () => {
    expect(assetDetailSectionsPresentation).toBe(assetDetailSectionOrder);
  });

  it('builds detail badges for the page identity section', () => {
    expect(assetDetailBadges({
      kindLabel: 'Container',
      customTypeLabel: 'Documents'
    })).toEqual([
      { label: 'Container', kind: 'kind' },
      { label: 'Documents', kind: 'type' }
    ]);

    expect(assetDetailBadges({
      kindLabel: 'Item',
      customTypeLabel: undefined
    })).toEqual([
      { label: 'Item', kind: 'kind' }
    ]);
  });

  it('builds stable metadata rows for the detail page section', () => {
    expect(assetDetailMetadataRows({
      lifecycleLabel: 'Active',
      updatedAtLabel: 'Updated today'
    })).toEqual([
      { label: 'Status', value: 'Active' },
      { label: 'Updated', value: 'Updated today' }
    ]);
  });

  it('promotes the location path as spatial context instead of generic metadata', () => {
    expect(assetDetailLocationContext({
      locationTrailLabel: 'Office / Filing cabinet'
    })).toEqual({
      label: 'Location',
      value: 'Office / Filing cabinet'
    });
  });

  it('collapses empty descriptions instead of rendering placeholder card copy', () => {
    expect(visibleAssetDescription({ description: '  Important documents.  ' })).toBe('Important documents.');
    expect(visibleAssetDescription({ description: '   ' })).toBeUndefined();
  });
});
