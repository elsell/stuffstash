import { describe, expect, it } from 'vitest';
import type { ImportJob } from '$lib/domain/inventory';
import { sourceDescription, sourceOptionsSummary, statusSentence, visiblePreviewMessages } from './importWorkspacePresentation';

describe('importWorkspacePresentation', () => {
  it('deduplicates fallback job messages before limiting preview-visible messages', () => {
    const job = importJobWithMessages([
      ...Array.from({ length: 8 }, () => message('duplicate-source')),
      message('distinct-source')
    ]);

    expect(visiblePreviewMessages(job).map((item) => item.sourceId)).toEqual(['duplicate-source', 'distinct-source']);
  });

  it('keeps default live Homebox source copy focused on source identity instead of noisy options', () => {
    const job = importJobWithMessages([]);
    job.source.baseUrl = 'https://stuff.jsksell.com/api/v1';
    job.source.version = 'v0.24.2';
    job.source.imageImport = 'enabled';
    job.source.allowPrivateNetwork = false;
    job.source.allowInsecureTLS = false;

    expect(sourceDescription(job)).toBe('stuff.jsksell.com · v0.24.2');
    expect(sourceOptionsSummary(job)).toEqual(['Connected directly to Homebox']);
  });

  it('uses the durable diagnostic code rather than count heuristics for image-session failure copy', () => {
    const earlierFailure = importJobWithMessages([]);
    earlierFailure.status = 'failed';
    earlierFailure.counts.attachments = 139;
    earlierFailure.counts.assetsCreated = 2;
    earlierFailure.progress = { phase: 'terminal', done: 0, total: 139 };
    expect(statusSentence(earlierFailure)).toBe('Import stopped. 2 records were kept.');

    earlierFailure.messages = [{
      code: 'attachment-session-unavailable', severity: 'error', summary: 'Image import could not start', detail: 'Source session failed.'
    }];
    expect(statusSentence(earlierFailure)).toBe('Image import did not start. 2 earlier records were kept.');
  });
});

function importJobWithMessages(messages: ImportJob['messages']): ImportJob {
  return {
    id: 'job-preview-messages',
    status: 'succeeded',
    source: {
      type: 'legacy_homebox',
      name: 'Homebox',
      imageImport: 'enabled'
    },
    counts: {
      fields: 0,
      locations: 0,
      assets: 0,
      attachments: 0,
      warnings: messages.length,
      errors: 0,
      fieldsCreated: 0,
      fieldsExisting: 0,
      locationsCreated: 0,
      assetsCreated: 0,
      assetsSkipped: 0,
      attachmentsCreated: 0,
      attachmentsSkipped: 0,
      recordsDiscarded: 0,
      sourceLinksDiscarded: 0
    },
    preview: {
      fields: [],
      locations: [],
      assets: [],
      attachments: [],
      messages: [],
      fieldsTruncated: false,
      locationsTruncated: false,
      assetsTruncated: false,
      attachmentsTruncated: false,
      messagesTruncated: false
    },
    progress: { phase: 'terminal', done: 1, total: 1 },
    progressHistory: [],
    createdAt: '2026-07-06T12:00:00Z',
    updatedAt: '2026-07-06T12:00:00Z',
    resources: [],
    messages
  };
}

function message(sourceId: string): ImportJob['messages'][number] {
  return {
    code: 'partial-date',
    severity: 'warning',
    summary: 'Homebox partial date imported as text',
    detail: '0001-09-28',
    sourceId
  };
}
