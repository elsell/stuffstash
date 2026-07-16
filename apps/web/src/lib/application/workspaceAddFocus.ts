export function addReturnFocusTarget(
  capturedOpener: HTMLElement | null,
  root: Pick<Document, 'querySelector'> = document,
  mobile = typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(max-width: 900px)').matches
): HTMLElement | null {
  const responsiveTrigger = mobile ? 'mobile' : 'desktop';
  const capturedResponsiveKind = capturedOpener?.dataset.workspaceAddTrigger;
  const capturedDocumentRoot = capturedOpener?.tagName === 'BODY' || capturedOpener?.tagName === 'HTML';
  if (capturedOpener?.isConnected && !capturedDocumentRoot && (!capturedResponsiveKind || capturedResponsiveKind === responsiveTrigger)) {
    return capturedOpener;
  }
  const localReturnFocusKey = capturedOpener?.dataset.workspaceAddReturnFocus;
  if (localReturnFocusKey) {
    const restoredLocalOpener = root.querySelector<HTMLElement>(
      `[data-workspace-add-return-focus="${localReturnFocusKey}"]`
    );
    if (restoredLocalOpener) {
      return restoredLocalOpener;
    }
  }
  return root.querySelector<HTMLElement>(`[data-workspace-add-trigger="${responsiveTrigger}"]`);
}
