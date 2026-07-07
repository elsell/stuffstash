import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
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
  it('summarizes each plan section with table headings and compact counts', () => {
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
    expect(document.body.textContent).toContain('1 record');
    expect(document.body.textContent).toContain('1-2 of 2+');
    expect(document.body.textContent).toContain('Partial list');
    expect(document.body.textContent).toContain('Photos/files');
    expect(document.body.querySelector<HTMLTableElement>('table[aria-label="Assets plan preview"]')?.tagName).toBe('TABLE');
    expect(Array.from(document.body.querySelectorAll('th')).map((node) => node.textContent)).toEqual(
      expect.arrayContaining(['Field', 'Key', 'Type', 'Location', 'Kind', 'Context', 'Asset', 'File', 'Size'])
    );
    expect(document.body.textContent).not.toContain('shown');
    expect(document.body.textContent).not.toContain('Sample');
  });

  it('pages large plan sections without rendering an unbounded wall', async () => {
    component = mount(ImportPreviewSamples, {
      target: document.body,
      props: {
        preview: {
          fields: [],
          locations: [],
          assets: Array.from({ length: 10 }, (_, index) => ({
            title: `Imported asset ${index + 1}`,
            kind: index % 2 === 0 ? 'item' : 'container',
            archived: false
          })),
          attachments: [],
          messages: [],
          fieldsTruncated: false,
          locationsTruncated: false,
          assetsTruncated: false,
          attachmentsTruncated: false,
          messagesTruncated: false
        } satisfies ImportJobPreview
      }
    });

    expect(document.body.textContent).toContain('1-8 of 10');
    expect(document.body.textContent).toContain('Imported asset 8');
    expect(document.body.textContent).not.toContain('Imported asset 9');
    expect(document.body.textContent).toContain('Page 1 of 2');

    const next = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
      (button) => button.getAttribute('aria-label') === 'Next assets plan page'
    );
    expect(next).toBeTruthy();
    next?.click();
    await tick();

    expect(document.body.textContent).toContain('9-10 of 10');
    expect(document.body.textContent).toContain('Imported asset 10');
    expect(document.body.textContent).not.toContain('Imported asset 8');
  });
});
