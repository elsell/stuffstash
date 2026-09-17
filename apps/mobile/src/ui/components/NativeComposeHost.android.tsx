import { Host, type HostProps } from '@expo/ui/jetpack-compose';
import { useAppearance } from '../theme/AppearanceContext';

/** Keep independent Compose roots aligned with the app's appearance override. */
export function NativeComposeHost(props: Omit<HostProps, 'colorScheme'>) {
  const { resolvedColorScheme } = useAppearance();
  return <Host {...props} colorScheme={resolvedColorScheme} />;
}
