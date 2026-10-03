import type { RotationTarget } from '$lib/ports/printRotationRepository';
const fields = ['tenantId', 'inventoryId', 'connectorId'] as const;
// Undefined means ordinary registration; null means a malformed rotation intent.
export function rotationIntent(query: URLSearchParams): RotationTarget | null | undefined {
    if (!fields.some(key => query.has(key)))
        return undefined;
    if (fields.some(key => query.getAll(key).length !== 1))
        return null;
    const values = fields.map(key => query.get(key)!);
    if (values.some(value => !value || value.length > 200 || value.trim() !== value || /[\u0000-\u001f\u007f]/.test(value)))
        return null;
    return { tenantId: values[0], inventoryId: values[1], connectorId: values[2] };
}
export function pairingReturnPath(id: string, target: RotationTarget | undefined): string {
    const path = `/print-connectors/pair/${encodeURIComponent(id)}`;
    return target ? `${path}?${new URLSearchParams(target)}` : path;
}
