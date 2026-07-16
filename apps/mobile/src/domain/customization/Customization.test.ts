import { describe, expect, it } from 'vitest';
import { customizationKeyIsValid, customizationNameOrder, normalizeTagColor, suggestedCustomizationKey } from './Customization';

describe('mobile customization values', () => {
  it('derives stable reviewable keys', () => {
    expect(suggestedCustomizationKey('  Expiration Dáte  ')).toBe('expiration-date');
    expect(suggestedCustomizationKey('123 Batteries')).toBe('batteries');
    expect(customizationKeyIsValid('expiration-date')).toBe(true);
    expect(customizationKeyIsValid('Emoji-🔧')).toBe(false);
  });

  it('sorts by localized display name with stable id tie breaking', () => {
    expect(customizationNameOrder([
      { id: 'b', displayName: 'tools' },
      { id: 'c', displayName: 'Camping' },
      { id: 'a', displayName: 'Tools' }
    ]).map((item) => item.id)).toEqual(['c', 'a', 'b']);
  });

  it('normalizes valid colors and rejects stale invalid values', () => {
    expect(normalizeTagColor('2f80ed')).toBe('#2F80ED');
    expect(normalizeTagColor('oops')).toBeUndefined();
    expect(normalizeTagColor('')).toBeUndefined();
  });
});
