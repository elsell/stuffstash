import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import AssetFilesSection from './AssetFilesSection.svelte';
import type { AssetAttachment, AttachmentContentType } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetFilesSection', () => {
  it('renders active files with archive and route-backed delete actions', () => {
    const opened: string[] = [];
    component = mount(AssetFilesSection, {
      target: document.body,
      props: {
        attachments: [attachment('manual-one', 'manual.pdf', 'application/pdf', 2048)],
        canEdit: true,
        saving: false,
        active: true,
        onChooseFile: () => {},
        onArchiveAttachment: (file) => opened.push(`archive:${file.id}`),
        onOpenAttachmentDelete: (event, file) => {
          event.preventDefault();
          opened.push(`delete:${file.id}`);
        },
        attachmentDeleteHref: (file) => `/files/${file.id}/delete`
      }
    });

    expect(document.body.textContent).toContain('manual.pdf');
    expect(document.body.textContent).toContain('application/pdf / 2 KB');
    expect(link('Delete').getAttribute('href')).toBe('/files/manual-one/delete');

    button('Archive').click();
    link('Delete').click();

    expect(opened).toEqual(['archive:manual-one', 'delete:manual-one']);
  });

  it('keeps upload and row actions disabled for inactive or denied states', () => {
    component = mount(AssetFilesSection, {
      target: document.body,
      props: {
        attachments: [attachment('manual-one', 'manual.pdf')],
        canEdit: false,
        saving: false,
        active: true,
        onChooseFile: () => {},
        onArchiveAttachment: () => {},
        onOpenAttachmentDelete: () => {},
        attachmentDeleteHref: (file) => `/files/${file.id}/delete`
      }
    });

    expect(button('Upload file').disabled).toBe(true);
    expect(button('Archive').disabled).toBe(true);
    expect(link('Delete').getAttribute('aria-disabled')).toBe('true');
    expect(link('Delete').hasAttribute('href')).toBe(false);

    unmount(component);
    component = mount(AssetFilesSection, {
      target: document.body,
      props: {
        attachments: [],
        canEdit: true,
        saving: false,
        active: false,
        onChooseFile: () => {},
        onArchiveAttachment: () => {},
        onOpenAttachmentDelete: () => {},
        attachmentDeleteHref: (file) => `/files/${file.id}/delete`
      }
    });

    expect(button('Upload file').disabled).toBe(true);
    expect(document.body.textContent).toContain('No active files.');
  });

  it('keeps file-operation failures in the Files section and associates them with upload', () => {
    component = mount(AssetFilesSection, {
      target: document.body,
      props: {
        attachments: [],
        canEdit: true,
        saving: false,
        active: true,
        error: { operation: 'upload', message: 'Unable to upload file.' },
        onChooseFile: () => {},
        onArchiveAttachment: () => {},
        onOpenAttachmentDelete: () => {},
        attachmentDeleteHref: (file) => `/files/${file.id}/delete`
      }
    });

    const section = document.body.querySelector('.attachment-section');
    const alert = section?.querySelector('[role="alert"]');
    expect(alert?.textContent).toContain('Unable to upload file.');
    expect(button('Upload file').getAttribute('aria-describedby')).toBe(alert?.id);
  });

  it('associates an archive failure only with the affected row action', () => {
    component = mount(AssetFilesSection, {
      target: document.body,
      props: {
        attachments: [attachment('manual-one', 'manual.pdf')],
        canEdit: true,
        saving: false,
        active: true,
        error: { operation: 'archive', attachmentId: 'manual-one', message: 'Unable to archive manual.pdf.' },
        onChooseFile: () => {},
        onArchiveAttachment: () => {},
        onOpenAttachmentDelete: () => {},
        attachmentDeleteHref: (file) => `/files/${file.id}/delete`
      }
    });

    const alert = document.body.querySelector<HTMLElement>('[role="alert"]');
    expect(alert?.textContent).toContain('Unable to archive manual.pdf.');
    expect(button('Archive').getAttribute('aria-describedby')).toBe(alert?.id);
    expect(button('Upload file').hasAttribute('aria-describedby')).toBe(false);
  });
});

function attachment(id: string, fileName: string, contentType: AttachmentContentType = 'application/pdf', sizeBytes = 512): AssetAttachment {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    assetId: 'asset-one',
    fileName,
    contentType,
    sizeBytes,
    lifecycleState: 'active'
  };
}

function button(name: string): HTMLButtonElement {
  const match = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent?.includes(name));
  if (!match) {
    throw new Error(`Missing button ${name}`);
  }
  return match;
}

function link(name: string): HTMLAnchorElement {
  const match = Array.from(document.body.querySelectorAll('a')).find((candidate) => candidate.textContent?.includes(name));
  if (!match) {
    throw new Error(`Missing link ${name}`);
  }
  return match;
}
