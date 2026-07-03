import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AssetDetail from './AssetDetail.svelte';
import type {
  AssetAttachment,
  AssetViewModel,
  CustomFieldDefinition,
  MediaUploadPolicy,
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
    expect(pickerText).toContain('Selected Garage shelf');
    expect(pickerText).toContain('Garage shelf');
    expect(pickerText).toContain('Search 3 available locations and containers.');
    expect(pickerText).not.toContain('Pantry bin');
    expect(pickerText).not.toContain('Hall closet');

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

    expect(link('Edit').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/edit');
    expect(link('Move').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/move');
    expect(link('Archive').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/archive');
    expect(link('Delete').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/delete');
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

  it('disables attachment upload when the viewer cannot edit assets', () => {
    mountAssetDetail({ canEdit: false });

    expect(buttons('Upload')[0]?.disabled).toBe(true);
    expect(fileInput().disabled).toBe(true);
  });
});

function mountAssetDetail(
  props: Partial<{
    asset: AssetViewModel;
    action: 'edit' | 'move' | 'archive' | 'restore' | 'delete' | null;
    canEdit: boolean;
    parentTargets: AssetViewModel[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    attachments: AssetAttachment[];
    mediaPolicy: MediaUploadPolicy;
    onBack: () => void;
    onActionOpen: (action: 'edit' | 'move' | 'archive' | 'restore' | 'delete') => void;
    onActionClose: () => void;
    onSave: (draft: UpdateAssetDraft) => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
    onUploadAttachment: (attachment: SelectedAttachment) => Promise<void>;
    onArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
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
      mediaPolicy: {
        supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
        maxBytes: 5 * 1024 * 1024
      },
      onBack: () => {},
      onActionOpen: () => {},
      onActionClose: () => {},
      onSave: async () => {},
      onArchive: async () => {},
      onRestore: async () => {},
      onDelete: async () => {},
      onUploadAttachment: async () => {},
      onArchiveAttachment: async () => {},
      onDeleteAttachment: async () => {},
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

function parentTarget(id: string, title: string, containmentTrail: string): AssetViewModel {
  return {
    ...asset(),
    id,
    title,
    kind: 'container',
    parentAssetId: null,
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
