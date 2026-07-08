export function redactUnsafeVoiceText(value: string): string {
  return redactUnsafeVoiceStructuredText(value)
    .replace(/\b(raw prompt|stack trace|raw query|raw transcript|raw provider response|raw model response|provider session id)\b/gi, '[redacted]');
}

export function redactUnsafeVoiceStructuredText(value: string): string {
  return value
    .replace(/\bhttps?:\/\/[^\s"',\]}]+/gi, '[redacted-url]')
    .replace(/\b(?:ph|file|content):\/\/[^\s"',\]}]+/gi, '[redacted-uri]')
    .replace(/\b(raw prompt|stack trace|raw query|raw transcript|raw provider response|raw model response|provider session id)\b\s*[:=]\s*[^;\n\r]+/gi, '[redacted]')
    .replace(/["']?\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|secret|token|asset[-_ ]?id|parent[-_ ]?asset[-_ ]?id|inventory[-_ ]?id|tenant[-_ ]?id|tool[-_ ]?call[-_ ]?id)\b["']?\s*[:=]\s*["']?bearer\s+[^"',\s}\n]+["']?/gi, '$1: [redacted]')
    .replace(/["']?\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|secret|token|asset[-_ ]?id|parent[-_ ]?asset[-_ ]?id|inventory[-_ ]?id|tenant[-_ ]?id|tool[-_ ]?call[-_ ]?id)\b["']?\s*[:=]\s*["']?[^"',\s}\n]+["']?/gi, '$1: [redacted]')
    .replace(/bearer\s+[^"',\s}\]\)]+/gi, 'bearer [redacted]');
}
