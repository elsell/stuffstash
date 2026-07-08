import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { AssetTag } from '$lib/domain/inventory';
import AssetTagChips from './AssetTagChips.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetTagChips', () => {
  it('renders every assigned tag outside compact contexts', () => {
    component = mount(AssetTagChips, {
      target: document.body,
      props: {
        tags: [tag('tools', 'Tools'), tag('camping', 'Camping'), tag('kids', 'Kids')]
      }
    });

    expect(document.body.textContent).toContain('Tools');
    expect(document.body.textContent).toContain('Camping');
    expect(document.body.textContent).toContain('Kids');
    expect(document.body.querySelector('.tag-chip-overflow')).toBeNull();
  });

  it('keeps compact edit summaries fully visible unless an overflow limit is requested', () => {
    component = mount(AssetTagChips, {
      target: document.body,
      props: {
        compact: true,
        tags: [tag('tools', 'Tools'), tag('camping', 'Camping'), tag('kids', 'Kids'), tag('garage', 'Garage')]
      }
    });

    expect(document.body.textContent).toContain('Tools');
    expect(document.body.textContent).toContain('Camping');
    expect(document.body.textContent).toContain('Kids');
    expect(document.body.textContent).toContain('Garage');
    expect(document.body.querySelector('.tag-chip-overflow')).toBeNull();
  });

  it('shows compact overflow in list and card contexts when requested', () => {
    component = mount(AssetTagChips, {
      target: document.body,
      props: {
        compact: true,
        overflowLimit: 2,
        tags: [tag('tools', 'Tools'), tag('camping', 'Camping'), tag('kids', 'Kids'), tag('garage', 'Garage')]
      }
    });

    expect(document.body.textContent).toContain('Tools');
    expect(document.body.textContent).toContain('Camping');
    expect(document.body.textContent).not.toContain('Kids');
    expect(document.body.textContent).not.toContain('Garage');
    expect(document.body.querySelector('.tag-chip-overflow')?.textContent).toBe('+2');
    expect(document.body.querySelector('.tag-chip-overflow')?.getAttribute('aria-label')).toBe('2 more tags');
  });

  it('uses the tag color as the chip color treatment when present', () => {
    component = mount(AssetTagChips, {
      target: document.body,
      props: {
        tags: [tag('tools', 'Tools')]
      }
    });

    const chip = document.body.querySelector<HTMLElement>('.tag-chip');
    expect(chip?.getAttribute('style')).toContain('--tag-color: #2F80ED');
    expect(chip?.classList.contains('tag-chip-colored')).toBe(true);
  });

  it('can make visible tag chips search actions', () => {
    const selected: AssetTag[] = [];
    component = mount(AssetTagChips, {
      target: document.body,
      props: {
        tags: [tag('tools', 'Tools')],
        onTagSelect: (tag) => selected.push(tag)
      }
    });

    const chip = document.body.querySelector<HTMLButtonElement>('button.tag-chip');
    expect(chip?.getAttribute('aria-label')).toBe('Search for tag Tools');
    chip?.click();

    expect(selected).toEqual([tag('tools', 'Tools')]);
  });

  it('can render row-safe inline search actions with keyboard support', () => {
    const selected: AssetTag[] = [];
    component = mount(AssetTagChips, {
      target: document.body,
      props: {
        tags: [tag('tools', 'Tools')],
        actionMode: 'inline',
        onTagSelect: (tag) => selected.push(tag)
      }
    });

    const chip = document.body.querySelector<HTMLElement>('[role="button"].tag-chip');
    expect(chip?.tagName).toBe('SPAN');
    expect(chip?.getAttribute('tabindex')).toBe('0');
    expect(chip?.getAttribute('aria-label')).toBe('Search for tag Tools');

    chip?.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, cancelable: true, key: 'Enter' }));
    chip?.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, cancelable: true, key: ' ' }));

    expect(selected).toEqual([tag('tools', 'Tools'), tag('tools', 'Tools')]);
  });
});

function tag(key: string, displayName: string): AssetTag {
  return {
    id: `tag-${key}`,
    key,
    displayName,
    color: '#2F80ED'
  };
}
