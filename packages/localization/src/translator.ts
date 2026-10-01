/** Plain-text messages only; consumers render through native Text or escaped DOM text. */
export type PluralMessage = Readonly<Partial<Record<Intl.LDMLPluralRule, string>> & { other: string }>;
export type Message = string | PluralMessage;
export type Catalog = Readonly<Record<string, Message>>;
export type MessageValues = Readonly<Record<string, string | number>>;
export interface TranslatorOptions<C extends Catalog> {
  readonly locale?: string;
  readonly translations?: Readonly<Record<string, Partial<Record<keyof C, Message>>>>;
}

const placeholder = /\{([a-zA-Z][a-zA-Z0-9_]*)\}/g;

function expand(template: string): string {
  const accents: Readonly<Record<string, string>> = { a: 'á', e: 'ë', i: 'ï', o: 'ô', u: 'ü', A: 'Á', E: 'Ë', I: 'Ï', O: 'Ô', U: 'Ü' };
  return `[${template.split(/(\{[a-zA-Z][a-zA-Z0-9_]*\})/g).map((part) =>
    part.startsWith('{') ? part : part.replace(/[aeiouAEIOU]/g, (letter) => accents[letter])
  ).join('')}${'~'.repeat(Math.ceil(template.replace(placeholder, '').length * 0.3))}]`;
}

export function createTranslator<C extends Catalog>(catalog: C, options: TranslatorOptions<C> = {}) {
  const sourceCatalog: C = Object.freeze(Object.fromEntries(Object.entries(catalog).map(([key, value]) =>
    [key, typeof value === 'string' ? value : Object.freeze({ ...value })]
  ))) as C;
  const requested = options.locale || new Intl.DateTimeFormat().resolvedOptions().locale;
  const pseudo = requested.toLowerCase() === 'en-xa' ? 'expanded' : requested.toLowerCase() === 'ar-xb' ? 'rtl' : undefined;
  let locale: string;
  try { locale = pseudo ? 'en' : Intl.getCanonicalLocales(requested)[0] || 'en'; }
  catch { locale = 'en'; }
  const language = locale.split('-')[0];
  const translationLocale = options.translations?.[locale] ? locale : language;
  const translationInput = options.translations?.[translationLocale];
  const translated: Partial<Record<keyof C, Message>> = Object.freeze(Object.fromEntries(
    Object.entries(translationInput ?? {}).filter(([, value]) => value !== undefined).map(([key, value]) =>
      [key, typeof value === 'string' ? value : Object.freeze({ ...(value as PluralMessage) })]
    )
  )) as Partial<Record<keyof C, Message>>;
  const messageLanguage = translationInput ? translationLocale : 'en';
  const direction = pseudo === 'rtl' || ['ar', 'fa', 'he', 'ur'].includes(messageLanguage.split('-')[0]) ? 'rtl' : 'ltr';
  const numbers = new Intl.NumberFormat(locale);
  const englishPlurals = new Intl.PluralRules('en');
  const translatedPlurals = new Intl.PluralRules(translationLocale);

  function message(key: keyof C, values: MessageValues = {}): string {
    const localized = translated?.[key];
    const source = localized ?? sourceCatalog[key];
    if (source === undefined) throw new Error(`Unknown message: ${String(key)}`);
    let template: string;
    if (typeof source === 'string') template = source;
    else {
      if (typeof values.count !== 'number' || !Number.isFinite(values.count)) throw new Error(`Message ${String(key)} requires a finite count`);
      const category = (localized === undefined ? englishPlurals : translatedPlurals).select(values.count);
      template = source[category] ?? source.other;
    }
    if (pseudo === 'expanded') template = expand(template);
    const result = template.replace(placeholder, (_, name: string) => {
      if (!Object.hasOwn(values, name)) throw new Error(`Message ${String(key)} requires ${name}`);
      const value = values[name];
      return typeof value === 'number' ? numbers.format(value) : value;
    });
    return pseudo === 'rtl' ? `\u2067${result}\u2069` : result;
  }

  return Object.freeze({
    locale, messageLanguage, direction, message,
    number: (value: number, format?: Intl.NumberFormatOptions) => new Intl.NumberFormat(locale, format).format(value),
    date: (value: Date | number, format?: Intl.DateTimeFormatOptions) => new Intl.DateTimeFormat(locale, format).format(value),
    list: (values: readonly string[], format?: Intl.ListFormatOptions) => new Intl.ListFormat(locale, format).format(values),
    compare: (left: string, right: string) => new Intl.Collator(locale).compare(left, right),
  });
}
