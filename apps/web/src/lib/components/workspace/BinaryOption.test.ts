import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import BinaryOption from './BinaryOption.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('BinaryOption', () => {
  it('renders honest switch semantics with visible state copy', () => {
    component = mount(BinaryOption, {
      target: document.body,
      props: {
        label: 'Images',
        description: 'Import image attachments.',
        checked: true,
        onToggle: () => {}
      }
    });

    const control = switchControl('Images');
    expect(control?.getAttribute('aria-checked')).toBe('true');
    expect(control?.textContent).toContain('Import image attachments.');
    expect(control?.textContent).toContain('On');
    expect(control?.querySelector('.binary-option-track')).not.toBeNull();
  });

  it('does not toggle when disabled', () => {
    let toggled = false;
    component = mount(BinaryOption, {
      target: document.body,
      props: {
        label: 'Private network address',
        checked: false,
        disabled: true,
        onToggle: () => {
          toggled = true;
        }
      }
    });

    switchControl('Private network address')?.click();

    expect(toggled).toBe(false);
  });
});

function switchControl(label: string): HTMLButtonElement | null {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button[role="switch"]')).find((button) =>
    button.textContent?.includes(label)
  ) ?? null;
}
