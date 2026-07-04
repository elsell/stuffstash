import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { SelectedPhoto } from '$lib/domain/inventory';
import AddAssetPhotosSection, { type AddAssetPhotosSectionProps } from './AddAssetPhotosSection.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AddAssetPhotosSection', () => {
  it('renders accessible photo actions and hidden upload inputs', () => {
    component = mount(AddAssetPhotosSection, {
      target: document.body,
      props: sectionProps()
    });

    expect(document.body.querySelector('[aria-label="Photo actions"]')?.textContent).toContain('No photos');
    expect(document.body.querySelector('fieldset')?.getAttribute('aria-describedby')).toBe('photo-help photo-status');
    expect(document.body.querySelector('[aria-label="Photo actions"]')?.getAttribute('aria-describedby')).toBe('photo-help photo-status');
    expect(button('Upload').getAttribute('aria-describedby')).toBe('photo-help photo-status');
    expect(button('Camera').getAttribute('aria-describedby')).toBe('photo-help photo-status');
    expect(document.body.querySelector('#photo-status')?.getAttribute('aria-live')).toBe('polite');
    expect(document.body.querySelector<HTMLInputElement>('#asset-photos')?.getAttribute('accept')).toBe('image/jpeg,image/png,image/webp');
    expect(document.body.querySelector<HTMLInputElement>('#asset-photos')?.getAttribute('aria-label')).toBe('Upload photos');
    expect(document.body.querySelector<HTMLInputElement>('#asset-camera')?.getAttribute('capture')).toBe('environment');
    expect(document.body.querySelector<HTMLInputElement>('#asset-camera')?.getAttribute('aria-label')).toBe('Take photo');
    expect(document.body.textContent).toContain('Optional JPEG, PNG, or WebP up to 1 KB.');
  });

  it('derives upload affordances from the runtime media policy', () => {
    component = mount(AddAssetPhotosSection, {
      target: document.body,
      props: sectionProps({
        mediaPolicy: { supportedContentTypes: ['image/png'], maxBytes: 2048 }
      })
    });

    expect(document.body.querySelector<HTMLInputElement>('#asset-photos')?.getAttribute('accept')).toBe('image/png');
    expect(document.body.querySelector<HTMLInputElement>('#asset-camera')?.getAttribute('accept')).toBe('image/png');
    expect(document.body.textContent).toContain('Optional PNG up to 2 KB.');
  });

  it('renders selected photo previews, removal controls, and validation errors', () => {
    const removedIds: string[] = [];
    component = mount(AddAssetPhotosSection, {
      target: document.body,
      props: sectionProps({
        summary: '1 photo',
        photos: [selectedPhoto('photo-one', 'front.jpg')],
        error: 'back.gif is not a supported image type.',
        onRemove: (id) => {
          removedIds.push(id);
        }
      })
    });

    expect(document.body.querySelector('img[alt="front.jpg"]')).not.toBeNull();
    expect(document.body.querySelector('[role="list"][aria-label="Selected photos"]')?.textContent).toContain('front.jpg');
    expect(document.body.querySelector('[role="listitem"]')?.textContent).toContain('front.jpg');
    expect(document.body.querySelector('fieldset')?.getAttribute('aria-describedby')).toBe('photo-help photo-status photo-error');
    expect(button('Upload').getAttribute('aria-describedby')).toBe('photo-help photo-status photo-error');
    expect(button('Camera').getAttribute('aria-describedby')).toBe('photo-help photo-status photo-error');
    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('back.gif is not a supported image type.');

    button('Remove front.jpg').click();

    expect(removedIds).toEqual(['photo-one']);
  });
});

function sectionProps(overrides: Partial<AddAssetPhotosSectionProps> = {}): AddAssetPhotosSectionProps {
  return {
    photos: [],
    summary: 'No photos',
    mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
    inputKey: 0,
    error: '',
    onFiles: () => {},
    onRemove: () => {},
    ...overrides
  };
}

function selectedPhoto(id: string, name: string): SelectedPhoto {
  return {
    id,
    name,
    sizeBytes: 1200,
    contentType: 'image/jpeg',
    previewUrl: `blob:${id}`,
    file: new File(['photo'], name, { type: 'image/jpeg' })
  };
}

function button(name: string): HTMLButtonElement {
  const target = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
    (candidate) => candidate.getAttribute('aria-label') === name || candidate.textContent?.includes(name)
  );
  if (!target) {
    throw new Error(`Missing button ${name}`);
  }
  return target;
}
