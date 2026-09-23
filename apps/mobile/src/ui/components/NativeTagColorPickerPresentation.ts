export function nativeTagColorSelection(value: string): string | null {
  return /^#[0-9A-F]{6}$/i.test(value) ? value.toUpperCase() : null;
}
