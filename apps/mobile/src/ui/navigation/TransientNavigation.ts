export function navigateAfterTransientDismissal(
  dismiss: () => void,
  navigate: () => void
): void {
  dismiss();
  navigate();
}
