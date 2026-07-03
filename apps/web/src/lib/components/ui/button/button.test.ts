import { afterEach, describe, expect, it } from 'vitest';
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
		expect(link?.getAttribute('tabindex')).toBe('-1');
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
		expect(link?.getAttribute('tabindex')).toBe('-1');
	});
});
