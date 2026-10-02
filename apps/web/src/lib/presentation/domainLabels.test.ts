import { expect, it } from 'vitest';
import { t } from './localization';
import { assetLifecycleLabel } from './assetKindLabel';
import { customizationScopeLabel, customizationFieldTypeOptions } from '../application/workspaceCustomizationPresentation';
import { caseOperationLabel, caseOutcomeLabel } from './conversationCaseLabels';

it('uses cataloged asset and customization labels without changing option values', () => {
  expect(assetLifecycleLabel('active')).toBe(t('asset.lifecycle.active'));
  expect(assetLifecycleLabel('archived')).toBe(t('asset.lifecycle.archived'));
  expect(customizationScopeLabel('tenant')).toBe(t('web.workspaceCustomizationPresentation.tenant'));
  expect(customizationFieldTypeOptions().find(option => option.value === 'boolean')).toEqual({ value: 'boolean', label: t('web.FieldSettingsManager.yesNo') });
});
it('presents conversation outcomes and read/write operations through catalog labels', () => {
  expect(caseOperationLabel('checkout')).toBe(t('case.operation.checkout'));
  expect(caseOperationLabel('list_contents')).toBe(t('case.operation.list_contents'));
  expect(caseOutcomeLabel('clarification')).toBe(t('case.outcome.clarification'));
});
