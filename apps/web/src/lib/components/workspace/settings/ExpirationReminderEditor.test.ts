import { mount, unmount, tick } from 'svelte';
import { expect, it, vi } from 'vitest';
import ExpirationReminderEditor from './ExpirationReminderEditor.svelte';
const policy = { enabled: true, upcoming: true, expired: true, advanceDays: 30 };
function checkbox(name: string): HTMLInputElement {
  const label = Array.from(document.querySelectorAll('label')).find((element) => element.textContent?.includes(name));
  if (!label) throw new Error(`Missing ${name}`);
  return label.querySelector('input')!;
}
function days(): HTMLInputElement { return document.querySelector('input[type="number"]')!; }
function submit(): HTMLButtonElement { return document.querySelector('button[type="submit"]')!; }
async function setDays(value: string) { days().value = value; days().dispatchEvent(new Event('input', { bubbles: true })); await tick(); }
it('saves disabled defaults and retains invalid or failed drafts', async () => {
  const saved: unknown[] = [];
  const component = mount(ExpirationReminderEditor, { target: document.body, props: { initialPolicy: policy, onSave: async (value) => { saved.push(value); throw new Error('server'); } } });
  try {
    Array.from(document.querySelectorAll('button')).find(button => button.textContent?.includes('Before expiration'))!.click();
    await tick();
    await setDays('12.5');
    expect(submit().disabled).toBe(true);
    await setDays('14');
    submit().click();
    await vi.waitFor(() => expect(document.querySelector('[role="alert"]')?.textContent).toContain('could not be saved'));
    expect(saved).toEqual([{ ...policy, advanceDays: 14 }]);
    expect(days().value).toBe('14');
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
it('allows a type to override disabled defaults and restore inheritance', async () => {
  const saved: unknown[] = [];
  const component = mount(ExpirationReminderEditor, { target: document.body, props: { initialPolicy: null, inheritedPolicy: { ...policy, enabled: false }, onSave: async (value) => { saved.push(value); } } });
  try {
    expect(days()).toBeNull();
    Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Custom')!.click();
    await vi.waitFor(() => expect(saved).toEqual([policy]));
    await tick();
    Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Use defaults')!.click();
    await vi.waitFor(() => expect(saved).toEqual([policy, null]));
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
