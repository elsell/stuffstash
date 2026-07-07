import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { ImportMessage } from '$lib/domain/inventory';
import ImportMessagesList from './ImportMessagesList.svelte';

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
