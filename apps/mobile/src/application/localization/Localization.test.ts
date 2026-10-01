import { describe, expect, it } from 'vitest';
import { createTranslator } from '../../../../../packages/localization/src/translator';

const catalog = {
  filters: { one: '{count} filter applied', other: '{count} filters applied' },
  welcome: 'Welcome, {name}.',
  plain: 'Browse inventory',
} as const;

describe('shared client localization', () => {
  it('selects English plural forms for fallback copy and formats counts using the requested locale', () => {
    const t = createTranslator(catalog, { locale: 'fr-FR' });
    expect(t.message('filters', { count: 0 })).toBe('0 filters applied');
    expect(t.message('filters', { count: 1 })).toBe('1 filter applied');
    expect(t.message('filters', { count: 2500 })).toBe(`${new Intl.NumberFormat('fr-FR').format(2500)} filters applied`);
    expect(() => t.message('filters', { count: NaN })).toThrow();
  });
  it('interpolates once as plain text without treating parameter content as another template', () => {
    const t = createTranslator(catalog, { locale: 'en-US' });
    expect(t.message('welcome', { name: '<script>{count}</script>' })).toBe('Welcome, <script>{count}</script>.');
    expect(() => t.message('welcome')).toThrow(/name/);
  });
  it('expands catalog copy but preserves external text and keeps instances independent', () => {
    const expanded = createTranslator(catalog, { locale: 'en-XA' });
    const normal = createTranslator(catalog, { locale: 'en-US' });
    const message = expanded.message('welcome', { name: 'Garage / Bin 12' });
    expect(message).toContain('Garage / Bin 12');
    expect(message.length).toBeGreaterThan(normal.message('welcome', { name: 'Garage / Bin 12' }).length);
    expect(normal.message('plain')).toBe('Browse inventory');
    expect(expanded.direction).toBe('ltr');
    const rtl = createTranslator(catalog, { locale: 'ar-XB' });
    expect(rtl.direction).toBe('rtl');
    expect(rtl.message('welcome', { name: 'Bin 12' })).toContain('Bin 12');
  });
  it('uses translated plural categories and falls back per message', () => {
    const translated = createTranslator(catalog, { locale: 'ar', translations: { ar: {
      filters: { zero: 'zero {count}', one: 'one {count}', two: 'two {count}', few: 'few {count}', many: 'many {count}', other: 'other {count}' },
    } } });
    expect(translated.message('filters', { count: 2 })).toBe(`two ${new Intl.NumberFormat('ar').format(2)}`);
    expect(translated.message('plain')).toBe('Browse inventory');
  });
  it('snapshots source and translated catalogs, including nested plural maps', () => {
    const source = { title: 'Original', items: { one: '{count} item', other: '{count} items' } };
    const arabic = { items: { other: 'translated {count}' } };
    const normal = createTranslator(source, { locale: 'en' });
    const translated = createTranslator(source, { locale: 'ar', translations: { ar: arabic } });
    source.title = 'Changed'; source.items.other = 'changed {count}'; arabic.items.other = 'changed {count}';
    expect(normal.message('title')).toBe('Original');
    expect(normal.message('items', { count: 2 })).toBe('2 items');
    expect(translated.message('items', { count: 3 })).toBe(`translated ${new Intl.NumberFormat('ar').format(3)}`);
  });

});
