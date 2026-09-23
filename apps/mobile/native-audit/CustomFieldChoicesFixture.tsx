import { useState } from 'react';
import { Button, ScrollView, Text } from 'react-native';
import type { CustomAssetTypeDefinition, CustomFieldApplicability, CustomFieldType } from '../src/domain/customization/Customization';
import { CustomizationFieldControls } from '../src/ui/components/CustomizationEditorFields';

const eligibleTypes: readonly CustomAssetTypeDefinition[] = Array.from({ length: 12 }, (_, index) => ({
  kind: 'asset-type', id: `type-${index + 1}`, tenantId: 'audit-household', inventoryId: 'audit-inventory',
  scope: 'inventory', key: `type-${index + 1}`, displayName: `Audit type ${String(index + 1).padStart(2, '0')}`,
  description: '', lifecycle: 'active'
}));

export function CustomFieldChoicesFixture({ onBack }: { readonly onBack: () => void }) {
  const [fieldType, setFieldType] = useState<CustomFieldType>('text');
  const [applicability, setApplicability] = useState<CustomFieldApplicability>('all_assets');
  const [targetIds, setTargetIds] = useState<readonly string[]>([]);
  const [options, setOptions] = useState<readonly string[]>(['ready']);
  const [newOption, setNewOption] = useState('');
  return <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={{ padding: 20, gap: 20 }}>
    <Button title="Back to audit menu" onPress={onBack} />
    <Text>{`Field type: ${fieldType}`}</Text>
    <Text>{`Applicability: ${applicability}`}</Text>
    <Text>{`Selected targets: ${targetIds.join(', ') || 'none'}`}</Text>
    <CustomizationFieldControls mode="create" canMutate fieldType={fieldType} onFieldType={setFieldType}
      applicability={applicability} onApplicability={setApplicability} eligibleTypes={eligibleTypes}
      targetIds={targetIds} persistedTargetIds={[]} onTargets={setTargetIds}
      enumOptions={options} persistedEnumOptions={[]} onEnumOptions={setOptions}
      newOption={newOption} onNewOption={setNewOption} />
  </ScrollView>;
}
