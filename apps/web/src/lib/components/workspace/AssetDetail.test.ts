import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AssetDetail from './AssetDetail.svelte';
import type {
  AssetAttachment,
  AssetViewModel,
  CustomFieldDefinition,
  MediaUploadPolicy,
  ParentTargetViewModel,
  SelectedAttachment,
  UpdateAssetDraft
} from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetDetail', () => {
  it('scrolls route-opened action panels into view after focusing them', async () => {
    const scrollIntoView = vi.fn();
    const originalScrollIntoView = HTMLElement.prototype.scrollIntoView;
    Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: scrollIntoView });
    try {
      mountAssetDetail({ action: 'edit' });
      await flush();

      expect(document.activeElement).toBe(requiredElement('.detail-action-panel'));
      expect(scrollIntoView).toHaveBeenCalledWith({ block: 'start', inline: 'nearest' });
    } finally {
      Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: originalScrollIntoView });
    }
  });

  it('does not scroll ordinary action-button opened panels', async () => {
    const scrollIntoView = vi.fn();
    const originalScrollIntoView = HTMLElement.prototype.scrollIntoView;
    Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: scrollIntoView });
    try {
      mountAssetDetail();
      clickFirst('Edit');
      await flush();

      expect(document.body.textContent).toContain('Edit asset');
      expect(scrollIntoView).not.toHaveBeenCalled();
    } finally {
      Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: originalScrollIntoView });
    }
  });

  it('promotes photos to the hero and keeps documents in the files section', () => {
    mountAssetDetail({
      asset: {
        ...asset(),
        description: 'Pain reliever bottle with child-safe cap.',
        photo: { id: 'photo-one', assetId: 'asset-one', url: 'blob:primary-photo', alt: 'Ibuprofen bottle' }
      },
      attachments: [
        attachment('photo-one', 'front.jpg', 'image/jpeg', 'blob:front-thumb'),
        attachment('manual-one', 'manual.pdf', 'application/pdf')
      ]
    });

    const hero = document.body.querySelector<HTMLImageElement>('.asset-hero-photo img');
    expect(hero?.src).toBe('blob:primary-photo');
    expect(hero?.alt).toBe('Ibuprofen bottle');
    expect(document.body.querySelectorAll('.photo-rail button')).toHaveLength(1);
    expect(document.body.textContent).toContain('Files');
    expect(document.body.textContent).toContain('manual.pdf');
    expect(document.body.textContent).not.toContain('front.jpgimage/jpeg');
    expect(document.body.querySelector('.asset-detail-copy')?.textContent).not.toContain('Pain reliever bottle');
    expect(document.body.querySelector('.detail-section')?.textContent).toContain('Pain reliever bottle');
  });

  it('keeps the photo-first detail reading order', () => {
    mountAssetDetail({
      asset: {
        ...asset(),
        photo: { id: 'photo-one', assetId: 'asset-one', url: 'blob:primary-photo', alt: 'Ibuprofen bottle' }
      },
      attachments: [
        attachment('photo-one', 'front.jpg', 'image/jpeg', 'blob:front-thumb'),
        attachment('manual-one', 'manual.pdf', 'application/pdf')
      ]
    });

    const hero = requiredElement('.asset-hero-photo');
    const identity = requiredElement('.asset-detail-copy');
    const gallery = requiredElement('.photo-gallery-section');
    const details = requiredElement('.detail-section');
    const files = requiredElement('.attachment-section');

    expect(hero.compareDocumentPosition(identity) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(identity.compareDocumentPosition(gallery) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(gallery.compareDocumentPosition(details) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(details.compareDocumentPosition(files) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(document.body.querySelector<HTMLButtonElement>('button[aria-label="More asset actions"]')).toBeNull();
  });

  it('renders helper-backed detail fallbacks and edit denial copy', () => {
    mountAssetDetail({ canEdit: false });

    expect(document.body.querySelector('.detail-section')?.textContent).toContain('No description.');
    expect(document.body.textContent).toContain('Edit actions require asset edit access.');
  });

  it('searches by detail tag without opening another action', async () => {
    const searchedTags: string[] = [];
    mountAssetDetail({
      asset: {
        ...asset(),
        tags: [{ id: 'tag-medicine', key: 'medicine', displayName: 'Medicine', color: '#2F80ED' }]
      },
      onTagSearch: async (tag) => {
        searchedTags.push(tag.displayName);
      }
    });

    buttonWithLabel('Search for tag Medicine').click();
    await flush();

    expect(searchedTags).toEqual(['Medicine']);
    expect(document.body.querySelector('.detail-action-panel')).toBeNull();
  });

  it('ignores a photo owned by a different asset', () => {
    mountAssetDetail({
      asset: {
        ...asset(),
        photo: { id: 'photo-other', assetId: 'asset-other', url: 'blob:wrong-photo', alt: 'Wrong asset photo' }
      }
    });

    expect(document.body.querySelector('.asset-hero-photo img')).toBeNull();
    expect(document.body.textContent).toContain('No photos yet.');
  });

  it('ignores image attachments owned by a different asset', () => {
    mountAssetDetail({
      attachments: [
        {
          ...attachment('photo-other', 'wrong.jpg', 'image/jpeg', 'blob:wrong-thumb'),
          assetId: 'asset-other'
        }
      ]
    });

    expect(document.body.querySelector('.asset-hero-photo img')).toBeNull();
    expect(document.body.textContent).toContain('No photos yet.');
    expect(document.body.textContent).not.toContain('wrong.jpg');
  });

  it('seeds route-opened edit panels from the current asset', async () => {
    mountAssetDetail({
      action: 'edit',
      asset: {
        ...asset(),
        id: 'asset-two',
        title: 'Current asset',
        description: 'Fresh detail state',
        customFields: { 'expiration-date': '2028-02-02' }
      },
      customFieldDefinitions: [
        {
          id: 'field-expiration',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          scope: 'inventory',
          key: 'expiration-date',
          displayName: 'Expiration date',
          type: 'date',
          enumOptions: [],
          applicability: 'custom_asset_types',
          customAssetTypeIds: ['type-medicine'],
          lifecycleState: 'active'
        }
      ]
    });
    await flush();

    expect((requiredElement('#edit-asset-title') as HTMLInputElement).value).toBe('Current asset');
    expect((requiredElement('#edit-custom-field-expiration-date') as HTMLInputElement).value).toBe('2028-02-02');
    expect(document.activeElement).toBe(requiredElement('.detail-action-panel'));
    expect(
      Boolean(
        requiredElement('.detail-action-panel').compareDocumentPosition(requiredElement('#asset-description-title')) &
          Node.DOCUMENT_POSITION_FOLLOWING
      )
    ).toBe(true);
  });

  it('preserves custom field values when moving an asset', async () => {
    let savedDraft: UpdateAssetDraft | null = null;
    mountAssetDetail({
      customFieldDefinitions: [
        {
          id: 'field-expiration',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          scope: 'inventory',
          key: 'expiration-date',
          displayName: 'Expiration date',
          type: 'date',
          enumOptions: [],
          applicability: 'custom_asset_types',
          customAssetTypeIds: ['type-medicine'],
          lifecycleState: 'active'
        }
      ],
      parentTargets: [parentTarget('parent-one', 'Hall closet', '')],
      onSave: async (draft) => {
        savedDraft = draft;
      }
    });

    clickFirst('Move');
    await flush();
    clickFirst('Inventory root');
    clickLast('Move');
    await flush();

    expect(savedDraft).toMatchObject({
      parentAssetId: null,
      customFields: { 'expiration-date': '2027-01-01' }
    });
  });

  it('filters move targets with grouped picker semantics', async () => {
    let savedDraft: UpdateAssetDraft | null = null;
    mountAssetDetail({
      parentTargets: [
        parentTarget('garage-shelf', 'Garage shelf', 'Garage'),
        parentTarget('hall-closet', 'Hall closet', 'Hall'),
        parentTarget('pantry-bin', 'Pantry bin', 'Kitchen')
      ],
      onSave: async (draft) => {
        savedDraft = draft;
      }
    });

    clickFirst('Move');
    await flush();

    const fieldset = Array.from(document.body.querySelectorAll('fieldset')).find((candidate) =>
      candidate.textContent?.includes('Parent')
    );
    expect(fieldset).toBeTruthy();

    setInputValue(requiredElement('#move-parent-search') as HTMLInputElement, 'closet');
    await flush();

    expect(document.body.textContent).toContain('Hall closet');
    expect(document.body.textContent).not.toContain('Garage shelf');
    clickFirst('Hall closet');
    await flush();

    expect(buttons('Hall closet')[0]?.getAttribute('aria-pressed')).toBe('true');
    clickLast('Move');
    await flush();

    expect(savedDraft).toMatchObject({ parentAssetId: 'hall-closet' });
  });

  it('opens route-linked move panels with a search-first parent picker', async () => {
    mountAssetDetail({
      action: 'move',
      asset: {
        ...asset(),
        parentAssetId: 'garage-shelf'
      },
      parentTargets: [
        parentTarget('garage-shelf', 'Garage shelf', 'Garage'),
        parentTarget('hall-closet', 'Hall closet', 'Hall'),
        parentTarget('pantry-bin', 'Pantry bin', 'Kitchen')
      ]
    });
    await flush();

    expect((requiredElement('#move-parent-search') as HTMLInputElement).value).toBe('');
    const pickerText = parentPickerText();
    expect(pickerText).toContain('Current destination');
    expect(pickerText).toContain('Garage shelf');
    expect(pickerText).toContain('Container / Garage');
    expect(pickerText).toContain('3 possible destinations');
    expect(pickerText).toContain('Suggested destinations');
    expect(pickerText).toContain('Showing 2 suggested destinations.');
    expect(pickerText).toContain('Pantry bin');
    expect(pickerText).toContain('Hall closet');

    setInputValue(requiredElement('#move-parent-search') as HTMLInputElement, 'closet');
    await flush();

    expect(document.body.textContent).toContain('Hall closet');
    expect(parentPickerText()).not.toContain('Pantry bin');
  });

  it('opens route-linked archive and restore confirmation panels', async () => {
    const calls: string[] = [];
    mountAssetDetail({
      action: 'archive',
      onArchive: async () => {
        calls.push('archive');
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Archive asset');
    expect(document.body.textContent).toContain('Move Ibuprofen out of active browsing?');
    clickLast('Archive');
    await flush();
    expect(calls).toEqual(['archive']);

    unmount(component!);
    component = null;
    document.body.innerHTML = '';

    mountAssetDetail({
      action: 'restore',
      asset: {
        ...asset(),
        lifecycleState: 'archived'
      },
      onRestore: async () => {
        calls.push('restore');
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Restore asset');
    expect(document.body.textContent).toContain('Return Ibuprofen to active browsing?');
    clickLast('Restore');
    await flush();
    expect(calls).toEqual(['archive', 'restore']);
  });

  it('routes archive and restore buttons through durable action state', async () => {
    const actions: string[] = [];
    mountAssetDetail({
      onActionOpen: (action) => {
        actions.push(action);
      }
    });

    clickFirst('Archive');
    await flush();

    expect(actions).toEqual(['archive']);
    expect(document.body.textContent).toContain('Archive asset');

    unmount(component!);
    component = null;
    document.body.innerHTML = '';

    mountAssetDetail({
      asset: {
        ...asset(),
        lifecycleState: 'archived'
      },
      onActionOpen: (action) => {
        actions.push(action);
      }
    });

    clickFirst('Restore');
    await flush();

    expect(actions).toEqual(['archive', 'restore']);
    expect(document.body.textContent).toContain('Restore asset');
  });

  it('exposes canonical hrefs for durable asset actions', () => {
    mountAssetDetail();

    expect(link('Back').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one');
    expect(link('Edit').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/edit');
    expect(link('Move').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/move');
    expect(link('Archive').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/archive');
    expect(link('Delete').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/delete');
  });

  it('preserves modified clicks on the asset detail back link', () => {
    let backOpened = false;
    mountAssetDetail({
      onBack: () => {
        backOpened = true;
      }
    });

    let componentPreventedModifiedClick = false;
    const target = link('Back');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(backOpened).toBe(false);
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('uses the location edit route for location asset edit actions', () => {
    mountAssetDetail({
      asset: {
        ...asset(),
        kind: 'location'
      }
    });

    expect(link('Edit').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/locations/asset-one/edit');
  });

  it('exposes canonical hrefs for attachment delete confirmations', async () => {
    let openedAttachmentId = '';
    let closed = false;
    mountAssetDetail({
      attachments: [attachment('manual-one', 'manual.pdf', 'application/pdf')],
      onAttachmentDeleteOpen: (attachmentId) => {
        openedAttachmentId = attachmentId;
      },
      onAttachmentDeleteClose: () => {
        closed = true;
      }
    });

    const deleteLink = link('Delete');
    expect(deleteLink.getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/manual-one/delete'
    );
    deleteLink.click();
    await flush();

    expect(openedAttachmentId).toBe('manual-one');
    expect(document.body.textContent).toContain('Delete attachment');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one');

    link('Cancel').click();
    await flush();

    expect(closed).toBe(true);
    expect(document.body.textContent).not.toContain('Delete attachment');
  });

  it('opens attachment delete confirmation from route state', async () => {
    mountAssetDetail({
      attachmentId: 'manual-one',
      attachmentAction: 'delete',
      attachments: [attachment('manual-one', 'manual.pdf', 'application/pdf')]
    });
    await flush();

    expect(document.body.textContent).toContain('Delete attachment');
    expect(document.body.textContent).toContain('Delete manual.pdf permanently?');
  });

  it('exposes canonical hrefs for asset action cancel controls', async () => {
    let actionClosed = false;
    mountAssetDetail({
      action: 'edit',
      onActionClose: () => {
        actionClosed = true;
      }
    });
    await flush();

    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one');
    link('Cancel').click();
    await flush();

    expect(actionClosed).toBe(true);
    expect(document.body.textContent).not.toContain('Edit asset');
  });

  it('preserves modified clicks on asset action cancel links', async () => {
    let actionClosed = false;
    mountAssetDetail({
      action: 'move',
      onActionClose: () => {
        actionClosed = true;
      }
    });
    await flush();

    let componentPreventedModifiedClick = false;
    const target = link('Cancel');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(actionClosed).toBe(false);
    expect(componentPreventedModifiedClick).toBe(false);
    expect(document.body.textContent).toContain('Move asset');
  });

  it('uses the focused location route for location action cancel links', async () => {
    mountAssetDetail({
      action: 'edit',
      asset: {
        ...asset(),
        kind: 'location'
      }
    });
    await flush();

    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/locations/asset-one');
  });

  it('uses pressed states for enum custom fields in the edit panel', async () => {
    mountAssetDetail({
      asset: {
        ...asset(),
        customFields: { condition: 'open' }
      },
      customFieldDefinitions: [
        {
          id: 'field-condition',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          scope: 'inventory',
          key: 'condition',
          displayName: 'Condition',
          type: 'enum',
          enumOptions: ['new', 'open', 'closed'],
          applicability: 'all_assets',
          customAssetTypeIds: [],
          lifecycleState: 'active'
        }
      ]
    });

    clickFirst('Edit');
    await flush();

    expect(buttons('open')[0]?.getAttribute('aria-pressed')).toBe('true');
    clickFirst('closed');
    await flush();
    expect(buttons('closed')[0]?.getAttribute('aria-pressed')).toBe('true');
  });

  it('uploads a selected attachment through the detail callback', async () => {
    const uploads: SelectedAttachment[] = [];
    mountAssetDetail({
      onUploadAttachment: async (attachment) => {
        uploads.push(attachment);
      }
    });

    chooseAttachment(new File(['manual'], 'manual.pdf', { type: 'application/pdf' }), 'Choose file');
    await flush();

    expect(uploads[0]).toMatchObject({
      name: 'manual.pdf',
      sizeBytes: 6,
      contentType: 'application/pdf'
    });
    expect(uploads[0]?.file.name).toBe('manual.pdf');
  });

  it('blocks attachments that exceed the media policy size', async () => {
    let uploadCount = 0;
    mountAssetDetail({
      mediaPolicy: {
        supportedContentTypes: ['image/jpeg'],
        maxBytes: 4
      },
      onUploadAttachment: async () => {
        uploadCount += 1;
      }
    });

    chooseAttachment(new File(['larger'], 'photo.jpg', { type: 'image/jpeg' }), 'Choose photo');
    await flush();

    expect(uploadCount).toBe(0);
    expect(document.body.textContent).toContain('Attachment must be 4 B or smaller.');
  });

  it('explains disabled photo upload states', () => {
    const cases: Array<{
      name: string;
      props: Parameters<typeof mountAssetDetail>[0];
      reason: string;
    }> = [
      {
        name: 'missing edit access',
        props: { canEdit: false },
        reason: 'Photo upload requires asset edit access.'
      },
      {
        name: 'inactive asset',
        props: { asset: { ...asset(), lifecycleState: 'archived' } },
        reason: 'Restore this asset before adding photos.'
      },
      {
        name: 'save in progress',
        props: { saving: true },
        reason: 'Finish the current change before adding photos.'
      },
      {
        name: 'no supported image types',
        props: { mediaPolicy: { supportedContentTypes: ['application/pdf'], maxBytes: 1024 } },
        reason: 'Photo uploads are unavailable for this media policy.'
      }
    ];

    for (const testCase of cases) {
      if (component) {
        unmount(component);
        component = null;
      }
      document.body.innerHTML = '';
      mountAssetDetail(testCase.props);

      expect(buttons('Add photo'), testCase.name).toHaveLength(2);
      for (const button of buttons('Add photo')) {
        expect(button.disabled, testCase.name).toBe(true);
        expect(button.getAttribute('aria-describedby'), testCase.name).toBe('asset-photo-upload-disabled');
        for (const id of (button.getAttribute('aria-describedby') ?? '').split(' ')) {
          expect(document.getElementById(id), `${testCase.name} ${id}`).not.toBeNull();
        }
      }
      expect(document.body.textContent, testCase.name).toContain(testCase.reason);
      expect(fileInput('Choose photo').disabled, testCase.name).toBe(true);
    }
  });

  it('keeps the photo input image-only even when other attachment types are supported', async () => {
    let uploadCount = 0;
    mountAssetDetail({
      mediaPolicy: { supportedContentTypes: ['image/jpeg', 'application/pdf'], maxBytes: 1024 },
      onUploadAttachment: async () => {
        uploadCount += 1;
      }
    });

    chooseAttachment(new File(['pdf'], 'receipt.pdf', { type: 'application/pdf' }), 'Choose photo');
    await flush();

    expect(uploadCount).toBe(0);
    expect(document.body.textContent).toContain('Unsupported image type.');
  });

  it('shows a helper-backed error for unsupported file uploads', async () => {
    let uploadCount = 0;
    mountAssetDetail({
      mediaPolicy: { supportedContentTypes: ['image/jpeg'], maxBytes: 1024 },
      onUploadAttachment: async () => {
        uploadCount += 1;
      }
    });

    chooseAttachment(new File(['pdf'], 'receipt.pdf', { type: 'application/pdf' }), 'Choose file');
    await flush();

    expect(uploadCount).toBe(0);
    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('Unsupported file type.');
  });
});

function mountAssetDetail(
  props: Partial<{
    asset: AssetViewModel;
    action: 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'checkout' | 'return' | null;
    attachmentId: string | null;
    attachmentAction: 'delete' | null;
    canEdit: boolean;
    parentTargets: ParentTargetViewModel[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    attachments: AssetAttachment[];
    mediaPolicy: MediaUploadPolicy;
    backHref: string;
    onBack: () => void;
    onActionOpen: (action: 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'checkout' | 'return') => void;
    onActionClose: () => void;
    onSave: (draft: UpdateAssetDraft) => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
    onCheckout: (details: string) => Promise<void>;
    onReturn: (details: string) => Promise<void>;
    onUploadAttachment: (attachment: SelectedAttachment) => Promise<void>;
    onArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onAttachmentDeleteOpen: (attachmentId: string) => void;
    onAttachmentDeleteClose: () => void;
    onDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
    onTagSearch: (tag: NonNullable<AssetViewModel['tags']>[number]) => Promise<void>;
  }> = {}
): void {
  component = mount(AssetDetail, {
    target: document.body,
    props: {
      asset: asset(),
      canEdit: true,
      parentTargets: [],
      customFieldDefinitions: [],
	      saving: false,
	      attachments: [],
	      checkoutHistory: [],
      mediaPolicy: {
        supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
        maxBytes: 5 * 1024 * 1024
      },
      backHref: '/tenants/tenant-one/inventories/inventory-one',
      onBack: () => {},
      onActionOpen: () => {},
      onActionClose: () => {},
      onSave: async () => {},
      onArchive: async () => {},
	      onRestore: async () => {},
	      onDelete: async () => {},
	      onCheckout: async () => {},
	      onReturn: async () => {},
	      onUploadAttachment: async () => {},
      onArchiveAttachment: async () => {},
      onAttachmentDeleteOpen: () => {},
      onAttachmentDeleteClose: () => {},
      onDeleteAttachment: async () => {},
      onTagSearch: async () => {},
      ...props
    }
  });
}

function asset(): AssetViewModel {
  return {
    id: 'asset-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title: 'Ibuprofen',
    description: '',
    parentAssetId: 'parent-one',
    lifecycleState: 'active',
    customAssetTypeId: 'type-medicine',
    customAssetTypeLabel: 'Medicine',
    customFields: { 'expiration-date': '2027-01-01' },
    containmentTrail: 'Hall closet'
  };
}

function attachment(
  id: string,
  fileName: string,
  contentType: AssetAttachment['contentType'],
  thumbnailUrl?: string
): AssetAttachment {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    assetId: 'asset-one',
    fileName,
    contentType,
    sizeBytes: 12,
    lifecycleState: 'active',
    thumbnailUrl
  };
}

function parentTarget(id: string, title: string, containmentTrail: string): ParentTargetViewModel {
  return {
    ...asset(),
    id,
    title,
    kind: 'container',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}

function chooseAttachment(file: File, label: string): void {
  const input = fileInput(label);
  Object.defineProperty(input, 'files', { value: [file], configurable: true, writable: true });
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

function setInputValue(input: HTMLInputElement, value: string): void {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
  setter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
}

function fileInput(label = 'Choose file'): HTMLInputElement {
  const input = Array.from(document.body.querySelectorAll<HTMLInputElement>('input[type="file"]')).find(
    (candidate) => candidate.getAttribute('aria-label') === label
  );
  if (!input) throw new Error('Missing attachment file input');
  return input;
}

function clickFirst(text: string): void {
  const control = controls(text)[0];
  if (!control) throw new Error(`Missing control ${text}`);
  control.click();
}

function clickLast(text: string): void {
  const matching = controls(text);
  const control = matching[matching.length - 1];
  if (!control) throw new Error(`Missing control ${text}`);
  control.click();
}

function buttons(text: string): HTMLButtonElement[] {
  return Array.from(document.body.querySelectorAll('button')).filter((candidate) => candidate.textContent?.includes(text));
}

function controls(text: string): HTMLElement[] {
  return Array.from(document.body.querySelectorAll<HTMLElement>('button, a')).filter((candidate) =>
    candidate.textContent?.includes(text)
  );
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!target) throw new Error(`Missing link ${text}`);
  return target;
}

function buttonWithLabel(label: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
    (candidate) => candidate.getAttribute('aria-label') === label
  );
  if (!button) throw new Error(`Missing button ${label}`);
  return button;
}

function requiredElement(selector: string): Element {
  const element = document.body.querySelector(selector);
  if (!element) throw new Error(`Missing element ${selector}`);
  return element;
}

function parentPickerText(): string {
  const fieldset = Array.from(document.body.querySelectorAll('fieldset')).find((candidate) =>
    candidate.textContent?.includes('Parent')
  );
  if (!fieldset) {
    throw new Error('Missing parent picker fieldset');
  }
  return fieldset.textContent ?? '';
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
