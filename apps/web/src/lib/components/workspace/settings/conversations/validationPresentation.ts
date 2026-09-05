import type { CaseDraftIssue } from '$lib/application/caseDraftValidation';
export function validationMessages(issues: CaseDraftIssue[]): Record<string, string> {
  return issues.reduce<Record<string, string>>((messages, issue) => {
    messages[issue.field] = [messages[issue.field], issue.message].filter(Boolean).join(' '); return messages;
  }, {});
}
export function validationAttributes(message: string | undefined, field: string) {
  return { 'aria-invalid': !!message, 'aria-describedby': message ? `${field}-error` : undefined };
}
