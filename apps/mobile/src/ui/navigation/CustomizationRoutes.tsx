import { useLocalSearchParams, useRouter } from 'expo-router';
import type { CustomizationKind, CustomizationScope } from '../../domain/customization/Customization';
import { CustomizationCollectionScreen } from '../screens/CustomizationCollectionScreen';
import { CustomizationEditorScreen } from '../screens/CustomizationEditorScreen';
import { useAppServices } from './AppServicesContext';
import { customizationEditorTarget } from './CustomizationRoutesPresentation';

export function CustomizationCollectionRoute({ kind, scope }: { readonly kind: CustomizationKind; readonly scope: CustomizationScope }) {
  const router = useRouter(); const services = useAppServices(); const base = `/settings/${scope === 'tenant' ? 'household' : 'inventory'}/${segment(kind)}`;
  return <CustomizationCollectionScreen accessPolicy={services.customizationAccessPolicy} contextQuery={services.customizationContextQuery} kind={kind} query={services.customizationCollectionQuery} scope={scope} onAdd={() => router.push(`${base}/new` as never)} onOpen={(row, inherited, canManageInherited) => {
    const lifecycle = 'lifecycle' in row ? row.lifecycle : 'active';
    router.push(customizationEditorTarget(kind, scope, row.id, lifecycle, inherited, canManageInherited) as never);
  }} />;
}

export function CustomizationEditorRoute({ kind, mode, scope }: { readonly kind: CustomizationKind; readonly mode: 'create' | 'edit'; readonly scope: CustomizationScope }) {
  const router = useRouter(); const params = useLocalSearchParams<{ resourceId?: string; lifecycle?: 'active' | 'archived'; inherited?: string }>(); const services = useAppServices();
  const collection = `/settings/${scope === 'tenant' ? 'household' : 'inventory'}/${segment(kind)}`;
  return <CustomizationEditorScreen accessPolicy={services.customizationAccessPolicy} contextQuery={services.customizationContextQuery} inherited={params.inherited === 'true'} kind={kind} lifecycle={params.lifecycle ?? 'active'} manageAssetTypes={services.manageCustomAssetTypes} manageFields={services.manageCustomFields} manageTags={services.manageTags} mode={mode} onDone={() => router.replace(collection as never)} onManageInherited={params.resourceId ? () => router.replace(customizationEditorTarget(kind, 'tenant', params.resourceId!, params.lifecycle ?? 'active', false, false) as never) : undefined} query={services.customizationCollectionQuery} resourceId={params.resourceId} scope={scope} />;
}

function segment(kind: CustomizationKind) { return kind === 'asset-type' ? 'asset-types' : kind === 'field' ? 'fields' : 'tags'; }
