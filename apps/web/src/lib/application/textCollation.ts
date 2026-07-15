const naturalCollator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' });

export function compareNaturalText(left: string, right: string): number {
  return naturalCollator.compare(left, right);
}
