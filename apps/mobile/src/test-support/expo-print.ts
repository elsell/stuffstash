// Controlled native print surface; physical output is verified separately on device.
const documents: string[] = [];
let failure: Error | undefined;
export async function printAsync(options: { uri?: string; html?: string }): Promise<void> {
  if (failure) throw failure;
  const document = options.uri ?? options.html;
  if (!document) throw new Error('A printable document is required.');
  documents.push(document);
}
export function printedDocuments() { return [...documents]; }
export function setPrintFailure(error?: Error) { failure = error; }
export function resetPrintState() { documents.length = 0; failure = undefined; }
