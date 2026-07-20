export function nativeTagColorSelection(value: string): string | null {
  return /^#[0-9A-F]{6}$/i.test(value) ? value.toUpperCase() : null;
}

export function nativeTagColorInteraction(disabled: boolean, onChange: (value: string) => void) {
  return {
    pointerEvents: disabled ? 'none' as const : 'auto' as const,
    onSelectionChange: disabled ? undefined : onChange
  };
}
