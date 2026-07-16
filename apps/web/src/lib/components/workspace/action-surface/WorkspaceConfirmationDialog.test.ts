import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import WorkspaceConfirmationDialogHarness from './WorkspaceConfirmationDialogHarness.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

describe('WorkspaceConfirmationDialog', () => {
  it('renders a labelled alert dialog with the safe action before the destructive action', async () => {
    component = mount(WorkspaceConfirmationDialogHarness, { target: document.body });
    await tick();

    const dialog = document.body.querySelector<HTMLElement>('[role="alertdialog"]');
    expect(dialog).not.toBeNull();
    expect(dialog?.classList.contains('motion-reduce:animate-none!')).toBe(true);
    expect(document.body.querySelector('[data-slot="alert-dialog-overlay"]')?.classList.contains('motion-reduce:animate-none!')).toBe(true);
    expect(dialog?.textContent).toContain('Delete asset');
    const buttons = Array.from(dialog?.querySelectorAll('button') ?? []).map((button) => button.textContent?.trim());
    expect(buttons).toEqual(['Cancel', 'Delete']);
    const footer = dialog?.querySelector('[data-slot="alert-dialog-footer"]');
    expect(footer?.classList.contains('flex-col')).toBe(true);
    expect(footer?.classList.contains('flex-col-reverse')).toBe(false);
    expect((document.activeElement as HTMLElement | null)?.textContent).toContain('Cancel');
  });

  it('announces asynchronous confirmation progress without losing context', async () => {
    let deleteCount = 0;
    component = mount(WorkspaceConfirmationDialogHarness, {
      target: document.body,
      props: { busy: true, onDelete: () => { deleteCount += 1; } }
    });
    await tick();

    const dialog = document.body.querySelector<HTMLElement>('[role="alertdialog"]');
    expect(dialog?.getAttribute('aria-busy')).toBe('true');
    expect(dialog?.querySelector('[role="status"]')?.textContent).toContain('Working');
    expect(dialog?.textContent).toContain('Delete asset');
    const action = Array.from(dialog?.querySelectorAll<HTMLButtonElement>('button') ?? [])
      .find((candidate) => candidate.textContent?.includes('Delete'));
    action?.click();
    expect(deleteCount).toBe(0);
    expect(dialog?.querySelector('[data-workspace-confirmation-actions]')?.getAttribute('aria-disabled')).toBe('true');
  });

  it('moves focus to progress when the invoked action makes the dialog busy', async () => {
    component = mount(WorkspaceConfirmationDialogHarness, {
      target: document.body,
      props: { busyOnDelete: true }
    });
    await tick();

    const dialog = document.body.querySelector<HTMLElement>('[role="alertdialog"]');
    const action = Array.from(dialog?.querySelectorAll<HTMLButtonElement>('button') ?? [])
      .find((candidate) => candidate.textContent?.includes('Delete'));
    action?.focus();
    action?.click();
    await tick();
    await tick();

    const progress = dialog?.querySelector<HTMLElement>('[role="status"]');
    expect(progress?.getAttribute('tabindex')).toBe('-1');
    expect(document.activeElement).toBe(progress);
  });
});
