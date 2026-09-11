import { mount, unmount, tick } from 'svelte';
import { expect, it } from 'vitest';
import ExpirationField from './ExpirationField.svelte';
it('requires an explicit day after month precision and allows deliberate removal', async () => {
 const changes: unknown[] = [];
 const component = mount(ExpirationField, {target: document.body, props: {id: 'expiry', initialValue: {date: '2027-03', precision: 'month'}, onChange: (value, valid) => {changes.push({value, valid});}}});
 try {
  Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Exact date')!.click(); await tick();
  expect(changes.at(-1)).toEqual({value: undefined, valid: false});
  Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Clear expiration')!.click(); await tick();
  expect(changes.at(-1)).toEqual({value: undefined, valid: true});
  Array.from(document.querySelectorAll('button')).find(button => button.textContent === 'Month and year')!.click(); await tick();
  expect(document.querySelector('input')!.value).toBe('');
 } finally {await unmount(component); document.body.innerHTML = '';}
});
