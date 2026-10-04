import { printingLabelPNG } from './PrintingLabelSample';
import { Stack, useRouter } from 'expo-router';
import { PrintingFake } from '../src/test-support/PrintingFake';
import type { LabelFile, LabelRepository, LabelScope, LabelSelection, LabelWorkspace } from '../src/application/labels/LabelWorkspace';
import { PrinterSettingsScreen } from '../src/ui/printing/PrinterSettingsScreen';
import { PrinterDefaultsScreen } from '../src/ui/printing/PrinterDefaultsScreen';
import { PrinterDetailScreen } from '../src/ui/printing/PrinterDetailScreen';
import { PrintHistoryScreen } from '../src/ui/printing/PrintHistoryScreen';
import { LabelOptionsScreen } from '../src/ui/labels/LabelOptionsScreen';

const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
class ConnectedPrintingRepository extends PrintingFake {
  async catalog() {
    const catalog = await super.catalog();
    return { ...catalog, connectors: [{ id: 'garage-computer', name: 'Garage computer', state: 'active', availability: 'online', lastSeenAt: '2026-10-04T12:00:00Z', printerIds: [this.printer.id] }] };
  }
}
const repository = new ConnectedPrintingRepository();
repository.printer = { ...repository.printer, readiness: 'ready', reportedAt: '2026-10-04T12:00:00Z' };
repository.submitted.set('completed-label', { id: 'completed-label', printerId: 'printer', assetId: 'printing-item', status: 'completed', revision: 1, copies: 1, completedCopies: 1 });
const workspace = repository.workspace();
function open(path: string, router: ReturnType<typeof useRouter>) { router.push(path as never); }
const base = '/settings/printers';
const unsupported = () => { throw new Error('This fixture verifies layout and drafts, not hardware output.'); };

export function PrintingSettingsFixture() {
  const router = useRouter();
  return <><Stack.Screen options={{ title: 'Printers and labels' }} /><PrinterSettingsScreen workspace={workspace} scope={scope}
    onPrinter={() => open(`${base}/printer`, router)} onDefaults={() => open(`${base}/defaults`, router)} onHistory={() => open(`${base}/history`, router)} /></>;
}
export function PrintingDefaultsFixture() { return <><Stack.Screen options={{ title: 'Print defaults' }} /><PrinterDefaultsScreen workspace={workspace} scope={scope} canConfigure /></>; }
export function PrintingDetailFixture() { return <><Stack.Screen options={{ title: 'Garage' }} /><PrinterDetailScreen workspace={workspace} scope={scope} printerId="printer" canConfigure canPrint onJob={unsupported} /></>; }
export function PrintingHistoryFixture() { return <><Stack.Screen options={{ title: 'Print history' }} /><PrintHistoryScreen workspace={workspace} scope={scope} onJob={unsupported} /></>; }

class RecoveringLabelRepository implements LabelRepository {
  private renders = 0;
  async instance() { return 'audit-instance'; }
  async resolve() { return { ...scope, assetId: 'printing-item', archived: false }; }
  async catalog() {
    return { profiles: [{ id: 'brother-ql800-29x90', name: '29 × 90 mm', media: repository.printer.media }], templates: (await repository.catalog()).templates };
  }
  async render(selectedScope: LabelScope, assetId: string, selection: LabelSelection, format: 'png' | 'pdf', signal: AbortSignal): Promise<LabelFile> {
    if (signal.aborted || selectedScope.tenantId !== scope.tenantId || selectedScope.inventoryId !== scope.inventoryId || assetId !== 'printing-item') throw new Error('Unavailable');
    if (++this.renders === 1) throw new Error('Controlled first render failure');
    if (selection.template.id !== 'qr-title' || format !== 'png') throw new Error('This audit only renders the registered PNG preview.');
    return { bytes: Uint8Array.from(atob(printingLabelPNG), character => character.charCodeAt(0)), format, width: 306, height: 991, rotation: 270 };
  }
}
const labels: LabelWorkspace = {
  repository: new RecoveringLabelRepository(), parse: () => ({ instanceId: 'audit-instance', labelId: 'printing-item' }),
  files: {
    async preview(file, signal) {
      if (signal.aborted || file.format !== 'png') throw new Error('Unavailable');
      // The bundled image is the real catalog renderer's 29 × 90 mm sample.
      const encoded = btoa(Array.from(file.bytes, byte => String.fromCharCode(byte)).join(''));
      return { uri: `data:image/png;base64,${encoded}`, release() {} };
    },
    async deliver() { unsupported(); }
  }
};
export function PrintingLabelsFixture() { return <><Stack.Screen options={{ title: 'Label' }} /><LabelOptionsScreen workspace={labels} scope={scope} assetId="printing-item" /></>; }
