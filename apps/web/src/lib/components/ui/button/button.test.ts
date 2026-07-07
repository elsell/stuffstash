import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import ButtonHarness from './button.test-harness.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
	if (component) {
		unmount(component);
		component = null;
	}
	document.body.innerHTML = '';
});

describe('Button', () => {
	it('renders disabled links with disabled semantics and visual affordance classes', () => {
		component = mount(ButtonHarness, {
			target: document.body
		});

		const link = document.body.querySelector<HTMLAnchorElement>('a');
		expect(link).not.toBeNull();
		expect(link?.getAttribute('href')).toBeNull();
		expect(link?.getAttribute('aria-disabled')).toBe('true');
		expect(link?.getAttribute('tabindex')).toBe('0');
		expect(link?.className).toContain('aria-disabled:pointer-events-none');
		expect(link?.className).toContain('aria-disabled:opacity-50');
	});

	it('does not allow caller props to override disabled link semantics', () => {
		component = mount(ButtonHarness, {
			target: document.body
		});

		const link = document.body.querySelector<HTMLAnchorElement>('[data-testid="conflicting-disabled-link"]');
		expect(link).not.toBeNull();
		expect(link?.getAttribute('href')).toBeNull();
		expect(link?.getAttribute('aria-disabled')).toBe('true');
		expect(link?.getAttribute('tabindex')).toBe('0');
	});

	it('prevents disabled link activation before caller handlers run', () => {
		const onDisabledActivate = vi.fn();
		component = mount(ButtonHarness, {
			target: document.body,
			props: { onDisabledActivate }
		});

		const link = document.body.querySelector<HTMLAnchorElement>('[data-testid="disabled-action-link"]');
		expect(link).not.toBeNull();

		link?.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
		link?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
		link?.dispatchEvent(new KeyboardEvent('keydown', { key: ' ', bubbles: true, cancelable: true }));

		expect(onDisabledActivate).not.toHaveBeenCalled();
	});

	it('renders explicit busy button content with a spinner and operation label', () => {
		component = mount(ButtonHarness, {
			target: document.body
		});

		const ready = document.body.querySelector<HTMLButtonElement>('[data-testid="ready-button"]');
		const busy = document.body.querySelector<HTMLButtonElement>('[data-testid="busy-button"]');

		expect(ready?.textContent).toContain('Confirm connection');
		expect(ready?.querySelector('.busy-button-spinner')).toBeNull();
		expect(ready?.querySelector('.busy-button-content')).not.toBeNull();
		expect(busy?.textContent).toContain('Confirming connection');
		expect(busy?.querySelector('.busy-button-spinner')).not.toBeNull();
		expect(busy?.querySelector('.busy-button-content')?.textContent?.trim()).toBe('Confirming connection');
		expect(busy?.querySelector('.busy-button-content')?.children).toHaveLength(2);
		expect(busy?.disabled).toBe(true);
	});
});
