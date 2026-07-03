import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import AssetDetailHero, { type DetailPhoto } from './AssetDetailHero.svelte';
import AssetDetailHeroHarness from './AssetDetailHeroHarness.test.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetDetailHero', () => {
  it('owns the hero grid while preserving identity reading order', () => {
    const selected: string[] = [];
    component = mount(AssetDetailHeroHarness, {
      target: document.body,
      props: {
        heroPhoto: photo('photo-one', true),
        photos: [photo('photo-one', true), photo('photo-two', false)],
        onSelectPhoto: (photoId) => selected.push(photoId)
      }
    });

    const hero = requiredElement('.asset-hero-photo');
    const identity = requiredElement('.asset-detail-copy');
    const gallery = requiredElement('.photo-gallery-section');
    expect(hero.compareDocumentPosition(identity) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(identity.compareDocumentPosition(gallery) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(requiredElement<HTMLImageElement>('.asset-hero-photo img').alt).toBe('photo-one alt');
    expect(button('Show photo-one.jpg').getAttribute('aria-pressed')).toBe('true');

    button('Show photo-two.jpg').click();

    expect(selected).toEqual(['photo-two']);
  });

  it('shows the kind fallback and disabled upload affordance when photo upload is unavailable', () => {
    component = mount(AssetDetailHero, {
      target: document.body,
      props: {
        kind: 'container',
        heroPhoto: undefined,
        photos: [],
        canAddPhoto: false,
        uploadError: 'Attachment must be 4 B or smaller.',
        onChoosePhoto: () => {},
        onSelectPhoto: () => {}
      }
    });

    expect(document.body.querySelector('.asset-hero-fallback svg')).toBeTruthy();
    expect(document.body.textContent).toContain('No photos yet.');
    expect(document.body.textContent).toContain('Attachment must be 4 B or smaller.');
    expect(button('Add photo').disabled).toBe(true);
  });
});

function photo(id: string, isPrimary: boolean): DetailPhoto {
  return {
    id,
    url: `blob:${id}`,
    alt: `${id} alt`,
    fileName: `${id}.jpg`,
    isPrimary
  };
}

function button(name: string): HTMLButtonElement {
  const match = Array.from(document.body.querySelectorAll('button')).find(
    (candidate) => candidate.textContent?.includes(name) || candidate.getAttribute('aria-label') === name
  );
  if (!match) {
    throw new Error(`Missing button ${name}`);
  }
  return match;
}

function requiredElement<T extends Element = Element>(selector: string): T {
  const element = document.body.querySelector<T>(selector);
  if (!element) {
    throw new Error(`Missing ${selector}`);
  }
  return element;
}
