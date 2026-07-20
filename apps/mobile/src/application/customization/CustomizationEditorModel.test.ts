import { describe, expect, it } from 'vitest';
import {
  customizationEditorIsValid,
  customizationEditorIsDirty,
  customizationEditorSnapshot,
  effectiveInheritedOwnership,
  emptyCustomizationEditorDraft,
  withEditorName,
  withManualEditorKey
} from './CustomizationEditorModel';

describe('CustomizationEditorModel', () => {
  it('tracks the complete multiword name until stable key is explicitly edited', () => {
    let draft = emptyCustomizationEditorDraft();
    draft = withEditorName(draft, 'W');
    expect(draft.key).toBe('w');
    draft = withEditorName(draft, 'Winter clothing');
    expect(draft.key).toBe('winter-clothing');

    draft = withManualEditorKey(draft, 'seasonal-gear');
    draft = withEditorName(draft, 'Winter clothing and shoes');
    expect(draft.key).toBe('seasonal-gear');
  });

  it('keeps validation and authoritative inherited ownership out of UI components', () => {
    const validTag = { ...emptyCustomizationEditorDraft(), name: 'Tools', key: 'tools', color: '#2F80ED' };
    expect(customizationEditorIsValid(validTag, 'tag', 'create')).toBe(true);
    expect(effectiveInheritedOwnership({ routeHint: false, recordScope: 'tenant', screenScope: 'inventory' })).toBe(true);
    expect(effectiveInheritedOwnership({ routeHint: true, recordScope: 'inventory', screenScope: 'inventory' })).toBe(false);
  });

  it('owns dirty-state comparison and suppresses it after completion', () => {
    const initial = { ...emptyCustomizationEditorDraft(), name: 'Winter gear', key: 'winter-gear' };
    const snapshot = customizationEditorSnapshot(initial);
    expect(customizationEditorIsDirty(initial, snapshot, 'edit', false)).toBe(false);
    expect(customizationEditorIsDirty({ ...initial, name: 'Cold weather gear' }, snapshot, 'edit', false)).toBe(true);
    expect(customizationEditorIsDirty({ ...initial, name: 'Cold weather gear' }, snapshot, 'edit', true)).toBe(false);
    expect(customizationEditorIsDirty({ ...emptyCustomizationEditorDraft(), key: 'manual-key', keyManuallyEdited: true }, '', 'create', false)).toBe(true);
    expect(customizationEditorIsDirty({ ...emptyCustomizationEditorDraft(), fieldType: 'number' }, '', 'create', false)).toBe(true);
  });
});
