import { useEffect, useState } from 'react';
import type { LabelFile } from '../../application/labels/LabelWorkspace';
import type { PrintScope, PrintTemplate, PrintingWorkspace, RegisteredPrinter } from '../../application/printing/PrintingWorkspace';
type Preview = { uri: string; release(): void; file: LabelFile; fingerprint: string; key: string };
/** A render/file lifetime is shorter than the authorized task and never owns output. */
export function usePrintPreview(workspace: PrintingWorkspace, scope: PrintScope, assetId: string, printer: RegisteredPrinter | undefined, template: PrintTemplate | undefined, owner: AbortController | undefined, enabled: boolean) {
  const key = enabled && printer && template ? JSON.stringify([scope.tenantId, scope.inventoryId, assetId, printer.id, printer.mediaFingerprint, template.id, template.version, template.showReference]) : '';
  const [state, setState] = useState<{ key: string; preview?: Preview; failed?: boolean; loading?: boolean }>({ key: '' });
  const [retry, setRetry] = useState(0);
  useEffect(() => {
    if (!key || !owner || owner.signal.aborted) { setState({ key }); return; }
    if (!printer || !template) return;
    const controller = new AbortController(); const abort = () => controller.abort();
    owner.signal.addEventListener('abort', abort, { once: true });
    let local: { release(): void } | undefined;
    setState({ key, loading: true });
    void (async () => {
      try {
        const result = await workspace.repository.preview(scope, assetId, printer, template, controller.signal);
        if (controller.signal.aborted) return;
        const file = await workspace.files.preview(result.file, controller.signal);
        if (controller.signal.aborted) { file.release(); return; }
        local = file; setState({ key, preview: { ...file, ...result, key } });
      } catch { if (!controller.signal.aborted) setState({ key, failed: true }); }
    })();
    return () => { controller.abort(); owner.signal.removeEventListener('abort', abort); local?.release(); };
  }, [workspace, key, owner, retry]);
  return { preview: state.key === key ? state.preview : undefined, loading: !!key && (state.key !== key || !!state.loading), failed: state.key === key && !!state.failed, retry: () => setRetry(value => value + 1) };
}
