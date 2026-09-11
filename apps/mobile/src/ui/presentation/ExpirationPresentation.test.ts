import { expect, it } from 'vitest';
import { formatAssetExpiration } from './ExpirationPresentation';
it('keeps the printed precision in readable dates', () => {
 expect(formatAssetExpiration({ date: '2028-02', precision: 'month' }, 'en-US')).toBe('February 2028');
 expect(formatAssetExpiration({ date: '2028-02-29', precision: 'day' }, 'en-US')).toBe('February 29, 2028');
});
