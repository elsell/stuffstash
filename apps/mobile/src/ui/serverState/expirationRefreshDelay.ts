// Scheduling belongs to the client adapter; expiration classification remains server-owned.
export function expirationRefreshDelay(data: unknown, now: Date, fetchedAt?: Date): number | false {
  const zones = new Set<string>();
  const seen = new WeakSet<object>();
  function visit(value: unknown) {
    if (!value || typeof value !== 'object' || seen.has(value)) return;
    seen.add(value);
    if ('expirationContext' in value) {
      const context = value.expirationContext;
      if (context && typeof context === 'object' && 'timezone' in context && typeof context.timezone === 'string') zones.add(context.timezone);
    }
    for (const child of Object.values(value)) visit(child);
  }
  visit(data);
  let delay = Infinity;
  for (const zone of zones) {
    const formatter = new Intl.DateTimeFormat('en-CA', { timeZone: zone, year: 'numeric', month: '2-digit', day: '2-digit' });
    const today = formatter.format(now);
    if (fetchedAt && formatter.format(fetchedAt) !== today) return 30_000;
    let low = now.getTime(), high = low + 36 * 60 * 60 * 1000;
    while (high - low > 1000) {
      const middle = Math.floor((low + high) / 2);
      if (formatter.format(new Date(middle)) === today) low = middle;
      else high = middle;
    }
    delay = Math.min(delay, high - now.getTime());
  }
  return Number.isFinite(delay) ? Math.max(1000, delay) : false;
}
