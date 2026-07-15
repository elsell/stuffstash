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
  it('keeps tag choices and color actions at least 44 CSS pixels tall and wide', () => {
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({ tags: [tag('tag-art', 'Art')] })
    });

    expect(button('Art').classList).toContain('min-h-11');
    expect(button('Art').classList).toContain('min-w-11');
    expect(document.body.querySelector('#new-tag-color-picker')?.classList).toContain('size-11');
    expect(button('Add').classList).toContain('min-h-11');
  });

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

  it('stages new tag colors chosen with the color picker', async () => {
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
    colorInput('new-tag-color-picker', '#2e7d32');
    await tick();
    button('Add').click();
    await tick();

    expect(newTags).toEqual([{ displayName: 'Workshop', color: '#2E7D32' }]);
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

  it('sorts available tags alphabetically without changing the input array', async () => {
    const tags = [
      tag('tag-workshop', 'Workshop'),
      tag('tag-attic-title', 'Attic'),
      tag('tag-attic-lower', 'attic'),
      tag('tag-camping', 'Camping')
    ];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({ tags })
    });
    await tick();

    const optionLabels = Array.from(document.querySelectorAll('.tag-options button')).map((element) => element.textContent?.trim());
    expect(optionLabels).toEqual(['Attic', 'attic', 'Camping', 'Workshop']);
    expect(tags.map((item) => item.displayName)).toEqual(['Workshop', 'Attic', 'attic', 'Camping']);
  });

  it('progressively discloses large naturally sorted tag lists', async () => {
    const tags = Array.from({ length: 14 }, (_, index) => tag(`tag-${index}`, `Tag ${14 - index}`));
    component = mount(AssetTagSelector, { target: document.body, props: props({ tags }) });
    await tick();

    expect(document.querySelectorAll('.tag-options button')).toHaveLength(12);
    button('Show all 14 tags').click();
    await tick();
    expect(document.querySelectorAll('.tag-options button')).toHaveLength(14);
    expect(Array.from(document.querySelectorAll('.tag-options button')).map((item) => item.textContent?.trim())).toEqual(
      Array.from({ length: 14 }, (_, index) => `Tag ${index + 1}`)
    );
    button('Show fewer tags').click();
    await tick();
    expect(document.querySelectorAll('.tag-options button')).toHaveLength(12);
  });

  it('keeps a selected option beyond the collapsed limit available for deselection', async () => {
    const tags = Array.from({ length: 14 }, (_, index) => tag(`tag-${index + 1}`, `Tag ${index + 1}`));
    let selectedIds = ['tag-14'];
    component = mount(AssetTagSelector, {
      target: document.body,
      props: props({
        tags,
        selectedIds,
        onSelectedIdsChange: (ids) => {
          selectedIds = ids;
        }
      })
    });
    await tick();

    expect(document.querySelectorAll('.tag-options button')).toHaveLength(13);
    expect(button('Tag 14').getAttribute('aria-pressed')).toBe('true');
    button('Tag 14').click();

    expect(selectedIds).toEqual([]);
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

function colorInput(id: string, value: string): void {
  const element = document.getElementById(id) as HTMLInputElement | null;
  if (!element) {
    throw new Error(`Missing input ${id}.`);
  }
  if (element.type !== 'color') {
    throw new Error(`Expected ${id} to be a native color input.`);
  }
  element.value = value;
  element.dispatchEvent(new Event('change', { bubbles: true }));
}

function button(label: string): HTMLButtonElement {
  const buttons = Array.from(document.querySelectorAll('button'));
  const element = buttons.find((candidate) => candidate.textContent?.trim() === label) as HTMLButtonElement | undefined;
  if (!element) {
    throw new Error(`Missing button ${label}.`);
  }
  return element;
}
