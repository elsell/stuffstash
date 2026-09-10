import { expect, it } from 'vitest';
import { validExpirationInput } from './expiration';

it('validates exact dates and month precision without browser normalization', () => {
  expect(validExpirationInput('', 'day')).toBe(true);
  expect(validExpirationInput('2028-02', 'month')).toBe(true);
  expect(validExpirationInput('2028-02-29', 'day')).toBe(true);
  for (const date of ['2027-02-29', '2028-04-31', '0000-01-01', '2028-1-01', 'soon']) {
    expect(validExpirationInput(date, 'day')).toBe(false);
  }
  for (const date of ['2028-13', '2028-00', '2028-02-01', '0000-01', 'soon']) {
    expect(validExpirationInput(date, 'month')).toBe(false);
  }
});
