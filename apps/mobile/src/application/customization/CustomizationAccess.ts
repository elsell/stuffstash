import type { CustomizationKind, CustomizationScope } from '../../domain/customization/Customization';
import type { CustomizationContext } from './CustomizationRepository';
import type { CustomizationObservability } from './CustomizationObservability';

export class CustomizationAccessPolicy {
  constructor(private readonly observability: CustomizationObservability) {}

  canRead(context: CustomizationContext, scope: CustomizationScope): boolean {
    return scope === 'tenant'
      ? context.tenantPermissions.includes('configure')
      : context.inventoryPermissions.includes('view');
  }

  canMutate(context: CustomizationContext, kind: CustomizationKind, scope: CustomizationScope, inherited = false): boolean {
    if (inherited) return false;
    if (kind === 'tag') return context.inventoryPermissions.includes('edit_asset');
    return scope === 'tenant'
      ? context.tenantPermissions.includes('configure')
      : context.inventoryPermissions.includes('configure');
  }

  readOrRecord(context: CustomizationContext, kind: CustomizationKind, scope: CustomizationScope): boolean {
    const allowed = this.canRead(context, scope);
    if (!allowed) this.observability.record({ name: 'customization.permission_denied', resource: kind, scope });
    return allowed;
  }

  mutationOrRecord(context: CustomizationContext, kind: CustomizationKind, scope: CustomizationScope, inherited = false): boolean {
    const allowed = this.canMutate(context, kind, scope, inherited);
    if (!allowed) this.observability.record({ name: 'customization.permission_denied', resource: kind, scope });
    return allowed;
  }
}
