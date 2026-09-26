import { useLayoutEffect, useMemo } from 'react';
import type { HeaderOptions } from '../components/NativeHeaderActions.types';
import type { AssetHeaderOverflowProps } from './AssetHeaderOverflow.types';
import { assetHeaderOverflowScreenOptions } from './AssetHeaderOverflow';

/** Native menu presentation is independent of the current command closures. */
export function useAssetHeaderOverflowOptions(props: AssetHeaderOverflowProps | undefined, resourceKey: string): HeaderOptions {
  const current = useMemo<{ current: AssetHeaderOverflowProps | undefined }>(() => ({ current: undefined }), [resourceKey]);
  useLayoutEffect(() => {
    current.current = props;
    return () => { current.current = undefined; };
  }, [props, current]);
  const title = props?.asset.title;
  const canArchive = props?.asset.canArchive;
  const canRestore = props?.asset.canRestore;
  const canDeletePermanently = props?.asset.canDeletePermanently;
  const hasEdit = Boolean(props?.onEdit);
  const hasMove = Boolean(props?.onMove);
  const hasPhotos = Boolean(props?.onAddPhotos);
  const hasCheckout = Boolean(props?.onCheckout);
  const photosDisabled = props?.photosDisabled ?? false;
  const disabled = props?.disabled ?? false;
  return useMemo(() => {
    if (title === undefined) return { headerRight: undefined, unstable_headerRightItems: undefined };
    const owner = () => current.current && !current.current.disabled ? current.current : undefined;
    return assetHeaderOverflowScreenOptions({
      asset: { title, canArchive: !!canArchive, canRestore: !!canRestore, canDeletePermanently: !!canDeletePermanently },
      disabled, photosDisabled,
      onMove: hasMove ? () => owner()?.onMove?.() : undefined,
      onAddPhotos: hasPhotos ? () => { const value = owner(); if (value && !value.photosDisabled) value.onAddPhotos?.(); } : undefined,
      onCheckout: hasCheckout ? () => owner()?.onCheckout?.() : undefined,
      onEdit: hasEdit ? () => owner()?.onEdit?.() : undefined,
      onHistory: () => owner()?.onHistory(),
      onCheckoutHistory: () => owner()?.onCheckoutHistory(),
      onLifecycleAction: action => {
        const value = owner();
        if (!value) return;
        const permitted = action === 'archive' ? value.asset.canArchive
          : action === 'restore' ? value.asset.canRestore : value.asset.canDeletePermanently;
        if (permitted) value.onLifecycleAction(action);
      }
    });
  }, [title, canArchive, canRestore, canDeletePermanently, disabled, hasEdit, hasMove, hasPhotos, hasCheckout, photosDisabled, current]);
}
