import { ApiLabelRepository } from '../src/adapters/labels/ApiLabelRepository';
import type { LabelWorkspace } from '../src/application/labels/LabelWorkspace';
import { Image } from 'react-native';
import { fetch as expoFetch } from 'expo/fetch';
import { LabelsClient, PrintingClient } from '@stuff-stash/api-client';
import { ApiPrintingRepository } from '../src/adapters/printing/ApiPrintingRepository';
import { ExpoLabelFiles } from '../src/adapters/labels/ExpoLabelFiles';
import { createExpoArchiveFetch } from '../src/adapters/archives/ExpoArchiveFetch';
import { PrintingFake } from '../src/test-support/PrintingFake';
import type { PrintScope, PrintTemplate, RegisteredPrinter, PrintingWorkspace } from '../src/application/printing/PrintingWorkspace';

export const completePrintScope = { tenantId: 'tenant', inventoryId: 'inventory' };
/** Jobs are in memory; binary rendering alone crosses a real loopback HTTP boundary. */
class CompletePrintingRepository extends PrintingFake {
  private attempts = 0;
  constructor(private readonly binary: ApiPrintingRepository) {
    super(); this.printer = { ...this.printer, readiness: 'ready' };
    this.submitted.set('uncertain', { id: 'uncertain', assetId: 'printing-item', printerId: 'printer', status: 'uncertain', revision: 1, copies: 1, completedCopies: 0, idleConfirmed: true, latestAttemptId: 'attempt' });
  }
  override async preview(scope?: PrintScope, assetId?: string, printer?: RegisteredPrinter, template?: PrintTemplate, signal?: AbortSignal) {
    if (!scope || !assetId || !printer || !template || !signal) throw new Error('Preview context required');
    if (++this.attempts <= 1) throw new Error('Controlled automatic preview failure');
    return this.binary.preview(scope, assetId, printer, template, signal);
  }
}
export function createCompletePrintingWorkspace(): PrintingWorkspace {
  const baseUrl = process.env.EXPO_PUBLIC_STUFF_STASH_PRINT_AUDIT_URL;
  if (!baseUrl || new URL(baseUrl).hostname !== '127.0.0.1') throw new Error('A loopback print audit peer is required');
  const options = { baseUrl, tokenProvider: async () => null, fetch: createExpoArchiveFetch(expoFetch) };
  const repository = new CompletePrintingRepository(new ApiPrintingRepository(new PrintingClient(options), new LabelsClient(options)));
  const files = new ExpoLabelFiles(() => { throw new Error('Audit preview cleanup failed'); });
  return { ...repository.workspace(), files: {
    async preview(file, signal) {
      const local = await files.preview(file, signal);
      try {
        const size = await Image.getSize(local.uri);
        if (size.width !== file.width || size.height !== file.height) throw new Error('Native PNG decoding returned different dimensions');
        return local;
      } catch (error) { local.release(); throw error; }
    },
    async deliver() { throw new Error('This audit must never send physical output'); }
  } };
}

export function createCompleteLabelWorkspace(printing: PrintingWorkspace): LabelWorkspace {
  const baseUrl = process.env.EXPO_PUBLIC_STUFF_STASH_PRINT_AUDIT_URL;
  if (!baseUrl || new URL(baseUrl).hostname !== '127.0.0.1') throw new Error('A loopback print audit peer is required');
  return { repository: new ApiLabelRepository(new LabelsClient({ baseUrl, tokenProvider: async () => null, fetch: createExpoArchiveFetch(expoFetch) })), files: printing.files, parse: () => { throw new Error('Not a scanner fixture'); } };
}
