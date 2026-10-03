/** React Native Blob lacks arrayBuffer on some runtimes; FileReader data URLs are supported. */
export async function labelBlobBytes(blob: Blob): Promise<Uint8Array> {
  if (typeof blob.arrayBuffer === 'function') return new Uint8Array(await blob.arrayBuffer());
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('Label content could not be read.'));
    reader.onload = () => {
      try {
        const result = String(reader.result); const comma = result.indexOf(',');
        if (comma < 0 || !result.slice(0, comma).endsWith(';base64')) throw new Error('Invalid label content.');
        const binary = atob(result.slice(comma + 1));
        resolve(Uint8Array.from(binary, char => char.charCodeAt(0)));
      } catch (error) { reject(error); }
    };
    reader.readAsDataURL(blob);
  });
}
