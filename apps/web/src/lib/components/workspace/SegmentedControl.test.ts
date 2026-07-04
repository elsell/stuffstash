import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import SegmentedControl from './SegmentedControl.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('SegmentedControl', () => {
  it('renders button-backed options as a pressed-button group', () => {
    const selected: string[] = [];

    component = mount(SegmentedControl, {
      target: document.body,
      props: {
        label: 'Lifecycle',
        value: 'active',
        options: [
          { value: 'active', label: 'Active' },
          { value: 'archived', label: 'Archived', description: 'Hidden from default browsing' }
        ],
        onSelect: (value) => selected.push(value)
      }
    });

    const group = namedGroup('Lifecycle');
    expect(group?.querySelectorAll('button[aria-pressed]')).toHaveLength(2);
    expect(button('Active')?.getAttribute('aria-pressed')).toBe('true');
    expect(button('Archived')?.getAttribute('aria-pressed')).toBe('false');
    expect(button('Archived')?.textContent).toContain('Hidden from default browsing');

    button('Archived')?.click();

    expect(selected).toEqual(['archived']);
  });

  it('renders route-backed options as current links with durable hrefs', () => {
    const selected: string[] = [];

    component = mount(SegmentedControl, {
      target: document.body,
      props: {
        label: 'Search mode',
        value: 'assets',
        options: [
          { value: 'assets', label: 'Assets', href: '/search?mode=assets' },
          { value: 'locations', label: 'Locations', href: '/search?mode=locations' }
        ],
        onSelect: (value) => selected.push(value)
      }
    });

    expect(namedGroup('Search mode')?.querySelectorAll('a[data-selected]')).toHaveLength(2);
    expect(link('Assets')?.getAttribute('href')).toBe('/search?mode=assets');
    expect(link('Assets')?.getAttribute('aria-current')).toBe('page');
    expect(link('Locations')?.getAttribute('aria-current')).toBeNull();

    link('Locations')?.click();

    expect(selected).toEqual(['locations']);
  });

  it('keeps disabled route-backed options non-actionable while exposing disabled link semantics', () => {
    const selected: string[] = [];

    component = mount(SegmentedControl, {
      target: document.body,
      props: {
        label: 'Audit scope',
        value: 'tenant',
        options: [
          { value: 'inventory', label: 'Inventory', href: '/settings/activity', disabled: true },
          { value: 'tenant', label: 'Tenant', href: '/settings/activity?auditScope=tenant' }
        ],
        onSelect: (value) => selected.push(value)
      }
    });

    expect(link('Inventory')?.getAttribute('href')).toBeNull();
    expect(link('Inventory')?.getAttribute('aria-disabled')).toBe('true');
    expect(link('Inventory')?.getAttribute('tabindex')).toBe('0');

    link('Inventory')?.click();

    expect(selected).toEqual([]);
  });
});

function namedGroup(label: string): HTMLElement | null {
  return document.body.querySelector<HTMLElement>(`[role="group"][aria-label="${label}"]`);
}

function button(label: string): HTMLButtonElement | undefined {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(label)
  );
}

function link(label: string): HTMLAnchorElement | undefined {
  return Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(label)
  );
}
