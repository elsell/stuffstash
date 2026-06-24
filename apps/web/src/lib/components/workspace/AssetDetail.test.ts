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

  it('uploads a selected attachment through the detail callback', async () => {
    const uploads: SelectedAttachment[] = [];
    mountAssetDetail({
      onUploadAttachment: async (attachment) => {
        uploads.push(attachment);
      }
    });

    chooseAttachment(new File(['manual'], 'manual.pdf', { type: 'application/pdf' }));
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

    chooseAttachment(new File(['larger'], 'photo.jpg', { type: 'image/jpeg' }));
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
    canEdit: boolean;
    parentTargets: AssetViewModel[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    attachments: AssetAttachment[];
    mediaPolicy: MediaUploadPolicy;
    onBack: () => void;
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

function chooseAttachment(file: File): void {
  const input = fileInput();
  Object.defineProperty(input, 'files', { value: [file], configurable: true, writable: true });
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

function fileInput(): HTMLInputElement {
  const input = document.body.querySelector<HTMLInputElement>('input[type="file"]');
  if (!input) throw new Error('Missing attachment file input');
  return input;
}

function clickFirst(text: string): void {
  const button = buttons(text)[0];
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

function clickLast(text: string): void {
  const matching = buttons(text);
  const button = matching[matching.length - 1];
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

function buttons(text: string): HTMLButtonElement[] {
  return Array.from(document.body.querySelectorAll('button')).filter((candidate) => candidate.textContent?.includes(text));
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
