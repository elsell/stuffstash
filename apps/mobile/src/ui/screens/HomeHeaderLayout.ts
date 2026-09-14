const nativeActionSlot = 52;
// Reserve navigation margins, the gap between groups and native glass insets.
const nativeGroupAllowance = 80;
const maximumSelectorWidth = 180;
const minimumControlWidth = 44;

export function homeInventoryControlWidth(viewportWidth: number, actionCount: number): number {
  return Math.max(minimumControlWidth, Math.min(maximumSelectorWidth,
    viewportWidth - actionCount * nativeActionSlot - nativeGroupAllowance));
}
