import { expect, it } from 'vitest';
import { keyboardBoundaryInset } from './keyboardBoundaryInset';

it('clears only the part of a docked keyboard covering the sheet bottom', () => {
  expect(keyboardBoundaryInset({ x: 0, y: 844, width: 390 }, { screenX: 0, screenY: 544, width: 390, height: 300 })).toBe(300);
  expect(keyboardBoundaryInset({ x: 80, y: 760, width: 580 }, { screenX: 0, screenY: 650, width: 744, height: 400 })).toBe(110);
});
it('does not double-offset a resized sheet or chase a nonoverlapping floating keyboard', () => {
  const boundary = { x: 80, y: 650, width: 580 };
  expect(keyboardBoundaryInset(boundary, { screenX: 0, screenY: 650, width: 744, height: 400 })).toBe(0);
  expect(keyboardBoundaryInset(boundary, { screenX: 0, screenY: 200, width: 300, height: 250 })).toBe(0);
  expect(keyboardBoundaryInset(boundary, { screenX: 700, screenY: 500, width: 300, height: 300 })).toBe(0);
  expect(keyboardBoundaryInset(boundary, undefined)).toBe(0);
});
