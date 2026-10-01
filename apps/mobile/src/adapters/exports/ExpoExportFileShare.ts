import { t } from '../../presentation/localization';
import * as Sharing from 'expo-sharing';
import type { ExportShareSheet } from './NativeExportFileDelivery';
export class ExpoExportFileShare implements ExportShareSheet {
  isAvailable() { return Sharing.isAvailableAsync(); }
  open(uri: string, format: 'json' | 'csv') {
    return Sharing.shareAsync(uri, { dialogTitle: t('sharing.export.title'), mimeType: format === 'json' ? 'application/json' : 'text/csv', UTI: format === 'json' ? 'public.json' : 'public.comma-separated-values-text' });
  }
}
