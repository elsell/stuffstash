import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import ChoiceGrid from './ChoiceGrid.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('ChoiceGrid', () => {
  it('renders choices as a named pressed-button group', () => {
    const selected: string[] = [];
    component = mount(ChoiceGrid, {
      target: document.body,
      props: {
        label: 'Custom asset type',
        selectedValues: ['tool'],
        options: [
          { value: '', label: 'Base asset' },
          { value: 'tool', label: 'Tool', description: 'Hand and power tools' },
          { value: 'supply', label: 'Supply' }
        ],
        onSelect: (value) => {
          selected.push(value);
        }
      }
    });

    expect(group('Custom asset type')?.classList.contains('parent-picker')).toBe(true);
    expect(group('Custom asset type')?.classList.contains('option-grid')).toBe(true);
    expect(button('Tool')?.getAttribute('aria-pressed')).toBe('true');
    expect(button('Supply')?.getAttribute('aria-pressed')).toBe('false');
    expect(button('Tool')?.textContent).toContain('Hand and power tools');

    button('Supply')?.click();

    expect(selected).toEqual(['supply']);
  });

  it('keeps disabled choices non-actionable and exposes calm empty state copy', () => {
    const selected: string[] = [];
    component = mount(ChoiceGrid, {
      target: document.body,
      props: {
        label: 'Field targets',
        selectedValues: [],
        options: [{ value: 'medicine', label: 'Medicine', disabled: true }],
        onSelect: (value) => {
          selected.push(value);
        }
      }
    });

    expect(button('Medicine')?.disabled).toBe(true);
    button('Medicine')?.click();
    expect(selected).toEqual([]);

    unmount(component);
    component = mount(ChoiceGrid, {
      target: document.body,
      props: {
        label: 'Field targets',
        selectedValues: [],
        options: [],
        emptyMessage: 'No eligible custom asset types for this scope.',
        onSelect: () => {}
      }
    });

    expect(document.body.textContent).toContain('No eligible custom asset types for this scope.');
  });
});

function group(name: string): HTMLElement | null {
  return document.body.querySelector<HTMLElement>(`[role="group"][aria-label="${name}"]`);
}

function button(name: string): HTMLButtonElement | undefined {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(name)
  );
}
