import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vitest/config';

const support = (file: string) => fileURLToPath(new URL(`./src/test-support/${file}`, import.meta.url));

export default defineConfig({
  test: { setupFiles: ['./native-runtime.setup.ts'], server: { deps: { inline: ['react-native-image-viewing'] } } },
  resolve: {
    alias: [
      { find: /^@expo\/ui\/jetpack-compose$/, replacement: support('expo-jetpack-compose.ts') },
      { find: /^@expo\/ui\/jetpack-compose\/modifiers$/, replacement: support('expo-jetpack-compose-modifiers.ts') },
      { find: /^expo$/, replacement: support('expo.ts') },
      { find: /^@react-navigation\/elements$/, replacement: support('react-navigation-elements.ts') },
      { find: /^@expo\/ui\/swift-ui$/, replacement: support('expo-swift-ui.ts') },
      { find: /^@expo\/ui\/swift-ui\/modifiers$/, replacement: support('expo-swift-ui-modifiers.ts') },
      { find: /^expo-notifications$/, replacement: support('expo-notifications.ts') },
      { find: /^expo-image-picker$/, replacement: support('expo-image-picker.ts') },
      { find: /^expo-crypto$/, replacement: support('expo-crypto.ts') },
      { find: /^@react-native-community\/datetimepicker$/, replacement: support('native-date-picker.ts') },
      { find: /^expo-network$/, replacement: support('expo-network.ts') },
      { find: /^react-native-image-viewing$/, replacement: support('react-native-image-viewing.ts') },
      { find: /^react-native$/, replacement: support('react-native.ts') },
      { find: /^react-native-safe-area-context$/, replacement: support('react-native-safe-area-context.tsx') },
      { find: /^react-native-keyboard-controller$/, replacement: support('react-native-keyboard-controller.ts') },
      { find: /^lucide-react-native$/, replacement: support('lucide-react-native.ts') },
      { find: /^expo-router$/, replacement: support('expo-router.ts') },
      { find: /^@react-navigation\/native$/, replacement: support('react-navigation-native.ts') },
      { find: /^@expo\/ui\/community\/segmented-control$/, replacement: support('expo-segmented-control.ts') },
      { find: /^react-native-svg$/, replacement: support('react-native-svg.ts') }
    ]
  }
});
