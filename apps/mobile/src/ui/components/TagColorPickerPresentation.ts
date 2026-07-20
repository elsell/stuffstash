export function tagColorModalLayout(input: { readonly availableHeight: number; readonly fontScale: number }) {
  const compactSpectrum = input.availableHeight > 0
    ? input.availableHeight < 640 || input.fontScale >= 1.4
    : input.fontScale >= 1.4;
  return { compactSpectrum } as const;
}
