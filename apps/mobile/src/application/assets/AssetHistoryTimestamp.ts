export type HistoryTimestampDetail = 'activity' | 'checkout' | 'exact';

const formats: Record<HistoryTimestampDetail, Intl.DateTimeFormatOptions> = {
  activity: { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' },
  checkout: { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' },
  exact: { dateStyle: 'long', timeStyle: 'long' }
};

/** The omitted locale and zone use the device's conventions. */
export function formatHistoryTimestamp(value: string, detail: HistoryTimestampDetail, locale?: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  return new Intl.DateTimeFormat(locale, formats[detail]).format(new Date(timestamp));
}
