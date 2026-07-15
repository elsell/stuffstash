import { describe, expect, it } from 'vitest';
import {
  buildImportSourceRequest,
  importSourceRequestKey,
  normalizeImportSourceRequest,
  readableImportActionError
} from './workspaceImportRequest';

describe('workspace import request helpers', () => {
  it('normalizes schemeless live Homebox URLs before crossing the repository boundary', () => {
    expect(
      normalizeImportSourceRequest({
        sourceType: 'legacy_homebox',
        baseUrl: 'stuff.jsksell.com',
        username: ' codex@jsksell.com ',
        password: 'secret',
        includeImages: true,
        allowPrivateNetwork: false,
        allowInsecureTLS: false
      })
    ).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'https://stuff.jsksell.com',
      username: 'codex@jsksell.com'
    });
  });

  it('preserves explicit http Homebox URLs', () => {
    expect(
      normalizeImportSourceRequest({
        sourceType: 'legacy_homebox',
        baseUrl: 'http://homebox.local:3100',
        username: 'codex@jsksell.com',
        password: 'secret'
      })
    ).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'http://homebox.local:3100'
    });
  });

  it('builds CSV requests without live connection fields', () => {
    const request = buildImportSourceRequest({
      sourceChoice: 'homebox_csv',
      baseUrl: 'stuff.jsksell.com',
      username: 'codex@jsksell.com',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: true,
      allowInsecureTLS: true,
      fileName: 'homebox.csv',
      contentBase64: 'a,b,c',
      csvSelection: { name: 'homebox.csv', size: 100, lastModified: 1 }
    });

    expect(request).toEqual({
      sourceType: 'legacy_homebox_csv',
      fileName: 'homebox.csv',
      contentBase64: 'a,b,c'
    });
  });

  it('does not include full CSV bytes in the preview key', () => {
    const key = importSourceRequestKey({
      sourceChoice: 'homebox_csv',
      baseUrl: '',
      username: '',
      password: '',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false,
      fileName: 'homebox.csv',
      contentBase64: 'very-large-base64-payload',
      csvSelection: { name: 'homebox.csv', size: 10, lastModified: 1 }
    });

    expect(key).toContain('"contentLength":25');
    expect(key).toContain('contentFingerprint');
    expect(key).not.toContain('very-large-base64-payload');
  });

  it('changes the preview key when CSV content changes with the same file metadata and length', () => {
    const first = importSourceRequestKey({
      sourceChoice: 'homebox_csv',
      baseUrl: '',
      username: '',
      password: '',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false,
      fileName: 'homebox.csv',
      contentBase64: 'HB.name,Drill',
      csvSelection: { name: 'homebox.csv', size: 128, lastModified: 1 }
    });
    const second = importSourceRequestKey({
      sourceChoice: 'homebox_csv',
      baseUrl: '',
      username: '',
      password: '',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false,
      fileName: 'homebox.csv',
      contentBase64: 'HB.name,Shelf',
      csvSelection: { name: 'homebox.csv', size: 128, lastModified: 1 }
    });

    expect(first).not.toBe(second);
    expect(first).not.toContain('HB.name,Drill');
    expect(second).not.toContain('HB.name,Shelf');
  });

  it('does not include the live Homebox password in the preview key', () => {
    const key = importSourceRequestKey({
      sourceChoice: 'homebox_live',
      baseUrl: 'http://homebox.local',
      username: 'codex@jsksell.com',
      password: 'super-secret-password',
      includeImages: true,
      allowPrivateNetwork: true,
      allowInsecureTLS: false,
      fileName: '',
      contentBase64: '',
      csvSelection: null
    });

    expect(key).toContain('passwordFingerprint');
    expect(key).not.toContain('super-secret-password');
  });

  it('maps generic validation errors to contextual import copy without hiding specific errors', () => {
    const generic = Object.assign(new Error('Invalid request.'), { status: 400, code: 'invalid_request' });
    const specific = Object.assign(new Error('Base URL must use http or https.'), { status: 400, code: 'invalid_request' });

    expect(readableImportActionError(generic, 'Homebox connection could not be confirmed.')).toBe(
      'Homebox connection could not be confirmed.'
    );
    expect(readableImportActionError(specific, 'Homebox connection could not be confirmed.')).toBe(
      'Base URL must use http or https.'
    );
  });

  it('does not expose untyped server diagnostics as import recovery copy', () => {
    const serverFailure = Object.assign(new Error('RAW_SENTINEL provider stack'), { status: 500 });
    const providerFailure = Object.assign(new Error('RAW_SENTINEL provider response'), { status: 400, code: 'provider_error' });

    expect(readableImportActionError(serverFailure, 'Cancellation could not be requested.')).toBe(
      'Cancellation could not be requested.'
    );
    expect(readableImportActionError(new Error('RAW_SENTINEL transport failure'), 'Import history could not be loaded.')).toBe(
      'Import history could not be loaded.'
    );
    expect(readableImportActionError(providerFailure, 'Homebox connection could not be confirmed.')).toBe(
      'Homebox connection could not be confirmed.'
    );
  });

  it('preserves import errors explicitly marked safe for users', () => {
    const safeFailure = Object.assign(new Error('This import has already finished.'), { safeForUser: true });
    expect(readableImportActionError(safeFailure, 'Cancellation could not be requested.')).toBe(
      'This import has already finished.'
    );
  });
});
