export type VoiceMarkdownSpan = { readonly text: string; readonly bold?: boolean; readonly italic?: boolean; readonly code?: boolean };
export type VoiceMarkdownBlock = { readonly prefix: string; readonly heading?: boolean; readonly spans: readonly VoiceMarkdownSpan[] };

// Small presentation-only subset. Navigation is resolved separately from trusted artifacts.
export function voiceResponseMarkdown(text: string): readonly VoiceMarkdownBlock[] {
  return text.replace(/\r\n?/g, '\n').split(/\n+/).filter(line => line.trim()).map(line => {
    const heading = /^#{1,6}\s+/.test(line);
    const list = /^\s*(?:([-*+])|(\d+[.)]))\s+/.exec(line);
    const prefix = list ? (list[1] ? '• ' : `${list[2]} `) : '';
    const content = line.replace(/^#{1,6}\s+/, '').slice(list?.[0].length ?? 0);
    return { prefix, ...(heading ? { heading: true } : {}), spans: inlineSpans(content) };
  });
}

function inlineSpans(text: string): VoiceMarkdownSpan[] {
  // Keep link labels as text; never trust a model-generated URL as an asset route.
  const source = text.replace(/!?\[([^\]]+)\]\([^)]*\)/g, '$1');
  const pattern = /\*\*([^*]+)\*\*|__([^_]+)__|\*([^*\n]+)\*|`([^`\n]+)`/g;
  const spans: VoiceMarkdownSpan[] = [];
  let cursor = 0;
  for (const match of source.matchAll(pattern)) {
    const start = match.index!;
    if (start > cursor) spans.push({ text: source.slice(cursor, start) });
    if (match[1] || match[2]) spans.push({ text: match[1] ?? match[2], bold: true });
    else if (match[3]) spans.push({ text: match[3], italic: true });
    else spans.push({ text: match[4], code: true });
    cursor = start + match[0].length;
  }
  if (cursor < source.length || !spans.length) spans.push({ text: source.slice(cursor) });
  return spans;
}
