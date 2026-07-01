export function isCurrentAuditHistoryRequest(
  currentRequestId: number,
  expectedRequestId: number
): boolean {
  return currentRequestId === expectedRequestId;
}
