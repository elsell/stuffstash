import { expect, type APIRequestContext, type APIResponse } from '@playwright/test';
import { createHash } from 'node:crypto';

export type ArchivePhoto = { bytes: Buffer; assetId: string; attachmentId: string };

// Keep transport failures from including bearer headers in Playwright call logs.
async function authenticated(operation: () => Promise<APIResponse>) {
  try { return await operation(); }
  catch { throw new Error('Connected archive media request did not return a response.'); }
}

export async function addArchivePhoto(request: APIRequestContext, assetURL: string, token: string, bytes: Buffer): Promise<ArchivePhoto> {
  const response = await authenticated(() => request.post(`${assetURL}/attachments`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { fileName: 'archive-photo.png', contentType: 'image/png', contentBase64: bytes.toString('base64') },
  }));
  expect(response.status()).toBe(201);
  const { data } = await response.json();
  expect(data.sha256).toBe(createHash('sha256').update(bytes).digest('hex'));
  return { bytes, assetId: data.assetId, attachmentId: data.id };
}

export async function verifyRestoredPhoto(request: APIRequestContext, inventoryPath: string, token: string, title: string, source: ArchivePhoto) {
  const url = `http://localhost:8080${inventoryPath}`;
  const headers = { Authorization: `Bearer ${token}` };
  const assetsResponse = await authenticated(() => request.get(`${url}/assets`, { headers }));
  expect(assetsResponse.status()).toBe(200);
  const { data: assets } = await assetsResponse.json();
  const restored = assets.find((asset: { title: string }) => asset.title === title);
  expect(restored).toBeDefined();
  expect(restored.id).not.toBe(source.assetId);
  const attachmentURL = `${url}/assets/${restored.id}/attachments`;
  const listResponse = await authenticated(() => request.get(attachmentURL, { headers }));
  expect(listResponse.status()).toBe(200);
  const { data: attachments } = await listResponse.json();
  expect(attachments).toHaveLength(1);
  const photo = attachments[0];
  expect(photo.id).not.toBe(source.attachmentId);
  expect(photo.fileName).toBe('archive-photo.png');
  expect(photo.sha256).toBe(createHash('sha256').update(source.bytes).digest('hex'));
  const content = await authenticated(() => request.get(`${attachmentURL}/${photo.id}/content`, { headers }));
  expect(content.status()).toBe(200);
  expect((await content.body()).equals(source.bytes)).toBe(true);
}
