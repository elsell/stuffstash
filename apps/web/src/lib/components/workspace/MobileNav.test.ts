import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { WorkspaceMode } from '$lib/domain/inventory';
import MobileNav from './MobileNav.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('MobileNav', () => {
  it('routes Places to the durable locations destination', () => {
    let selectedMode: WorkspaceMode | null = null;
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'home',
        canCreateAsset: true,
        onModeChange: (mode) => {
          selectedMode = mode;
        },
        onOpenAdd: () => {}
      }
    });

    buttonContaining('Places').click();

    expect(selectedMode).toBe('locations');
  });

  it('marks focused locations as the current Places section', () => {
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'location',
        canCreateAsset: true,
        onModeChange: () => {},
        onOpenAdd: () => {}
      }
    });

    expect(buttonContaining('Places').getAttribute('aria-current')).toBe('page');
  });
});

function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}
