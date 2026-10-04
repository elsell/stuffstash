import { useEffect } from 'react';
import { usePathname, useRouter } from 'expo-router';
import { usePendingLabel } from './LabelLinkContext';
import { useAppFeedback } from '../feedback/AppFeedback';
import { t } from '../../presentation/localization';
/** Warm links offer navigation rather than replacing an unrelated draft. */
export function PendingLabelNavigation() {
  const pending = usePendingLabel(); const feedback = useAppFeedback(); const router = useRouter(); const path = usePathname();
  useEffect(() => {
    if ((!pending.reference && !pending.invalid) || path === '/scan-label') return;
    feedback.showNotice({ tone: 'info', title: t('labels.mobile.scan'), message: t('labels.mobile.ready'),
      action: { label: t('labels.mobile.open'), onPress: () => router.push('/scan-label') } });
  }, [pending.revision, path, feedback, router]);
  return null;
}
