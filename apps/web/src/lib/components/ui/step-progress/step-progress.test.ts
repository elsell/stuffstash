import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import StepProgressHarness from './step-progress.test-harness.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('StepProgress', () => {
  it('renders a visual step list with current, complete, and upcoming states', () => {
    component = mount(StepProgressHarness, {
      target: document.body
    });

    const progress = document.body.querySelector<HTMLOListElement>('ol[aria-label="Onboarding progress"]');
    expect(progress).not.toBeNull();
    expect(progress?.style.getPropertyValue('--step-count')).toBe('4');
    expect(progress?.querySelectorAll('.step-progress-marker')).toHaveLength(4);
    expect(progress?.querySelectorAll('.step-progress-item.complete')).toHaveLength(2);
    expect(progress?.querySelectorAll('.step-progress-item.current')).toHaveLength(1);
    expect(progress?.querySelectorAll('.step-progress-item.upcoming')).toHaveLength(1);
    expect(progress?.querySelector('[aria-current="step"]')?.textContent).toContain('Invite');
  });

  it('only makes reachable steps interactive', () => {
    const onNavigateStep = vi.fn();
    component = mount(StepProgressHarness, {
      target: document.body,
      props: { onNavigateStep }
    });

    expect(buttonsNamed('Account')).toHaveLength(1);
    expect(buttonsNamed('Home')).toHaveLength(1);
    expect(buttonsNamed('Invite')).toHaveLength(1);
    expect(buttonsNamed('Done')).toHaveLength(0);

    buttonsNamed('Home')[0]?.click();
    expect(onNavigateStep).toHaveBeenCalledWith('home');
  });
});

function buttonsNamed(label: string): HTMLButtonElement[] {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).filter((button) => button.textContent?.trim() === label);
}
