import { expect, it } from 'vitest';

it('starts and formats messages when optional Hermes Intl APIs are absent', async () => {
  const names = ['getCanonicalLocales', 'Locale', 'PluralRules', 'ListFormat'] as const;
  const original = names.map(name => Object.getOwnPropertyDescriptor(Intl, name));
  try {
    for (const name of names) Reflect.deleteProperty(Intl, name);
    const { localization } = await import('../../presentation/localization');
    expect(localization.list(['Garage', 'Kitchen'])).toBe('Garage and Kitchen');
    expect(new Intl.PluralRules('en').select(1)).toBe('one');
    expect(new Intl.PluralRules('en').select(0)).toBe('other');
    expect(localization.number(42)).toBe('42');
  } finally {
    names.forEach((name, index) => {
      const descriptor = original[index];
      if (descriptor) Object.defineProperty(Intl, name, descriptor);
      else Reflect.deleteProperty(Intl, name);
    });
  }
});
