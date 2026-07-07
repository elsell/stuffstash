import { describe, expect, it } from 'vitest';
import { directUploadMethod, isDirectUploadTargetSupported } from './DirectUploadPolicy';

describe('DirectUploadPolicy', () => {
  it('allows HTTPS direct upload targets by default', () => {
    expect(isDirectUploadTargetSupported('https://uploads.example.test/object-one')).toBe(true);
  });

  it('rejects local development targets unless explicitly enabled', () => {
    expect(isDirectUploadTargetSupported('stuffstash-local://direct-uploads/upload-one')).toBe(false);
    expect(isDirectUploadTargetSupported('http://192.168.1.12:3900/object-one')).toBe(false);

    expect(
      isDirectUploadTargetSupported(
        'stuffstash-local://direct-uploads/upload-one',
        { allowLocalDevelopmentTargets: true }
      )
    ).toBe(true);
    expect(
      isDirectUploadTargetSupported(
        'http://192.168.1.12:3900/object-one',
        { allowLocalDevelopmentTargets: true }
      )
    ).toBe(true);
  });

  it('rejects public and malformed cleartext HTTP targets even when local targets are enabled', () => {
    const policy = { allowLocalDevelopmentTargets: true };

    expect(isDirectUploadTargetSupported('http://uploads.example.test/object-one', policy)).toBe(false);
    expect(isDirectUploadTargetSupported('http://192.168.1.1evil:3900/object-one', policy)).toBe(false);
    expect(isDirectUploadTargetSupported('http://192.168.one.12:3900/object-one', policy)).toBe(false);
    expect(isDirectUploadTargetSupported('http://256.168.1.12:3900/object-one', policy)).toBe(false);
  });

  it('normalizes supported direct upload methods', () => {
    expect(directUploadMethod('post')).toBe('POST');
    expect(directUploadMethod(' PUT ')).toBe('PUT');
    expect(directUploadMethod('PATCH')).toBe('PATCH');
    expect(() => directUploadMethod('DELETE')).toThrow('Unsupported direct attachment upload method.');
  });
});
