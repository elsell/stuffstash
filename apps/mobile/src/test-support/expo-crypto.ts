let sequence = 0;
export function randomUUID() { return `00000000-0000-4000-8000-${String(++sequence).padStart(12, '0')}`; }
