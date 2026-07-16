import { describe, expect, it } from 'vitest';
import {
  assetDetailAvailabilityAction,
  assetDetailBadges,
  assetDetailExceptionMetadataRows,
  assetDetailIdentity,
  assetDetailLocationContext,
  assetDetailMaintenanceActions,
  assetDetailMetadataRows,
  assetDetailPlacement,
  assetDetailSectionOrder,
  assetDetailSectionsPresentation,
  assetDetailUpdatedMetadata,
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

  it('presents the user title before secondary classification metadata', () => {
    expect(assetDetailIdentity({
      title: 'Family tent',
      kindLabel: 'Item',
      customTypeLabel: 'Camping gear'
    })).toEqual({
      title: 'Family tent',
      classificationLabel: 'Item · Camping gear'
    });

    expect(assetDetailIdentity({
      title: 'Garage',
      kindLabel: 'Location',
      customTypeLabel: undefined
    })).toEqual({
      title: 'Garage',
      classificationLabel: 'Location'
    });
  });

  it('keeps normal lifecycle and checkout state out of prominent metadata', () => {
    expect(assetDetailMetadataRows({
      checkoutLabel: 'Available',
      checkoutActorLabel: undefined,
      isActive: true,
      isCheckedOut: false,
      lifecycleLabel: 'Active',
      updatedAtLabel: 'Updated today'
    })).toEqual([]);
  });

  it('shows only exceptional archived and checked-out state near its action', () => {
    const asset = {
      checkoutLabel: 'Checked out Jul 14, 2026',
      checkoutActorLabel: 'Checked out by Alex',
      isActive: false,
      isCheckedOut: true,
      lifecycleLabel: 'Archived'
    };

    expect(assetDetailExceptionMetadataRows(asset)).toEqual([
      { label: 'Lifecycle', value: 'Archived' },
      { label: 'Availability', value: 'Checked out Jul 14, 2026 · Checked out by Alex' }
    ]);
    expect(assetDetailMetadataRows({ ...asset, updatedAtLabel: 'Updated yesterday' }))
      .toEqual(assetDetailExceptionMetadataRows(asset));
  });

  it('keeps updated-at copy concise and visually separate from status', () => {
    expect(assetDetailUpdatedMetadata({ updatedAtLabel: 'Updated today' })).toEqual({
      value: 'Updated today'
    });
  });

  it('builds clickable placement from structured parent crumbs without parsing titles', () => {
    expect(assetDetailPlacement({
      parentLocationTrail: [
        { id: 'garage', title: 'Garage / workshop', isImmediateParent: false },
        { id: 'camp-bin', title: 'Camp bin', isImmediateParent: true }
      ]
    })).toEqual({
      accessibilityLabel: 'Location Garage / workshop, Camp bin',
      crumbs: [
        { id: 'garage', title: 'Garage / workshop', isImmediateParent: false },
        { id: 'camp-bin', title: 'Camp bin', isImmediateParent: true }
      ]
    });
  });

  it('uses calm root-level placement copy instead of exposing inventory internals', () => {
    expect(assetDetailPlacement({ parentLocationTrail: [] })).toEqual({
      accessibilityLabel: 'Location No location',
      crumbs: [],
      fallbackLabel: 'No location'
    });

    expect(assetDetailLocationContext({
      parentLocationTrail: []
    })).toEqual({
      label: 'Location',
      value: 'No location'
    });
  });

  it('renders only the applicable direct availability action', () => {
    expect(assetDetailAvailabilityAction({ canCheckout: true, canReturn: false })).toEqual({
      id: 'check_out',
      label: 'Check out'
    });
    expect(assetDetailAvailabilityAction({ canCheckout: false, canReturn: true })).toEqual({
      id: 'return',
      label: 'Return'
    });
    expect(assetDetailAvailabilityAction({ canCheckout: false, canReturn: false })).toBeUndefined();
  });

  it('prefers Return if inconsistent capabilities claim both actions apply', () => {
    expect(assetDetailAvailabilityAction({ canCheckout: true, canReturn: true })).toEqual({
      id: 'return',
      label: 'Return'
    });
  });

  it('shows only applicable maintenance actions and never exceeds three', () => {
    expect(assetDetailMaintenanceActions({
      canAddPhotos: true,
      canEdit: false,
      canMove: true
    })).toEqual([
      { id: 'move', label: 'Move' },
      { id: 'add_photos', label: 'Add photos' }
    ]);

    expect(assetDetailMaintenanceActions({
      canAddPhotos: true,
      canEdit: true,
      canMove: true
    })).toEqual([
      { id: 'edit', label: 'Edit' },
      { id: 'move', label: 'Move' },
      { id: 'add_photos', label: 'Add photos' }
    ]);
  });

  it('collapses empty descriptions instead of rendering placeholder card copy', () => {
    expect(visibleAssetDescription({ description: '  Important documents.  ' })).toBe('Important documents.');
    expect(visibleAssetDescription({ description: '   ' })).toBeUndefined();
  });
});
