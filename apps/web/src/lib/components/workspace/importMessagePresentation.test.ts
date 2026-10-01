import { expect, it } from 'vitest';
import type { ImportMessage } from '$lib/domain/inventory';
import { importIssuePresentation } from './importMessagePresentation';

function issue(code: string, detail = '', summary = 'Source issue'): ImportMessage {
  return { code, detail, summary, severity: 'warning' };
}

it('classifies source diagnostics before rendering labels and preserves unknown diagnostics', () => {
  expect(importIssuePresentation(issue('duplicate-asset')).guidance).toBe('duplicate');
  expect(importIssuePresentation(issue('legacy', 'homebox-source-id already exists')).guidance).toBe('duplicate');
  expect(importIssuePresentation(issue('attachment-unavailable')).guidance).toBe('download');
  expect(importIssuePresentation(issue('attachment-session-unavailable')).guidance).toBe('download');
  expect(importIssuePresentation(issue('legacy', 'import validation failed')).guidance).toBe('validation');
  expect(importIssuePresentation(issue('partial-date')).guidance).toBe('partialDate');
  expect(importIssuePresentation(issue('legacy', '未知の診断')).cause).toBe('未知の診断');
});

it('groups recognized causes by stable identity without conflating unknown diagnostics', () => {
  expect(importIssuePresentation(issue('duplicate-asset', 'record A')).identity)
    .toBe(importIssuePresentation(issue('duplicate-asset', 'record B')).identity);
  expect(importIssuePresentation(issue('other', 'record A')).identity)
    .not.toBe(importIssuePresentation(issue('other', 'record B')).identity);
});

it('does not let conflicting diagnostic words override recognized cause semantics', () => {
  expect(importIssuePresentation(issue('attachment-unavailable', 'URL already expired')).guidance).toBe('download');
  expect(importIssuePresentation(issue('legacy', 'import validation failed after download')).guidance).toBe('validation');
});
