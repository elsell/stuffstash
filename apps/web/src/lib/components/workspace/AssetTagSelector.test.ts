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

  it('selects an existing tag even when the color field is invalid', async () => {
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

    input('new-tag-name', 'Workshop');
    input('new-tag-color', 'blue');
    await tick();

    expect(button('Add').disabled).toBe(false);
    button('Add').click();
    await tick();

    expect(selectedIds).toEqual(['tag-workshop']);
    expect(newTags).toEqual([]);
  });

  it('matches existing tags by generated tag key', async () => {
    let selectedIds: string[] = [];
    let newTags: AssetTagDraft[] = [];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        tags: [{ id: 'tag-camp-kitchen', key: 'camp-kitchen', displayName: 'Camp Kitchen' }],
        onSelectedIdsChange: (ids) => {
          selectedIds = ids;
        },
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', 'Camp / Kitchen');
    await tick();
    button('Add').click();
    await tick();

    expect(selectedIds).toEqual(['tag-camp-kitchen']);
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

  it('keeps new tag names within the backend display-name limit', async () => {
    let newTags: AssetTagDraft[] = [];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', 'a'.repeat(81));
    await tick();

    expect(button('Add').disabled).toBe(true);

    input('new-tag-name', 'a'.repeat(80));
    await tick();
    button('Add').click();
    await tick();

    expect(newTags).toEqual([{ displayName: 'a'.repeat(80) }]);
  });

  it('uses UTF-8 byte length for the backend display-name limit', async () => {
    let newTags: AssetTagDraft[] = [];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', `${'é'.repeat(40)}a`);
    await tick();

    expect(button('Add').disabled).toBe(true);
    expect(newTags).toEqual([]);
  });

  it('still selects existing tags when over-limit text resolves to their key', async () => {
    let selectedIds: string[] = [];
    let newTags: AssetTagDraft[] = [];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        tags: [{ id: 'tag-a', key: 'a', displayName: 'A' }],
        onSelectedIdsChange: (ids) => {
          selectedIds = ids;
        },
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', `${'é'.repeat(81)}a`);
    await tick();

    expect(button('Add').disabled).toBe(false);
    button('Add').click();
    await tick();

    expect(selectedIds).toEqual(['tag-a']);
    expect(newTags).toEqual([]);
  });

  it('still clears pending duplicate tags when over-limit text resolves to their key', async () => {
    let newTags: AssetTagDraft[] = [{ displayName: 'A' }];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        newTags,
        onNewTagsChange: (tags) => {
          newTags = tags;
        }
      })
    });

    input('new-tag-name', `${'é'.repeat(81)}a`);
    await tick();

    expect(button('Add').disabled).toBe(false);
    button('Add').click();
    await tick();

    expect(newTags).toEqual([{ displayName: 'A' }]);
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
    key: displayName.toLowerCase().replaceAll(' ', '-'),
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
