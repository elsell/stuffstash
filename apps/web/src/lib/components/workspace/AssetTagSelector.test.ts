import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { AssetTag, AssetTagDraft } from '$lib/domain/inventory';
import AssetTagSelector from './AssetTagSelector.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AssetTagSelector', () => {
  it('selects an existing tag instead of staging a duplicate new tag', async () => {
    let selectedIds: string[] = [];
    let newTags: AssetTagDraft[] = [];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        tags: [tag('tag-workshop', 'Workshop')],
        onSelectedIdsChange: (ids) => {
          selectedIds = ids;
        },
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', ' workshop ');
    await tick();
    button('Add').click();
    await tick();

    expect(selectedIds).toEqual(['tag-workshop']);
    expect(newTags).toEqual([]);
  });

  it('does not stage duplicate pending tag names', async () => {
    let newTags: AssetTagDraft[] = [{ displayName: 'Workshop' }];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        newTags,
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', 'workshop');
    await tick();
    button('Add').click();
    await tick();

    expect(newTags).toEqual([{ displayName: 'Workshop' }]);
  });

  it('requires a valid optional color before staging a new tag', async () => {
    let newTags: AssetTagDraft[] = [];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', 'Workshop');
    input('new-tag-color', 'blue');
    await tick();

    expect(button('Add').disabled).toBe(true);

    input('new-tag-color', '2f80ed');
    await tick();
    button('Add').click();
    await tick();

    expect(newTags).toEqual([{ displayName: 'Workshop', color: '#2F80ED' }]);
  });
});

function props(
  overrides: Partial<{
    tags: AssetTag[];
    selectedIds: string[];
    newTags: AssetTagDraft[];
    onSelectedIdsChange: (ids: string[]) => void;
    onNewTagsChange: (tags: AssetTagDraft[]) => void;
  }> = {}
) {
  return {
    tags: [],
    selectedIds: [],
    newTags: [],
    onSelectedIdsChange: () => {},
    onNewTagsChange: () => {},
    ...overrides
  };
}

function tag(id: string, displayName: string): AssetTag {
  return {
    id,
    key: displayName.toLowerCase(),
    displayName
  };
}

function input(id: string, value: string): void {
  const element = document.getElementById(id) as HTMLInputElement | null;
  if (!element) {
    throw new Error(`Missing input ${id}.`);
  }
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

function button(label: string): HTMLButtonElement {
  const buttons = Array.from(document.querySelectorAll('button'));
  const element = buttons.find((candidate) => candidate.textContent?.trim() === label) as HTMLButtonElement | undefined;
  if (!element) {
    throw new Error(`Missing button ${label}.`);
  }
  return element;
}
