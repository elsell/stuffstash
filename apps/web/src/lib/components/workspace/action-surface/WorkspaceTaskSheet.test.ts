import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import WorkspaceTaskSheetHarness from './WorkspaceTaskSheetHarness.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

describe('WorkspaceTaskSheet', () => {
  it('renders a labelled modal task surface in a portal', async () => {
    component = mount(WorkspaceTaskSheetHarness, { target: document.body });
    await tick();
    await tick();

    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]');
    expect(dialog).not.toBeNull();
    expect(dialog?.getAttribute('aria-modal')).toBe('true');
    expect(dialog?.style.width).toBe('100%');
    expect(dialog?.classList.contains('sm:max-w-xl')).toBe(true);
    expect(dialog?.classList.contains('motion-reduce:animate-none!')).toBe(true);
    expect(dialog?.classList.contains('motion-reduce:transition-none!')).toBe(true);
    expect(dialog?.classList.contains('[&_button]:min-h-11')).toBe(true);
    expect(dialog?.classList.contains('[&_input]:min-h-11')).toBe(true);
    expect(dialog?.classList.contains('[&_select]:min-h-11')).toBe(true);
    expect(dialog?.classList.contains('[&_textarea]:min-h-11')).toBe(true);
    expect(dialog?.querySelector('.workspace-task-sheet-body')?.classList.contains('grid')).toBe(true);
    expect(dialog?.querySelector('.workspace-task-sheet-body')?.classList.contains('gap-6')).toBe(true);
    const header = dialog?.querySelector('[data-slot="sheet-header"]');
    const footer = dialog?.querySelector('[data-slot="sheet-footer"]');
    expect(header?.classList.contains('bg-popover')).toBe(true);
    expect(footer?.classList.contains('bg-popover')).toBe(true);
    expect(footer?.classList.contains('shrink-0')).toBe(true);
    expect(footer?.classList.contains('sm:flex-row')).toBe(true);
    expect(dialog?.textContent).toContain('Edit asset');
    expect(dialog?.textContent).toContain('Change the asset details.');
    expect(dialog?.textContent).toContain('Name');
    expect(dialog?.textContent).toContain('Save');
    expect(dialog?.querySelector('[data-slot="sheet-close"]')?.classList.contains('size-11')).toBe(true);
    expect(dialog?.querySelector('[data-slot="sheet-close"]')?.classList.contains('z-20')).toBe(true);
    expect((document.activeElement as HTMLElement | null)?.id).toBe('task-name');
  });

  it('keeps route-backed close controls at least 44px and announces saving', async () => {
    component = mount(WorkspaceTaskSheetHarness, {
      target: document.body,
      props: { closeHref: '/asset', busy: true }
    });
    await tick();

    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]');
    expect(dialog?.getAttribute('aria-busy')).toBe('true');
    expect(dialog?.querySelector('[role="status"]')?.textContent).toContain('Saving changes');
    expect(dialog?.querySelector('[data-slot="sheet-header"]')?.textContent).toContain('Saving changes');
  });

  it('uses a 44px route-backed close target when the task is dismissible', async () => {
    component = mount(WorkspaceTaskSheetHarness, {
      target: document.body,
      props: { closeHref: '/asset' }
    });
    await tick();

    const close = document.body.querySelector<HTMLElement>('[aria-label="Close"]');
    expect(close?.classList.contains('size-11')).toBe(true);
    expect(close?.classList.contains('z-20')).toBe(true);
  });
});
