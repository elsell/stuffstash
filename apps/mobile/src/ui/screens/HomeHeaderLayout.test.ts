import { expect, it } from 'vitest';
import { homeInventoryControlWidth } from './HomeHeaderLayout';
it.each([320, 375, 390, 430, 768])('reserves native action and grouping space at viewport width %s', width => {
  const selector = homeInventoryControlWidth(width, 3);
  expect(selector).toBeGreaterThanOrEqual(44);
  expect(selector + 3 * 52 + 80).toBeLessThanOrEqual(width);
  expect(selector).toBeLessThanOrEqual(180);
});
