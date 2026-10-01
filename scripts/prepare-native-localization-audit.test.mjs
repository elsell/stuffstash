import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, copyFile, readFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { execFileSync } from 'node:child_process';

test('native audit generates locale-specific labels and explicit layout direction', async () => {
  const directory = await mkdtemp(join(tmpdir(), 'native-locale-test-'));
  try {
    const target = join(directory, 'FixtureAuditTests.swift');
    await copyFile('apps/mobile/native-audit/FixtureAuditTests.swift', target);
    for (const locale of ['ar-XB', 'en-XA', 'en']) {
      execFileSync(process.execPath, ['scripts/prepare-native-localization-audit.mjs', locale, target]);
      const result = await readFile(target, 'utf8');
      assert.ok(result.includes(`private let auditLocalizationRTL = ${locale === 'ar-XB'}`));
      assert.equal((result.match(/private let auditLocalizationRTL =/g) ?? []).length, 1);
      assert.ok(result.includes('"Asset name":'));
    }
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
