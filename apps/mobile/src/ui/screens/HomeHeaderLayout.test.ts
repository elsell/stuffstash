import { expect, it } from 'vitest';
import { homeInventoryControlWidth } from './HomeHeaderLayout';
it.each([320, 375, 390, 430, 768])('reserves native action and grouping space at viewport width %s', width => {
  const selector = homeInventoryControlWidth(width, 3);
  expect(selector).toBeGreaterThanOrEqual(44);
  expect(selector + 3 * 52 + 80).toBeLessThanOrEqual(width);
  expect(selector).toBeLessThanOrEqual(320);
});

it('gives a viewer inventory name the space freed by absent Add and notification actions', () => {
  const viewerWidth = homeInventoryControlWidth(402, 1);
  expect(viewerWidth).toBeGreaterThan(220);
  expect(viewerWidth + 52 + 80).toBeLessThanOrEqual(402);
  expect(viewerWidth).toBeGreaterThan(homeInventoryControlWidth(402, 3));
});
