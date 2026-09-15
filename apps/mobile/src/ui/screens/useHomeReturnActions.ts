import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import type { HomeCheckedOutAssetViewModel } from '../../application/home/HomeDashboardQuery';
import type { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import { useAppFeedback } from '../feedback/AppFeedback';

export type PendingHomeReturn = {
  readonly asset: AssetCardViewModel;
  readonly checkoutId: string;
  readonly undoableOperationId: string | undefined;
  readonly details: string;
  readonly isSaving: boolean;
};

/** Owned by a dashboard keyed to its tenant/inventory; presentation also owns focus. */
export function useHomeReturnActions(command: AssetCheckoutCommand, reconcile: (shouldNotify: () => boolean) => void | Promise<void>, checkedOutAssets: readonly HomeCheckedOutAssetViewModel[], canReturn: boolean) {
  const feedback = useAppFeedback();
  const permission = useRef(canReturn);
  useLayoutEffect(() => { permission.current = canReturn; }, [canReturn]);
  const mounted = useRef(true);
  const operationPending = useRef(false);
  const completedReturns = useRef(new Map<string, string>());
  const editor = useRef<PendingHomeReturn | undefined>(undefined);
  const focus = useRef<{ active: boolean } | undefined>(undefined);
  const [returningAssetId, setReturningAssetId] = useState<string | undefined>();
  const [pendingReturn, setPendingReturn] = useState<PendingHomeReturn | undefined>();
  useEffect(() => {
    const visible = new Set(checkedOutAssets.map(asset => asset.id));
    for (const id of completedReturns.current.keys()) if (!visible.has(id)) completedReturns.current.delete(id);
  }, [checkedOutAssets]);
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  useFocusEffect(useCallback(() => {
    const session = { active: true }; focus.current = session;
    return () => { session.active = false; };
  }, []));
  function updateEditor(value: PendingHomeReturn | undefined) {
    editor.current = value;
    if (mounted.current) setPendingReturn(value);
  }
  function notice(session: { active: boolean } | undefined, title: string, error: unknown, fallback: string) {
    if (mounted.current && session?.active) feedback.showNotice({ tone: 'error', title, message: error instanceof Error ? error.message : fallback });
  }
  function alreadyReturned(asset: HomeCheckedOutAssetViewModel) {
    return completedReturns.current.has(asset.id) && (!asset.checkoutId || completedReturns.current.get(asset.id) === asset.checkoutId);
  }
  async function returnAsset(asset: HomeCheckedOutAssetViewModel) {
    const session = focus.current;
    if (!permission.current || !mounted.current || !session?.active || operationPending.current || editor.current || alreadyReturned(asset)) return;
    operationPending.current = true; setReturningAssetId(asset.id);
    try {
      const result = await command.execute({ action: 'return', assetId: asset.id });
      completedReturns.current.set(asset.id, result.id);
      if (mounted.current && session.active) {
        updateEditor({ asset, checkoutId: result.id, undoableOperationId: result.undoableOperationId, details: '', isSaving: false });
        if (!result.undoableOperationId) feedback.showNotice({ tone: 'warning', title: 'Return completed without undo', message: 'The asset was returned, but this return cannot be canceled.' });
      }
      void reconcile(() => mounted.current && session.active);
    } catch (error) {
      notice(session, 'Could not return asset', error, 'The asset was not returned.');
    } finally {
      operationPending.current = false;
      if (mounted.current) setReturningAssetId(undefined);
    }
  }
  async function finishReturn(undo: boolean) {
    const session = focus.current;
    const draft = editor.current;
    if (!mounted.current || !session?.active || operationPending.current || !draft) return;
    if (undo && !draft.undoableOperationId) { updateEditor(undefined); return; }
    operationPending.current = true; updateEditor({ ...draft, isSaving: true });
    try {
      if (undo) {
        await command.undoOperation({ operationId: draft.undoableOperationId! });
        completedReturns.current.delete(draft.asset.id);
      } else {
        await command.updateReturnedCheckoutDetails({ assetId: draft.asset.id, checkoutId: draft.checkoutId, details: draft.details });
      }
      updateEditor(undefined);
      void reconcile(() => mounted.current && session.active);
    } catch (error) {
      updateEditor({ ...draft, isSaving: false });
      notice(session, undo ? 'Could not cancel return' : 'Could not save return details', error, undo ? 'The asset is still returned.' : 'Return details were not saved.');
    } finally { operationPending.current = false; }
  }
  return {
    returningAssetId, pendingReturn, returnAsset,
    isReturnDisabled: (asset: HomeCheckedOutAssetViewModel) => operationPending.current || editor.current !== undefined || alreadyReturned(asset),
    saveReturnDetails: () => finishReturn(false),
    cancelReturn: () => finishReturn(true),
    changeDetails: (details: string) => { if (!operationPending.current && editor.current) updateEditor({ ...editor.current, details }); }
  };
}
