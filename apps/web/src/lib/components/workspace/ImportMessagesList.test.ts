import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { ImportMessage } from '$lib/domain/inventory';
import ImportMessagesList from './ImportMessagesList.svelte';
import importMessagesListSource from './ImportMessagesList.svelte?raw';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('ImportMessagesList', () => {
  it('groups repeated import messages by severity and cause', () => {
    component = mount(ImportMessagesList, {
      target: document.body,
      props: {
        messages: [
          message('warning', 'Attachment could not be imported', 'unsupported file type', 'drill.png'),
          message('warning', 'Attachment could not be imported', 'unsupported file type', 'manual.tiff'),
          message('warning', 'Attachment could not be imported', 'download failed', 'receipt.png'),
          message('warning', 'Date imported as text', 'Homebox date has no year', 'Warranty expires'),
          message('error', 'Duplicate asset found', 'Cordless Drill already exists', 'Cordless Drill')
        ],
        emptyText: 'No blocking issues found.'
      }
    });

    const groups = Array.from(document.body.querySelectorAll<HTMLElement>('.message-group'));
    expect(document.body.querySelector('.issue-stat.warning')?.textContent).toContain('Warnings 4');
    expect(document.body.querySelector('.issue-stat.blocking')?.textContent).toContain('Blocking 1');
    expect(groups).toHaveLength(4);
    expect(groups[0]?.textContent).toContain('Warning');
    expect(groups[0]?.textContent).toContain('Attachment could not be imported');
    expect(groups[0]?.querySelector('.message-group-heading')?.textContent).toContain('unsupported file type');
    expect(groups[0]?.textContent).toContain('2 items');
    expect(groups[0]?.textContent).toContain('drill.png');
    expect(groups[0]?.textContent).toContain('manual.tiff');
    expect(groups[1]?.querySelector('.message-group-heading')?.textContent).toContain('download failed');
    expect(groups[1]?.textContent).toContain('receipt.png');
    expect(groups[2]?.textContent).toContain('Date imported as text');
    expect(groups[3]?.textContent).toContain('Blocking');
    expect(groups[3]?.textContent).toContain('Duplicate asset found');
  });

  it('keeps source IDs secondary when source names are unavailable', () => {
    component = mount(ImportMessagesList, {
      target: document.body,
      props: {
        messages: [
          sourceIDMessage('warning', 'Asset appears to have already been imported', 'homebox-source-id duplicate', 'source-wardrobe'),
          sourceIDMessage('warning', 'Asset appears to have already been imported', 'homebox-source-id duplicate', 'source-baby-hats')
        ],
        emptyText: 'No blocking issues found.'
      }
    });

    const group = document.body.querySelector<HTMLElement>('.message-group');
    expect(group?.querySelector('.message-group-heading')?.textContent).toContain('Already linked to an earlier import');
    expect(group?.textContent).toContain('2 items');
    expect(group?.textContent).toContain('Homebox record');
    expect(group?.textContent).toContain('Source ID source-wardrobe');
    expect(group?.textContent).toContain('Source ID source-baby-hats');
    expect(group?.querySelectorAll('.message-row')[0]?.textContent).not.toBe('homebox-source-id duplicate');
  });

  it('bounds large warning sets behind grouped progressive disclosure', async () => {
    component = mount(ImportMessagesList, {
      target: document.body,
      props: {
        messages: [
          ...Array.from({ length: 5 }, (_, index) =>
            message('warning', 'Attachment could not be imported', 'unsupported file type', `attachment-${index + 1}.tiff`)
          ),
          ...Array.from({ length: 19 }, (_, index) =>
            message('warning', `Import warning group ${index + 2}`, `group-${index + 2}`, `record-${index + 2}`)
          )
        ],
        emptyText: 'No blocking issues found.'
      }
    });

    const stats = Array.from(document.body.querySelectorAll<HTMLElement>('.issue-stat'));
    expect(stats.map((stat) => stat.textContent)).toEqual(expect.arrayContaining(['Groups 20', 'Affected 24', 'Warnings 24']));
    expect(document.body.querySelectorAll('.message-group')).toHaveLength(5);
    const boundedRegion = document.body.querySelector<HTMLElement>('.bounded-message-groups');
    expect(boundedRegion?.getAttribute('role')).toBe('region');
    expect(boundedRegion?.getAttribute('aria-label')).toBe('Grouped import issues');
    expect(boundedRegion?.getAttribute('tabindex')).toBe('0');
    expect(document.body.textContent).toContain('2 more in this group');
    expect(document.body.textContent).toContain('15 more issue groups hidden.');
    expect(document.body.textContent).toContain('attachment-3.tiff');
    expect(document.body.textContent).not.toContain('attachment-4.tiff');
    expect(document.body.textContent).not.toContain('Import warning group 7');

    document.body.querySelector<HTMLButtonElement>('button')?.click();
    await tick();

    expect(document.body.querySelectorAll('.message-group')).toHaveLength(15);
    expect(document.body.textContent).toContain('5 more issue groups hidden.');
    expect(document.body.textContent).toContain('attachment-4.tiff');
    expect(document.body.textContent).toContain('attachment-5.tiff');
    expect(document.body.textContent).toContain('Import warning group 15');
    expect(document.body.textContent).not.toContain('Import warning group 16');
  });

  it('keeps warning stats on warning tokens instead of destructive alarm tokens', () => {
    const warningRule = importMessagesListSource.match(/\.issue-stat\.warning\s*{(?<body>[^}]*)}/)?.groups?.body ?? '';
    const warningStrongRule = importMessagesListSource.match(/\.issue-stat\.warning strong\s*{(?<body>[^}]*)}/)?.groups?.body ?? '';

    expect(warningRule).toContain('var(--color-warning)');
    expect(warningRule).not.toContain('destructive');
    expect(warningStrongRule).toContain('var(--color-warning-foreground)');
    expect(warningStrongRule).not.toContain('destructive');
  });
});

function message(severity: ImportMessage['severity'], summary: string, detail: string, sourceName: string): ImportMessage {
  return {
    code: summary.toLowerCase().replaceAll(' ', '-'),
    severity,
    summary,
    detail,
    sourceName
  };
}

function sourceIDMessage(severity: ImportMessage['severity'], summary: string, detail: string, sourceId: string): ImportMessage {
  return {
    code: summary.toLowerCase().replaceAll(' ', '-'),
    severity,
    summary,
    detail,
    sourceId
  };
}
