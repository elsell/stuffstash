import { expect, it } from 'vitest';
import { expirationMonthCalendarNotice, expirationMonthOptions, formatAssetExpiration } from './ExpirationPresentation';
it('keeps the printed precision in readable dates', () => {
 expect(formatAssetExpiration({ date: '2028-02', precision: 'month' }, 'en-US')).toBe('February 2028');
 expect(formatAssetExpiration({ date: '2028-02-29', precision: 'day' }, 'en-US')).toBe('February 29, 2028');
});
it('preserves a Gregorian month period under an alternate calendar locale', () => {
 expect(formatAssetExpiration({ date: '2028-02', precision: 'month' }, 'en-US-u-ca-hebrew')).toBe('February 2028 (Gregorian)');
});
it('still localizes an exact day using the chosen calendar', () => {
 expect(formatAssetExpiration({ date: '2028-02-01', precision: 'day' }, 'en-US-u-ca-hebrew')).toContain('Shevat');
});

it('offers the same stored Gregorian months with localized names', () => {
 const choices = expirationMonthOptions('en-US-u-ca-hebrew');
 expect(choices).toHaveLength(12);
 expect(choices[0]).toEqual({ value: '1', label: 'January' });
 expect(choices[11]).toEqual({ value: '12', label: 'December' });
 expect(expirationMonthOptions('fr-FR')[1]).toEqual({ value: '2', label: 'février' });
});
it('explains the month-calendar restriction only when it differs from the locale', () => {
 expect(expirationMonthCalendarNotice('en-US')).toBeUndefined();
 expect(expirationMonthCalendarNotice('en-US-u-ca-hebrew')).toBe('Month and year use the Gregorian calendar.');
});
