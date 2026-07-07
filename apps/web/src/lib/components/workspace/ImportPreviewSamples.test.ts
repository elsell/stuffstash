import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { ImportJobPreview } from '$lib/domain/inventory';
import ImportPreviewSamples from './ImportPreviewSamples.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('ImportPreviewSamples', () => {
  it('summarizes each plan section with a compact count', () => {
    component = mount(ImportPreviewSamples, {
      target: document.body,
      props: {
        preview: {
          fields: [{ key: 'serial_number', displayName: 'Serial number', type: 'text' }],
          locations: [{ title: 'Garage', kind: 'location', archived: false }],
          assets: [
            { title: 'Cordless drill', kind: 'item', archived: false },
            { title: 'Toolbox', kind: 'container', archived: false }
          ],
          attachments: [{ fileName: 'drill.jpg', contentType: 'image/jpeg', sizeBytes: 4096, primary: true }],
          messages: [],
          fieldsTruncated: false,
          locationsTruncated: false,
          assetsTruncated: true,
          attachmentsTruncated: false,
          messagesTruncated: false
        } satisfies ImportJobPreview
      }
    });

    expect(document.body.textContent).toContain('Fields');
    expect(document.body.textContent).toContain('1 field');
    expect(document.body.textContent).toContain('1 location');
    expect(document.body.textContent).toContain('2 assets shown');
    expect(document.body.textContent).toContain('1 photo/file');
  });
});
