import type { KeyboardMetrics } from 'react-native';

export type SheetBottomBoundary = { readonly x: number; readonly y: number; readonly width: number };

export function keyboardBoundaryInset(boundary: SheetBottomBoundary, keyboard: KeyboardMetrics | undefined): number {
  if (!keyboard || keyboard.height <= 0 || keyboard.width <= 0) return 0;
  if (keyboard.screenX >= boundary.x + boundary.width || keyboard.screenX + keyboard.width <= boundary.x) return 0;
  if (keyboard.screenY >= boundary.y || keyboard.screenY + keyboard.height < boundary.y) return 0;
  return boundary.y - keyboard.screenY;
}
