import assert from 'node:assert/strict';
import test from 'node:test';
import { releaseHighlights, renderPlayNotes } from './release-notes.mjs';

test('both stores share reviewed highlights and retain the TestFlight marker', () => {
  const messages = ['fix: Fallback\n\nRelease notes:\n- Shared fix\n\nInternal detail', 'feat: Fallback\n\nTestFlight notes:\n- Shared feature', 'fix: Shared fix', 'ci: Tooling'];
  assert.deepEqual(releaseHighlights(messages), ['Shared fix', 'Shared feature']);
  assert.equal(renderPlayNotes(messages), '- Shared fix\n- Shared feature');
});
test('Play notes compact by Unicode code point without splitting emoji', () => {
  const result = renderPlayNotes(['feat: ' + '😀'.repeat(600)]);
  assert.equal([...result].length, 500);
  assert.ok(result.endsWith('…'));
  assert.ok(!result.includes('\uFFFD'));
  assert.ok(renderPlayNotes([]).length > 0);
});

import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { readReleaseMessages } from './release-notes.mjs';

test('release history is bounded by stable tags and excludes later commits', () => {
 const cwd=mkdtempSync(join(tmpdir(),'store-history-'));
 const git=(...args)=>execFileSync('git',['-c','core.hooksPath=/dev/null',...args],{cwd,encoding:'utf8',stdio:['ignore','pipe','pipe'],env:{...Object.fromEntries(Object.entries(process.env).filter(([key])=>!key.startsWith('GIT_'))),GIT_AUTHOR_NAME:'Test',GIT_AUTHOR_EMAIL:'test@example.com',GIT_COMMITTER_NAME:'Test',GIT_COMMITTER_EMAIL:'test@example.com'}});
 try {
  git('init'); git('commit','--allow-empty','-m','feat: Initial'); git('tag','v1.0.0');
  git('commit','--allow-empty','-m','fix: Shared fix'); git('tag','v1.1.0'); git('tag','v99.0.0');
  git('commit','--allow-empty','-m','feat: Later');
  const oldGitDir = process.env.GIT_DIR;
  process.env.GIT_DIR = join(cwd, 'not-the-repository');
  try { assert.deepEqual(releaseHighlights(readReleaseMessages('v1.1.0',cwd)),['Shared fix']); }
  finally { if (oldGitDir === undefined) delete process.env.GIT_DIR; else process.env.GIT_DIR = oldGitDir; }
  assert.deepEqual(releaseHighlights(readReleaseMessages('v1.0.0',cwd)),['Initial']);
  assert.throws(()=>readReleaseMessages('--all',cwd),/Invalid release tag/);
  assert.throws(()=>readReleaseMessages('v2.0.0',cwd));
 } finally {rmSync(cwd,{recursive:true,force:true});}
});
