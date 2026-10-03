/** Non-secret identity only. The scanned origin must never become an API base. */
export interface LabelReference { instanceId: string; labelId: string }
export class LabelLinkError extends Error {
  constructor(readonly code: 'invalid_label' | 'unsupported_version') { super(code); this.name = 'LabelLinkError'; }
}
const opaqueID = /^[0-7][0-9A-HJKMNP-TV-Z]{25}$/;
export function parseLabelLink(value: string): LabelReference {
  const invalid = () => { throw new LabelLinkError('invalid_label'); };
  if (value.length > 4096 || value !== value.trim() || /[\\\s%]/.test(value)) return invalid();
  let url: URL;
  try { url = new URL(value); } catch { return invalid(); }
  if (url.username || url.password || url.search || url.hash || value.includes('?') || value.includes('#')) return invalid();
  // Reject before URL normalization can erase dot segments.
  if (/(?:^|\/)\.{1,2}(?:\/|$)/.test(value)) return invalid();
  let match: RegExpMatchArray | null;
  if (url.protocol === 'https:') {
    if (!url.hostname || url.pathname.includes('//')) return invalid();
    match = url.pathname.match(/^(?:\/[A-Za-z0-9._~-]+)*\/l\/(v[0-9]+)\/([^/]+)\/([^/]+)$/);
  } else if (url.protocol === 'stuffstash:' && url.host === 'labels') {
    match = url.pathname.match(/^\/(v[0-9]+)\/([^/]+)\/([^/]+)$/);
  } else return invalid();
  if (!match) return invalid();
  if (match[1] !== 'v1') throw new LabelLinkError('unsupported_version');
  if (!opaqueID.test(match[2]) || !opaqueID.test(match[3])) return invalid();
  return {instanceId:match[2], labelId:match[3]};
}
