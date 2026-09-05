import type { CustomizationKind } from '../../domain/customization/Customization';
export type CustomizationMutation = { readonly tenantId: string; readonly inventoryId?: string; readonly kind: CustomizationKind };
export interface CustomizationMutationObserver { onCustomizationChanged(mutation: CustomizationMutation): void; }
